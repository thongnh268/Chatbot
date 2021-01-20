package runner

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"github.com/subiz/perm"
	ggrpc "github.com/subiz/sgrpc"
	"google.golang.org/protobuf/proto"

	"google.golang.org/grpc"
)

type ObjectMgr interface {
	TerminateBot(ctx context.Context, in *header.BotTerminated, opts ...grpc.CallOption) (*header.Event, error)
}

type ExecBotMgr interface {
	ReadExecBot(botId, id string) (*ExecBot, error)
	UpdateExecBot(execBot *ExecBot) (*ExecBot, error)
	CreateExecBotIndex(execBotIndex *ExecBotIndex) (*ExecBotIndex, error)
	LoadExecBots(accountId string, pks []*ExecBotPK) chan []*ExecBot
}

type BotSched interface {
	Push(schedAt int64, accountId string) bool
}

type ActionMgr interface {
	GetEvtCondTriggers(evt *header.Event) []string
	Do(node *ActionNode, req *header.RunRequest) ([]string, error)
	Wait(node *ActionNode) string
	Triggers(node *ActionNode) []string
}

// Default exec bot of user
// Require bot, event belongs to account
type BotRunner struct {
	*sync.Mutex
	AccountId  string
	execBotMgr ExecBotMgr
	actionMgr  ActionMgr
	sched      BotSched
	schedAt    int64
	Realtime   header.PubsubClient // TODO private
	ObjectMgr  ObjectMgr

	logicBots   map[string]*LogicBot
	triggerBots map[string][]string

	execBots        map[string]*ExecBot
	triggerExecBots map[string][]string // foreground
	execBotTriggers map[string][]string

	suspendExecBots map[string]*ExecBot
	objExecBots     map[string][]string // lockground

	runningExecBots map[string]int64   // required, any
	schedExecBots   map[int64][]string // background, any
}

func NewBotRunner(accountId string, execBotMgr ExecBotMgr, sched BotSched, actionMgr ActionMgr) *BotRunner {
	runner := &BotRunner{
		Mutex:      &sync.Mutex{},
		AccountId:  accountId,
		execBotMgr: execBotMgr,
		sched:      sched,
		actionMgr:  actionMgr,
	}

	runner.logicBots = make(map[string]*LogicBot)
	runner.triggerBots = make(map[string][]string)

	runner.runningExecBots = make(map[string]int64)

	runner.execBots = make(map[string]*ExecBot)
	runner.triggerExecBots = make(map[string][]string)
	runner.execBotTriggers = make(map[string][]string)
	runner.schedExecBots = make(map[int64][]string)

	runner.suspendExecBots = make(map[string]*ExecBot)
	runner.objExecBots = make(map[string][]string)

	return runner
}

func (runner *BotRunner) Status() *BotRunnerStatus {
	runner.Lock()
	defer runner.Unlock()
	status := NewBotRunnerStatus(runner.AccountId)
	status.LogicBotCount = len(runner.logicBots)
	status.ExecBotCount = len(runner.execBots)
	status.SuspendExecBotCount = len(runner.suspendExecBots)
	status.RunningExecBotCount = len(runner.runningExecBots)
	for _, arr := range runner.triggerBots {
		status.TriggerBotCount += len(arr)
	}
	for _, arr := range runner.triggerExecBots {
		status.TriggerExecBotCount += len(arr)
	}
	for _, arr := range runner.execBotTriggers {
		status.ExecBotTriggerCount += len(arr)
	}
	for _, arr := range runner.objExecBots {
		status.ObjExecBotCount += len(arr)
	}
	for _, arr := range runner.schedExecBots {
		status.SchedExecBotCount += len(arr)
	}
	return status
}

func (runner *BotRunner) TerminateExecBot(in *header.RunRequest) {
	runner.Lock()
	defer runner.Unlock()

	execBotId := NewExecBotID(in.GetAccountId(), in.GetBotId(), in.GetObjectId())
	var execBot *ExecBot
	execBot = runner.execBots[execBotId]
	if execBot == nil {
		execBot = runner.suspendExecBots[execBotId]
	}
	if execBot == nil {
		return
	}

	runner.rmTriggerExecBots(execBotId)
	delete(runner.execBots, execBotId)
	delete(runner.suspendExecBots, execBotId)
	if _, has := runner.runningExecBots[execBotId]; has {
		return
	}
	execBot.Terminate(header.BotTerminated_force.String())
}

// don't require evt of new exec bot
func (runner *BotRunner) StartBot(in *header.RunRequest) error {
	runner.Lock()
	defer runner.Unlock()

	botId := in.GetBotId()
	objId := in.GetObjectId()

	if objId == "" {
		return errors.New("obj id is required")
	}

	logicBot := runner.logicBots[botId]
	bot, actionMap, actionMaph := logicBot.GetBot(), logicBot.GetActionMap(), logicBot.GetActionMaph()
	// don't support action map, hash
	if in.GetMode() == Exec_bot_mode_debug {
		bot = in.GetBot()
		actionMap = nil
		actionMaph = ""
	}
	if bot == nil {
		return errors.New("bot is not exist")
	}
	if !RunCondition(bot, in) {
		return errors.New("false condition")
	}

	execBotId := NewExecBotID(bot.GetAccountId(), bot.GetId(), objId)
	if _, has := runner.runningExecBots[execBotId]; has {
		return errors.New("exec bot is running")
	}

	var execBot *ExecBot
	var err error

	execBot = runner.execBots[execBotId]
	if execBot == nil {
		execBot = runner.suspendExecBots[execBotId]
	}
	if execBot == nil {
		execBot, err = MakeExecBot(bot, actionMaph, objId, execBotId, in.GetObjectContexts())
		if err != nil {
			return err
		}
		execBot.DebuggingMode(in.GetMode(), runner.Realtime)
		runner.addFuncExecBot(execBot)
		runner.execBots[execBotId] = execBot
		runner.addTriggerExecBots(execBot) // Keep trigger
	}

	in.BotRunType = "onstart"
	runner.runningExecBots[execBot.Id] = time.Now().UnixNano() / 1e6 // TODO move to run exec bot
	// TODO set bot
	logicBotCopy := &LogicBot{AccountId: bot.GetAccountId(), Id: bot.GetId(), Bot: bot, actionMap: actionMap, actionMaph: actionMaph}
	go runExecBot(execBot, in, runner.execBotMgr, logicBotCopy)

	return nil
}

func (runner *BotRunner) OnEvent(in *header.RunRequest) {
	runner.Lock()
	defer runner.Unlock()

	objType := in.GetObjectType()
	objId := in.GetObjectId()
	evt := in.GetEvent()

	if objId == "" {
		log.Error("obj id is required")
		return
	}

	execBotTriggers := TriggerOnEvent(objType, objId, evt, runner.actionMgr)
	runnable := make(map[string]*ExecBot, 0)
	for _, trigger := range execBotTriggers {
		for _, execBotId := range runner.triggerExecBots[trigger] {
			if _, has := runnable[execBotId]; has {
				continue
			}
			execBot, has := runner.execBots[execBotId]
			if !has {
				continue
			}
			runnable[execBotId] = execBot
		}
	}

	for _, execBotId := range runner.objExecBots[objId] {
		if _, has := runnable[execBotId]; has {
			continue
		}
		if execBot, has := runner.suspendExecBots[execBotId]; has {
			runnable[execBotId] = execBot
		}
	}

	// TODO remove if not use event
	// runner.autoStartBot(in, runnable)

	in.BotRunType = "onevent"
	evtById := in.GetEvent().GetBy().GetId() // TODO type
	now := time.Now().UnixNano() / 1e6
	for _, execBot := range runnable {
		if execBot.BotId == evtById {
			continue // ignore self's event
		}
		if _, has := runner.runningExecBots[execBot.Id]; has {
			if execBot.reqQueue == nil {
				execBot.reqQueue = NewQueue()
			}
			execBot.reqQueue.Enqueue(in)
			continue
		}
		runner.runningExecBots[execBot.Id] = now

		go runExecBot(execBot, in, runner.execBotMgr, runner.logicBots[execBot.BotId])
	}

	// Keep trigger
	// for _, trigger := range execBotTriggers {
	// 	delete(runner.triggerExecBots, trigger)
	// }
}

func (runner *BotRunner) OnForgot() {
	runner.Lock()
	defer runner.Unlock()
	if runner.schedAt != 0 {
		return
	}

	now := time.Now().UnixNano() / 1e6 / 1e2
	schedAtBound := now + Runner_forgot_time
	nextSchedAt := schedAtBound
	for schedAt := range runner.schedExecBots {
		if schedAt < nextSchedAt {
			nextSchedAt = schedAt
		}
	}

	if nextSchedAt == schedAtBound {
		return
	}

	if nextSchedAt < now {
		nextSchedAt = now
	}

	isScheduled := runner.sched.Push(nextSchedAt*1e2, runner.AccountId)
	if isScheduled {
		runner.schedAt = nextSchedAt
	}
}

func (runner *BotRunner) OnTime(schedNow int64) {
	runner.Lock()
	defer runner.Unlock()

	schedAtBound := schedNow + Runner_forgot_time
	nextSchedAt := schedAtBound
	rmArr := make([]int64, 0)
	runnable := make(map[string]*ExecBot)

	for schedAt, execBotIds := range runner.schedExecBots {
		if schedNow < schedAt {
			if schedAt < nextSchedAt {
				nextSchedAt = schedAt
			}
			continue
		}
		rmArr = append(rmArr, schedAt)
		for _, execBotId := range execBotIds {
			if _, has := runner.execBots[execBotId]; has {
				runnable[execBotId] = runner.execBots[execBotId]
				continue
			}

			if _, has := runner.suspendExecBots[execBotId]; has {
				runnable[execBotId] = runner.suspendExecBots[execBotId]
			}
		}
	}

	req := &header.RunRequest{Created: schedNow * 1e2}
	req.BotRunType = "ontime"
	for _, execBot := range runnable {
		if _, has := runner.runningExecBots[execBot.Id]; has {
			continue
		}
		runner.runningExecBots[execBot.Id] = schedNow

		go runExecBot(execBot, req, runner.execBotMgr, runner.logicBots[execBot.BotId])
	}

	for _, schedAt := range rmArr {
		delete(runner.schedExecBots, schedAt)
	}

	delete(runner.schedExecBots, runner.schedAt)
	if len(runner.schedExecBots) == 0 {
		runner.schedAt = 0
	}

	if nextSchedAt == schedAtBound {
		return
	}

	isScheduled := runner.sched.Push(nextSchedAt*1e2, runner.AccountId)
	if isScheduled {
		runner.schedAt = nextSchedAt
	}
}

func (runner *BotRunner) OnLock() {
	runner.Lock()
	now := time.Now().UnixNano() / 1e6
	lgExecBots := make([]*ExecBot, 0)
	for _, execBot := range runner.execBots {
		if _, has := runner.runningExecBots[execBot.Id]; has {
			continue
		}
		if now > execBot.LastSeen+Runner_lock_time_out {
			lgExecBots = append(lgExecBots, execBot)
		}
	}
	execBotRaws := make([]*ExecBot, len(lgExecBots))
	for i := 0; i < len(lgExecBots); i++ {
		execBotRaws[i] = lgExecBots[i].ToRaw()
	}
	runner.Unlock()

	updatedExecBotMap := make(map[string]struct{})
	for _, execBotRaw := range execBotRaws {
		_, err := runner.execBotMgr.UpdateExecBot(execBotRaw)
		if err != nil {
			log.Error(execBotRaw.ObjId, "exec-bot-mgr.update", fmt.Sprintf("account-id=%s bot-id=%s", execBotRaw.AccountId, execBotRaw.BotId), fmt.Sprintf("err=%s", err))
			continue
		}
		updatedExecBotMap[execBotRaw.Id] = struct{}{}
	}

	runner.Lock()
	defer runner.Unlock()
	for _, execBot := range lgExecBots {
		if _, has := updatedExecBotMap[execBot.Id]; !has {
			continue
		}
		runner.suspendExecBots[execBot.Id] = execBot.Suspend()
		if runner.objExecBots[execBot.ObjId] == nil {
			runner.objExecBots[execBot.ObjId] = []string{}
		}
		runner.objExecBots[execBot.ObjId] = append(runner.objExecBots[execBot.ObjId], execBot.Id)

		runner.rmTriggerExecBots(execBot.Id)
		delete(runner.execBots, execBot.Id)
	}
}

func (runner *BotRunner) AfterRunExecBot(execBot *ExecBot) {
	runner.Lock()
	defer runner.Unlock()
	// terminate if not found
	if _, has := runner.execBots[execBot.Id]; !has {
		log.Info(execBot.ObjId, "exec-bot.after-run", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("not found"))
		execBot.Terminate(header.BotTerminated_force.String())
		delete(runner.runningExecBots, execBot.Id)
		return
	}
	// TODO action tree is nil
	// TODO just growed
	if runner.execBotMgr != nil {
		go runner.execBotMgr.UpdateExecBot(execBot.ToRaw())
	}
	if bot := runner.logicBots[execBot.BotId].GetBot(); bot != nil {
		bot.LastExecuted = time.Now().UnixNano() / 1e6
	}

	if execBot.reqQueue != nil {
		reqs := execBot.reqQueue.Peek(100)
		if len(reqs) > 0 {
			log.Info(execBot.ObjId, "exec-bot.run", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("reqs.len=%d", len(reqs)))
			go execBot.Run(reqs)
			return
		}
	}

	delete(runner.runningExecBots, execBot.Id)

	if execBot.Status == Exec_bot_state_terminated {
		runner.rmTriggerExecBots(execBot.Id)
		delete(runner.execBots, execBot.Id)
		delete(runner.suspendExecBots, execBot.Id)
		return
	}

	runner.execBots[execBot.Id] = execBot

	runner.rmTriggerExecBots(execBot.Id)
	runner.addTriggerExecBots(execBot)

	runner.addSchedExecBots(execBot)
}

func (runner *BotRunner) GetBotMap() map[string]*header.Bot {
	runner.Lock()
	defer runner.Unlock()
	botMap := make(map[string]*header.Bot, len(runner.logicBots))
	for botId, logicBot := range runner.logicBots {
		if logicBot.GetBot() != nil {
			botMap[botId] = logicBot.GetBot()
		}
	}
	return botMap
}

func (runner *BotRunner) GetChangedBotReports() []*ReportRun {
	runner.Lock()
	runner.Unlock()
	reportArr := make([]*ReportRun, 0)
	for _, logicBot := range runner.logicBots {
		if logicBot == nil || logicBot.Report == nil {
			continue
		}
		if logicBot.Report.ShouldUpdate() {
			reportCopy := &ReportRun{AccountId: runner.AccountId, BotId: logicBot.Id}
			reportArr = append(reportArr, reportCopy)
			reportCopy.Metrics = make([]*header.Metric, len(logicBot.Report.Metrics))
			for i, metric := range logicBot.Report.Metrics {
				reportCopy.Metrics[i] = proto.Clone(metric).(*header.Metric)
			}
		}
	}
	return reportArr
}

func (runner *BotRunner) GetExecBotMap() map[string]*ExecBot {
	runner.Lock()
	defer runner.Unlock()
	execBotMap := make(map[string]*ExecBot)
	for execBotId, execBot := range runner.suspendExecBots {
		execBotMap[execBotId] = execBot
	}
	for execBotId, execBot := range runner.execBots {
		execBotMap[execBotId] = execBot
	}
	return execBotMap
}

type ExecBotPK struct {
	AccountId string
	BotId     string
	Id        string
}

// TODO sched actions is not empty
func (runner *BotRunner) CleanExecBots() {
	runner.Lock()
	now := time.Now().UnixNano() / 1e6
	rmObjIds := make([]string, 0)
	rmExecBotIds := make([]string, 0)
	suspendPKs := make([]*ExecBotPK, 0)
	for objId, execBotIds := range runner.objExecBots {
		if len(execBotIds) == 0 {
			rmObjIds = append(rmObjIds, objId)
			continue
		}

		isRm := true
		for _, execBotId := range execBotIds {
			if _, has := runner.runningExecBots[execBotId]; has {
				continue
			}
			execBot, has := runner.suspendExecBots[execBotId]
			if !has {
				continue
			}
			if now < execBot.LastSeen+Runner_lock_time_expire {
				isRm = false
				continue
			}
			if execBot.isSuspend {
				suspendPKs = append(suspendPKs, &ExecBotPK{AccountId: runner.AccountId, BotId: execBot.BotId, Id: execBot.Id})
				isRm = false
				continue
			}
			rmExecBotIds = append(rmExecBotIds, execBot.Id)
			execBot.Terminate(header.BotTerminated_expire.String())
		}
		if isRm {
			rmObjIds = append(rmObjIds, objId)
		}
	}

	for _, objId := range rmObjIds {
		delete(runner.objExecBots, objId)
	}
	for _, execBotId := range rmExecBotIds {
		delete(runner.suspendExecBots, execBotId)
	}
	runner.Unlock()

	c := runner.execBotMgr.LoadExecBots(runner.AccountId, suspendPKs)
	if c == nil {
		return
	}
	for srcExecBots := range c {
		runner.Lock()
		for _, srcExecBot := range srcExecBots {
			if _, has := runner.runningExecBots[srcExecBot.Id]; has {
				continue
			}
			execBot, has := runner.suspendExecBots[srcExecBot.Id]
			if !has {
				continue
			}
			if execBot.isSuspend {
				execBot.Resume(srcExecBot)
			}
		}
		runner.Unlock()
	}
}

func (runner *BotRunner) CleanIndexings() {
	runner.Lock()
	defer runner.Unlock()
	var rmArr []string

	rmArr = make([]string, 0)
	for key := range runner.triggerBots {
		if len(runner.triggerBots[key]) == 0 {
			rmArr = append(rmArr, key)
		}
	}
	for _, key := range rmArr {
		delete(runner.triggerBots, key)
	}

	rmArr = make([]string, 0)
	for key := range runner.triggerExecBots {
		if len(runner.triggerExecBots[key]) == 0 {
			rmArr = append(rmArr, key)
		}
	}
	for _, key := range rmArr {
		delete(runner.triggerExecBots, key)
	}

	rmArr = make([]string, 0)
	for key := range runner.execBotTriggers {
		if len(runner.execBotTriggers[key]) == 0 {
			rmArr = append(rmArr, key)
		}
	}
	for _, key := range rmArr {
		delete(runner.execBotTriggers, key)
	}

	rmArr = make([]string, 0)
	for key := range runner.objExecBots {
		if len(runner.objExecBots[key]) == 0 {
			rmArr = append(rmArr, key)
		}
	}
	for _, key := range rmArr {
		delete(runner.objExecBots, key)
	}

	rmArrInt := make([]int64, 0)
	for key := range runner.schedExecBots {
		if len(runner.schedExecBots[key]) == 0 {
			rmArrInt = append(rmArrInt, key)
		}
	}
	for _, key := range rmArrInt {
		delete(runner.schedExecBots, key)
	}
}

func (runner *BotRunner) ResetBotReport(botId string, metrics []*header.Metric) {
	runner.Lock()
	defer runner.Unlock()
	logicBot, has := runner.logicBots[botId]
	if !has {
		return
	}
	if logicBot.Report == nil {
		logicBot.Report = NewReportRun(runner.AccountId, logicBot.Id)
	}
	logicBot.Report.Reset(metrics)
}

// Ref with load bots
func (runner *BotRunner) NewBot(srcBot *header.Bot, srcBotReport *ReportRun) {
	runner.Lock()
	defer runner.Unlock()

	bot := runner.logicBots[srcBot.Id].GetBot()
	if bot != nil && bot.Updated >= srcBot.Updated {
		return // ignore
	}

	if bot == nil {
		logicBot := &LogicBot{AccountId: srcBot.GetAccountId(), Id: srcBot.Id}
		logicBot.SetBot(srcBot)
		logicBot.SetReport(srcBotReport)
		runner.logicBots[srcBot.Id] = logicBot
		runner.addTriggerBots(logicBot.Bot)
	} else {
		runner.rmTriggerBots(bot)
		logicBot := &LogicBot{AccountId: srcBot.GetAccountId(), Id: srcBot.Id}
		logicBot.SetBot(srcBot)
		botReport := runner.logicBots[srcBot.Id].GetReport()
		if botReport == nil {
			botReport = srcBotReport
		}
		logicBot.SetReport(botReport)
		runner.logicBots[srcBot.Id] = logicBot
		runner.addTriggerBots(logicBot.Bot)
	}
}

// Update bots (add, edit, remove)
// Update triggers depend on bot
// Require all bots
func (runner *BotRunner) LoadBots(srcBots []*header.Bot, srcBotReports []*ReportRun) {
	runner.Lock()
	defer runner.Unlock()

	reportMap := make(map[string]*ReportRun)
	for _, report := range srcBotReports {
		reportMap[report.BotId] = report
	}
	isexists := make(map[string]struct{}, len(srcBots))

	for _, srcBot := range srcBots {
		if srcBot == nil {
			continue
		}

		isexists[srcBot.Id] = struct{}{}

		bot := runner.logicBots[srcBot.Id].GetBot()
		if bot != nil && bot.Updated >= srcBot.Updated {
			continue // ignore
		}

		if bot == nil {
			logicBot := &LogicBot{AccountId: srcBot.GetAccountId(), Id: srcBot.Id}
			logicBot.SetBot(srcBot)
			logicBot.SetReport(reportMap[srcBot.Id])
			runner.logicBots[srcBot.Id] = logicBot
			runner.addTriggerBots(logicBot.Bot)
		} else {
			runner.rmTriggerBots(bot)
			logicBot := &LogicBot{AccountId: srcBot.GetAccountId(), Id: srcBot.Id}
			logicBot.SetBot(srcBot)
			botReport := runner.logicBots[srcBot.Id].GetReport()
			if botReport == nil {
				botReport = reportMap[srcBot.Id]
			}
			logicBot.SetReport(botReport)
			runner.logicBots[srcBot.Id] = logicBot
			runner.addTriggerBots(logicBot.Bot)
		}
	}

	for id, logicBot := range runner.logicBots {
		if _, has := isexists[id]; !has {
			runner.rmTriggerBots(logicBot.GetBot())
			delete(runner.logicBots, id)
		}
	}
}

// Require all bots
func (runner *BotRunner) LoadExecBots(srcExecBots []*ExecBot) {
	runner.Lock()
	defer runner.Unlock()
	for _, srcExecBot := range srcExecBots {
		if srcExecBot.ActionTree == nil {
			continue
		}
		if _, has := runner.runningExecBots[srcExecBot.Id]; has {
			continue
		}
		if srcExecBot.Status == Exec_bot_state_terminated {
			continue
		}
		logicBot, hasLogicBot := runner.logicBots[srcExecBot.BotId]
		if !hasLogicBot || (logicBot.GetActionMaph() != srcExecBot.ActionTree.ActionHash) {
			log.Info(srcExecBot.ObjId, "exec-bot.terminate", fmt.Sprintf("account-id=%s bot-id=%s", srcExecBot.AccountId, srcExecBot.BotId), fmt.Sprintf("bot-h=%s tree-h=%s", logicBot.GetActionMaph(), srcExecBot.ActionTree.ActionHash))
			srcExecBot.onTerminate = runner.onTerminateExecBot
			srcExecBot.Terminate(header.BotTerminated_self.String())
			continue
		}

		srcExecBot.Bot = logicBot.GetBot()
		now := time.Now().UnixNano() / 1e6
		if now > srcExecBot.LastSeen+Runner_lock_time_out {
			if _, has := runner.suspendExecBots[srcExecBot.Id]; !has {
				execBot := NewExecBot(srcExecBot, nil)
				runner.addFuncExecBot(execBot)
				runner.suspendExecBots[execBot.Id] = execBot.Suspend()
				if runner.objExecBots[execBot.ObjId] == nil {
					runner.objExecBots[execBot.ObjId] = make([]string, 0)
				}
				runner.objExecBots[execBot.ObjId] = append(runner.objExecBots[execBot.ObjId], execBot.Id)
			}
			continue
		}

		if _, has := runner.execBots[srcExecBot.Id]; !has {
			runner.execBots[srcExecBot.Id] = NewExecBot(srcExecBot, nil)
			runner.addFuncExecBot(runner.execBots[srcExecBot.Id])
			runner.addTriggerExecBots(runner.execBots[srcExecBot.Id])
			runner.addSchedExecBots(runner.execBots[srcExecBot.Id])
		}
	}
}

func (runner *BotRunner) addTriggerBots(bot *header.Bot) {
	triggerIds := TriggerOfBot(bot, runner.actionMgr)
	for _, triggerId := range triggerIds {
		if _, has := runner.triggerBots[triggerId]; !has {
			runner.triggerBots[triggerId] = make([]string, 0)
		}

		runner.triggerBots[triggerId] = append(runner.triggerBots[triggerId], bot.Id)
	}
}

func (runner *BotRunner) rmTriggerBots(bot *header.Bot) {
	triggerIds := TriggerOfBot(bot, runner.actionMgr)
	var index int
	for _, triggerId := range triggerIds {
		if runner.triggerBots[triggerId] == nil {
			continue
		}
		for i, id := range runner.triggerBots[triggerId] {
			if id == bot.Id {
				index = i + 1
				break
			}
		}

		if index > 0 {
			runner.triggerBots[triggerId] = removeIndex(runner.triggerBots[triggerId], index-1)
			index = 0 // reset
		}
		if len(runner.triggerBots[triggerId]) == 0 {
			delete(runner.triggerBots, triggerId)
		}
	}
}

// Exec bot register trigger, update trigger
func (runner *BotRunner) addTriggerExecBots(execBot *ExecBot) {
	triggerIds := TriggerOfHeadNode(execBot, runner.actionMgr) // or TriggerOfExecBot
	for _, triggerId := range triggerIds {
		if _, has := runner.triggerExecBots[triggerId]; !has {
			runner.triggerExecBots[triggerId] = make([]string, 0)
		}
		runner.triggerExecBots[triggerId] = append(runner.triggerExecBots[triggerId], execBot.Id)
	}

	runner.execBotTriggers[execBot.Id] = triggerIds
}

func (runner *BotRunner) rmTriggerExecBots(execBotId string) {
	if runner.execBotTriggers[execBotId] == nil {
		return
	}

	var index int
	for _, triggerId := range runner.execBotTriggers[execBotId] {
		for i, id := range runner.triggerExecBots[triggerId] {
			if id == execBotId {
				index = i + 1
				break
			}
		}

		if index > 0 {
			runner.triggerExecBots[triggerId] = removeIndex(runner.triggerExecBots[triggerId], index-1)
			index = 0 // reset
		}
	}

	delete(runner.execBotTriggers, execBotId)
}

func (runner *BotRunner) addSchedExecBots(execBot *ExecBot) {
	if execBot.SchedActs == nil {
		execBot.SchedActs = make(map[string]int64)
	}
	if len(execBot.SchedActs) == 0 {
		return
	}

	scheduleArr := make([]ActionSchedule, len(execBot.SchedActs))
	index := 0
	for actId, scheduleAt := range execBot.SchedActs {
		scheduleArr[index] = ActionSchedule{
			ActionId:   actId,
			ScheduleAt: scheduleAt,
		}
		index++
	}
	sort.Slice(scheduleArr, func(i, j int) bool { return scheduleArr[i].ScheduleAt < scheduleArr[j].ScheduleAt })

	schedAtHead := time.Now().UnixNano() / 1e6 / 1e2
	for i := 0; i < len(scheduleArr); i++ {
		if scheduleArr[i].ScheduleAt >= schedAtHead {
			schedAtHead = scheduleArr[i].ScheduleAt
			break
		}
	}

	if runner.schedExecBots[schedAtHead] == nil {
		runner.schedExecBots[schedAtHead] = make([]string, 0)
	}
	runner.schedExecBots[schedAtHead] = append(runner.schedExecBots[schedAtHead], execBot.Id)

	if runner.schedAt != 0 && runner.schedAt <= schedAtHead {
		return
	}

	isScheduled := runner.sched.Push(schedAtHead*1e2, runner.AccountId)
	if isScheduled {
		runner.schedAt = schedAtHead
	}
}

func (runner *BotRunner) addFuncExecBot(execBot *ExecBot) {
	execBot.afterRun = runner.AfterRunExecBot
	execBot.actionMgr = runner.actionMgr
	execBot.onStart = func(execBot *ExecBot) {
		if runner.execBotMgr == nil {
			log.Warn("NILLLLLLL")
			return
		}
		execBotIndex := execBot.GetExecBotIndex()
		if execBotIndex == nil {
			return
		}
		createdExecBotIndex, err := runner.execBotMgr.CreateExecBotIndex(execBotIndex)
		if err != nil {
			log.Error(execBot.ObjId, "exec-bot-mgr.create-exec-bot-index", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("err=%s", err))
		}
		if createdExecBotIndex != nil {
			execBot.ActionTree.IndexIncre = createdExecBotIndex.IndexIncre
		}
		runner.Lock()
		logicBot := runner.logicBots[execBot.BotId]
		if logicBot != nil && logicBot.Report != nil {
			logicBot.Report.AddObject(execBotIndex.BeginTimeDay)
		}
		runner.Unlock()
	}
	execBot.onConvertLead = func(execBot *ExecBot) {
		if runner.execBotMgr == nil {
			log.Warn("NILLLLLLL")
			return
		}
		execBotIndex := execBot.GetExecBotIndex()
		if execBotIndex != nil && execBot.ActionTree.IndexIncre > 0 {
			go runner.execBotMgr.CreateExecBotIndex(execBotIndex)
			runner.Lock()
			logicBot := runner.logicBots[execBot.BotId]
			if logicBot != nil && logicBot.Report != nil {
				logicBot.Report.AddLead(execBotIndex.BeginTimeDay)
			}
			runner.Unlock()
		}
	}
	execBot.onTerminate = runner.onTerminateExecBot
}

func (runner *BotRunner) onTerminateExecBot(execBot *ExecBot) {
	if runner.execBotMgr == nil {
		log.Warn("NILLLLLLL")
		return
	}
	var execBotIndex *ExecBotIndex
	if execBot.ActionTree != nil && execBot.ActionTree.IndexIncre > 0 {
		execBotIndex = execBot.GetExecBotIndex()
	}
	raw := execBot.ToRaw()
	go func(execBot *ExecBot, execBotIndex *ExecBotIndex) {
		if _, err := runner.execBotMgr.UpdateExecBot(raw); err != nil {
			log.Error(execBot.ObjId, "exec-bot-mgr.create-exec-bot-index", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("err=%s", err))
			return
		}
		if execBotIndex == nil {
			return
		}
		if _, err := runner.execBotMgr.CreateExecBotIndex(execBotIndex); err != nil {
			log.Error(execBot.ObjId, "exec-bot-mgr.create-exec-bot-index", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("err=%s", err))
		}
		if runner.ObjectMgr != nil {
			allperm := perm.MakeBase()
			cred := ggrpc.ToGrpcCtx(&pb.Context{Credential: &pb.Credential{AccountId: execBot.AccountId, Issuer: execBot.BotId, Type: pb.Type_agent, Perm: &allperm}})
			var success bool
			if execBot.StatusNote == header.BotTerminated_complete.String() {
				success = true
			}
			_, err := runner.ObjectMgr.TerminateBot(cred, &header.BotTerminated{AccountId: execBot.AccountId, BotId: execBot.BotId, ConversationId: execBot.ObjId, Success: success, Code: execBot.StatusNote})
			if err != nil {
				log.Error(execBot.ObjId, "object-mgr.terminate-bot", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("err=%s", err))
			}
		}
	}(raw, execBotIndex)
}

func (runner *BotRunner) autoStartBot(in *header.RunRequest, runnable map[string]*ExecBot) {
	objType := in.GetObjectType()
	objId := in.GetObjectId()
	evt := in.GetEvent()
	botTriggers := TriggerOnBotEvent(in)
	for _, trigger := range botTriggers {
		for _, botId := range runner.triggerBots[trigger] {
			logicBot := runner.logicBots[botId]
			bot, actionMaph := logicBot.GetBot(), logicBot.GetActionMaph()
			if bot == nil {
				continue
			}
			if objType != bot.GetCategory() {
				continue
			}

			execBotId := NewExecBotID(bot.GetAccountId(), bot.GetId(), objId)
			if _, has := runnable[execBotId]; has {
				continue
			}
			if !RunCondition(bot, in) {
				continue
			}

			// don't start exec bot has condition failure
			// why,when seed not wait at first time
			node := NewActionNode(bot.Action)
			node.Events = []*header.Event{evt}
			state := runner.actionMgr.Wait(node)
			if state != Node_state_ready {
				continue
			}

			execBot, err := MakeExecBot(bot, actionMaph, objId, execBotId, in.GetObjectContexts())
			if err == nil {
				// require evt of new exec bot
				runner.addFuncExecBot(execBot)

				runnable[execBotId] = execBot
				runner.execBots[execBotId] = execBot
				runner.addTriggerExecBots(execBot) // Keep trigger
			}
		}
	}
}

func runExecBot(execBot *ExecBot, req *header.RunRequest, mgr ExecBotMgr, logicBot *LogicBot) {
	// don't support action map, hash
	if execBot.mode != Exec_bot_mode_debug {
		// continuous
		bot := logicBot.GetBot()
		if bot != nil && execBot.BotId == bot.Id && execBot.ObjType == bot.GetCategory() {
			execBot.Bot = bot
			execBot.actionMap = logicBot.GetActionMap()
			execBot.actionMaph = logicBot.GetActionMaph()
		}
	}

	// is blocked, reload from db
	if execBot.isSuspend {
		if mgr == nil {
			log.Warn("NILLLLLLL")
			return
		}

		// TODO1 consume memory
		raw, err := mgr.ReadExecBot(execBot.BotId, execBot.Id)
		if err != nil {
			log.Error(err)
			return
		}
		// TODO1
		if raw == nil {
			log.Error(execBot.ObjId, "exec-bot-mgr.read", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), "err=NILLLLLLL")
			return
		}
		execBot.Resume(raw)
	}
	log.Info(execBot.ObjId, "exec-bot.run", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("req.run-type=%s", req.GetBotRunType()))
	execBot.Run([]*header.RunRequest{req})
}

func removeIndex(s []string, i int) []string {
	if len(s) < i {
		return s
	}
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}

const (
	Runner_lock_time_out    = 3600000  // millisecond of hour 1 * 60 * 60 * 1e3
	Runner_lock_time_expire = 86400000 // millisecond of day 1 * 24 * 60 * 60 * 1e3
	Runner_forgot_time      = 864000   // decisecond of day 24 * 60 * 60
)

// don't support update exec bots
func ApplyExecBot(bot *header.Bot) {
	// don't require, because new exec bot run when created
}

type ActionSchedule struct {
	ActionId   string
	ScheduleAt int64
}

type ReportRun struct {
	AccountId string
	BotId     string
	Metrics   []*header.Metric
	isChange  bool
}

func NewReportRun(accountId, botId string) *ReportRun {
	return &ReportRun{
		AccountId: accountId,
		BotId:     botId,
		Metrics:   make([]*header.Metric, ReportRun_length),
	}
}

func (report *ReportRun) Reset(srcMetrics []*header.Metric) {
	for _, srcMetric := range srcMetrics {
		index := srcMetric.DateDim % ReportRun_length
		metric := report.Metrics[index]
		if metric == nil || metric.DateDim <= srcMetric.DateDim {
			report.Metrics[index] = srcMetric
			report.isChange = true
		}
	}
}

func (report *ReportRun) ShouldUpdate() bool {
	defer func() { report.isChange = false }()
	return report.isChange
}

func (report *ReportRun) AddLead(unixDay int64) {
	index := unixDay % ReportRun_length
	metric := report.Metrics[index]
	if metric == nil || metric.DateDim < unixDay {
		metric = &header.Metric{DateDim: unixDay, ObjectCount: 0, LeadCount: 0}
		report.Metrics[index] = metric
	}
	metric.LeadCount++
	report.isChange = true
}

func (report *ReportRun) AddObject(unixDay int64) {
	index := unixDay % ReportRun_length
	metric := report.Metrics[index]
	if metric == nil || metric.DateDim < unixDay {
		metric = &header.Metric{DateDim: unixDay, ObjectCount: 0, LeadCount: 0}
		report.Metrics[index] = metric
	}
	metric.ObjectCount++
	report.isChange = true
}

func (report *ReportRun) List(dayFrom int64, dayTo int64) []*header.Metric {
	if report == nil {
		return []*header.Metric{}
	}
	arr := make([]*header.Metric, 0)
	for _, metric := range report.Metrics {
		if metric == nil {
			continue
		}
		if metric.DateDim >= dayFrom && metric.DateDim <= dayTo {
			arr = append(arr, &header.Metric{DateDim: metric.DateDim, ObjectCount: metric.ObjectCount, LeadCount: metric.LeadCount})
		}
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].DateDim > arr[j].DateDim })
	return arr
}

func (report *ReportRun) Get(unixDay int64) *header.Metric {
	metric := report.Metrics[unixDay%ReportRun_length]
	if metric != nil {
		return &header.Metric{DateDim: metric.DateDim, ObjectCount: metric.ObjectCount, LeadCount: metric.LeadCount}
	}
	return nil
}

const (
	ReportRun_length = 90
)

type LogicBot struct {
	AccountId  string
	Id         string
	Bot        *header.Bot
	actionMaph string
	actionMap  map[string]*header.Action
	Report     *ReportRun
}

func (l *LogicBot) SetReport(raw *ReportRun) {
	if raw != nil {
		l.Report = raw
	}
	if l.Report == nil {
		l.Report = NewReportRun(l.AccountId, l.Id)
	}
}

func (l *LogicBot) SetBot(bot *header.Bot) {
	if l.Id != bot.Id {
		return
	}
	l.Bot = bot
	h, m := hashAction(bot.GetAction())
	l.actionMaph = h
	l.actionMap = m
}

func (l *LogicBot) GetBot() *header.Bot {
	if l != nil {
		return l.Bot
	}
	return nil
}

func (l *LogicBot) GetReport() *ReportRun {
	if l != nil {
		return l.Report
	}
	return nil
}

func (l *LogicBot) GetActionMap() map[string]*header.Action {
	if l != nil {
		return l.actionMap
	}
	return nil
}

func (l *LogicBot) GetActionMaph() string {
	if l != nil {
		return l.actionMaph
	}
	return ""
}

func hashAction(action *header.Action) (string, map[string]*header.Action) {
	rootIdaction := &header.Action{Id: action.GetId()}
	flatActions := map[string]*header.Action{}
	flatActions[action.GetId()] = action
	idactions := []*header.Action{rootIdaction}
	deep := 0
	nodeIndex := 0
	if action.GetId() == "" {
		action.Id = strconv.Itoa(nodeIndex)
		nodeIndex++
	}
	var flatAction *header.Action
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			break
		}
		widthIdactions := make([]*header.Action, 0)
		for _, idaction := range idactions {
			flatAction = flatActions[idaction.GetId()]
			if flatAction == nil || len(flatAction.GetNexts()) == 0 {
				continue
			}
			idaction.Nexts = make([]*header.NextAction, 0)
			for _, actNext := range flatAction.GetNexts() {
				if actNext.GetAction() == nil {
					continue
				}
				if actNext.GetAction().GetId() == "" {
					actNext.Action.Id = strconv.Itoa(nodeIndex)
					nodeIndex++
				}
				flatActions[actNext.GetAction().GetId()] = actNext.GetAction()
				nextIdaction := &header.Action{Id: actNext.GetAction().GetId()}
				idaction.Nexts = append(idaction.Nexts, &header.NextAction{Action: nextIdaction})
				widthIdactions = append(widthIdactions, nextIdaction)
			}
		}
		if len(widthIdactions) == 0 {
			break
		}
		idactions = widthIdactions
		deep++
	}

	actionb, _ := proto.Marshal(rootIdaction)
	hasher := sha1.New()
	hasher.Write([]byte(actionb))
	return hex.EncodeToString(hasher.Sum(nil)), flatActions
}

type BotRunnerStatus struct {
	*sync.Mutex
	AccountId           string
	LogicBotCount       int
	TriggerBotCount     int
	ExecBotCount        int
	TriggerExecBotCount int
	ExecBotTriggerCount int
	SuspendExecBotCount int
	ObjExecBotCount     int
	RunningExecBotCount int
	SchedExecBotCount   int
	SourceStatus        []*BotRunnerStatus
}

func NewBotRunnerStatus(accountId string) *BotRunnerStatus {
	return &BotRunnerStatus{
		AccountId:    accountId,
		Mutex:        &sync.Mutex{},
		SourceStatus: make([]*BotRunnerStatus, 0),
	}
}

func (s *BotRunnerStatus) Add(src *BotRunnerStatus) {
	s.Lock()
	s.Unlock()
	s.LogicBotCount += src.LogicBotCount
	s.TriggerBotCount += src.TriggerBotCount
	s.ExecBotCount += src.ExecBotCount
	s.TriggerExecBotCount += src.TriggerExecBotCount
	s.ExecBotTriggerCount += src.ExecBotTriggerCount
	s.SuspendExecBotCount += src.SuspendExecBotCount
	s.ObjExecBotCount += src.ObjExecBotCount
	s.RunningExecBotCount += src.RunningExecBotCount
	s.SchedExecBotCount += src.SchedExecBotCount
	s.SourceStatus = append(s.SourceStatus, src)
}

// Diff way actions schedule, E.g. with 2 queue
// 1. Foreground(80% resources): round robin serve
// 2. Background(20% resources): first come first serve
// Note: resources are cpu, mem
// Run exec bot
// 1. Trigger>Bot is process
// 2. ExecBot>Action is subprocess

// TODO shard bot exec by obj id
// TODO grows
// TODO1 not suspend and suspend
