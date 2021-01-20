package main

import (
	"crypto/sha1"
	"encoding/hex"
	"hash/crc32"
	"strconv"
	"sync"
	"time"

	cr "git.subiz.net/bizbot/runner"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
	cpb "github.com/subiz/header/common"
	"github.com/subiz/idgen"
)

type BotDB interface {
	UpsertBot(*header.Bot) error
	ReadAllBots(accid string) ([]*header.Bot, error)
	ReadBot(accid, id string) (*header.Bot, error)
	ReadBotReport(accid, id string) (*cr.ReportRun, error)
	DeleteBot(accid, id string) error
	UpsertBotTag(b *header.Bot, tag string) error
	ReadBotTag(accid, id, tag string) (*header.Bot, error)
	ReadAllTags(accid, id string) ([]*header.Bot, error)
}

type BotReportMgr interface {
	AggregateExecBotIndex(botId string, dayFrom, dayTo int64) []*header.Metric
}

type BotMgr struct {
	*sync.Mutex
	AccountId string
	db        BotDB
	report    BotReportMgr
}

func NewBotMgr(accountId string, db BotDB, report BotReportMgr) *BotMgr {
	mgr := &BotMgr{
		Mutex:     &sync.Mutex{},
		AccountId: accountId,
		db:        db,
		report:    report,
	}
	return mgr
}

func (mgr *BotMgr) UpdateBot(bot *header.Bot) (*header.Bot, error) {
	if _, err := cr.MakeTree(bot.GetAction()); err != nil {
		return nil, err
	}
	// TODO bot revision
	// bot.ActionHash = hashAction(bot.GetAction())
	if err := mgr.db.UpsertBot(bot); err != nil {
		return nil, err
	}
	b, err := mgr.db.ReadBot(bot.GetAccountId(), bot.GetId())
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (mgr *BotMgr) CreateBot(bot *header.Bot) (*header.Bot, error) {
	bot.AccountId = mgr.AccountId
	if bot.GetId() == "" {
		bot.Id = idgen.NewBizbotID()
	}
	now := time.Now().UnixNano() / 1e6
	bot.Created = now
	bot.Updated = now
	// default obj type
	if bot.Category == "" {
		bot.Category = header.BotCategory_conversations.String()
	}

	if _, err := cr.MakeTree(bot.GetAction()); err != nil {
		return nil, err
	}
	// TODO bot revision
	// bot.ActionHash = hashAction(bot.GetAction())
	if err := mgr.db.UpsertBot(bot); err != nil {
		return nil, err
	}

	b, err := mgr.db.ReadBot(bot.GetAccountId(), bot.GetId())
	if err != nil {
		return nil, err
	}
	return b, nil
}

func hashAction(action *header.Action) string {
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
		if deep >= 128 {
			log.Warn("TOODEEPPP") // ref Bizbot_limit_act_deep
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
	return hex.EncodeToString(hasher.Sum(nil))
}

func (mgr *BotMgr) ListBots(id *cpb.Id) ([]*header.Bot, error) {
	bots, err := mgr.db.ReadAllBots(mgr.AccountId)
	if err != nil {
		return nil, err
	}
	dayTo := time.Now().Unix() / (24 * 3600)
	dayFrom := dayTo - 6
	wg := &sync.WaitGroup{}
	wg.Add(len(bots))
	for _, bot := range bots {
		go func(bot *header.Bot) {
			defer wg.Done()
			botReport, err := mgr.db.ReadBotReport(mgr.AccountId, bot.GetId())
			if err != nil {
				log.Error(err)
				return
			}
			var totalLead, totalObject int64
			for _, metric := range botReport.List(dayFrom, dayTo) {
				totalLead += metric.LeadCount
				totalObject += metric.ObjectCount
			}
			bot.CountLeadInLast_7Days = totalLead
			bot.CountConvInLast_7Days = totalObject
		}(bot)
	}
	wg.Wait()
	return bots, nil
}

func (mgr *BotMgr) GetBotReport(id *cpb.Id) (*cr.ReportRun, error) {
	botReport, err := mgr.db.ReadBotReport(mgr.AccountId, id.GetId())
	if err != nil {
		return nil, err
	}
	return botReport, nil
}

func (mgr *BotMgr) GetBot(id *cpb.Id) (*header.Bot, error) {
	bot, err := mgr.db.ReadBot(mgr.AccountId, id.GetId())
	if err != nil {
		return nil, err
	}
	dayTo := time.Now().Unix() / (24 * 3600)
	dayFrom := dayTo - 6
	botReport, err := mgr.db.ReadBotReport(mgr.AccountId, id.GetId())
	if err != nil {
		return nil, err
	}
	var totalLead, totalObject int64
	for _, metric := range botReport.List(dayFrom, dayTo) {
		totalLead += metric.LeadCount
		totalObject += metric.ObjectCount
	}
	bot.CountLeadInLast_7Days = totalLead
	bot.CountConvInLast_7Days = totalObject
	return bot, nil
}

func (mgr *BotMgr) DeleteBot(id *cpb.Id) error {
	var err error
	var bot *header.Bot
	bot, err = mgr.db.ReadBot(mgr.AccountId, id.GetId())
	if err != nil {
		return err
	}
	if bot == nil {
		return nil
	}

	err = mgr.db.UpsertBotTag(bot, "deleted")
	if err != nil {
		return err
	}

	err = mgr.db.DeleteBot(mgr.AccountId, id.GetId())
	if err != nil {
		return err
	}
	return nil
}

func (mgr *BotMgr) CreateBotTag(bot *header.Bot) (*header.Bot, error) {
	bot.AccountId = mgr.AccountId
	if err := mgr.db.UpsertBotTag(bot, bot.Version); err != nil {
		return nil, err
	}
	b, err := mgr.db.ReadBotTag(bot.GetAccountId(), bot.GetId(), bot.GetVersion())
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (mgr *BotMgr) ListBotTags(id *cpb.Id) ([]*header.Bot, error) {
	bot, err := mgr.db.ReadAllTags(mgr.AccountId, id.GetId())
	if err != nil {
		return nil, err
	}
	return bot, nil
}

type ExecBotDB interface {
	UpsertExecBot(*cr.ExecBot) error
	ReadExecBot(accid, botid, id string) (*cr.ExecBot, error)
	ReadExecBots(accid, botid string) ([]*cr.ExecBot, error)
	DeleteExecBot(accid, botid, id string) error
	UpsertExecBotIndex(execBotIndex *cr.ExecBotIndex) error
	LastExecBotIndex(accountId, botId, id string) (int64, error)
	LoadExecBotIndexs(accountId, botId string, beginTimeDayFrom, beginTimeDayTo int64, orderDesc []string, c chan []*cr.ExecBotIndex, size int) error
	LoadExecBots(accountId string, c chan []*cr.ExecBot, size int, botPKMap map[string][]*cr.ExecBotPK) error
}

type ExecBotMgr struct {
	*sync.Mutex
	AccountId   string
	db          ExecBotDB
	shardNumber int
	mutexShards []*sync.Mutex // shard by exec bot id
	realtime    header.PubsubClient
}

func NewExecBotMgr(accountId string, db ExecBotDB, realtime header.PubsubClient) *ExecBotMgr {
	mgr := &ExecBotMgr{
		Mutex:       &sync.Mutex{},
		AccountId:   accountId,
		db:          db,
		realtime:    realtime,
		shardNumber: 20,
	}
	mgr.mutexShards = make([]*sync.Mutex, mgr.shardNumber)
	for i := 0; i < mgr.shardNumber; i++ {
		mgr.mutexShards[i] = &sync.Mutex{}
	}
	return mgr
}

func (mgr *ExecBotMgr) ReadExecBot(botId, id string) (*cr.ExecBot, error) {
	out, err := mgr.db.ReadExecBot(mgr.AccountId, botId, id)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (mgr *ExecBotMgr) UpdateExecBot(execBot *cr.ExecBot) (*cr.ExecBot, error) {
	execBot.Updated = time.Now().UnixNano() / 1e6
	if err := mgr.db.UpsertExecBot(execBot); err != nil {
		return nil, err
	}

	out, err := mgr.db.ReadExecBot(execBot.AccountId, execBot.BotId, execBot.Id)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (mgr *ExecBotMgr) CreateExecBot(execBot *cr.ExecBot) (*cr.ExecBot, error) {
	execBot.AccountId = mgr.AccountId
	if execBot.Id == "" {
		execBot.Id = cr.NewExecBotID(execBot.AccountId, execBot.BotId, execBot.ObjId)
	}
	now := time.Now().UnixNano() / 1e6
	execBot.Created = now
	execBot.Updated = now

	if err := mgr.db.UpsertExecBot(execBot); err != nil {
		return nil, err
	}
	out, err := mgr.db.ReadExecBot(execBot.AccountId, execBot.BotId, execBot.Id)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (mgr *ExecBotMgr) DeleteExecBots(id *cpb.Id) error {
	execBots, err := mgr.db.ReadExecBots(mgr.AccountId, id.GetId())
	if err != nil {
		return err
	}
	for _, execBot := range execBots {
		if err := mgr.db.DeleteExecBot(execBot.AccountId, execBot.BotId, execBot.Id); err != nil {
			return err
		}
	}
	return nil
}

func (mgr *ExecBotMgr) CreateExecBotIndex(execBotIndex *cr.ExecBotIndex) (*cr.ExecBotIndex, error) {
	execIdNum := int(crc32.ChecksumIEEE([]byte(execBotIndex.Id)))
	shard := mgr.mutexShards[execIdNum%mgr.shardNumber]
	shard.Lock()
	defer shard.Unlock()

	if execBotIndex.IndexIncre == 0 {
		lastIndex, err := mgr.db.LastExecBotIndex(execBotIndex.AccountId, execBotIndex.BotId, execBotIndex.Id)
		if lastIndex >= 10 {
			log.Warn("TOOINDEXXX")
		}
		if err != nil {
			return nil, err
		}
		execBotIndex.IndexIncre = lastIndex + 1
	}
	execBotIndex.AccountId = mgr.AccountId
	now := time.Now().UnixNano() / 1e6
	execBotIndex.Created = now
	execBotIndex.Updated = now

	if err := mgr.db.UpsertExecBotIndex(execBotIndex); err != nil {
		return nil, err
	}
	return execBotIndex, nil
}

func (mgr *ExecBotMgr) AggregateExecBotIndex(botId string, beginTimeDayFrom, beginTimeDayTo int64) []*header.Metric {
	c := make(chan []*cr.ExecBotIndex)
	go func() {
		err := mgr.db.LoadExecBotIndexs(
			mgr.AccountId,
			botId,
			beginTimeDayFrom,
			beginTimeDayTo,
			[]string{"begin_time_day", "id"},
			c,
			100,
		)
		if err != nil {
			log.Error(err)
		}
	}()

	var objType string
	var dateDime int64
	var objId string
	var objLead bool

	metrics := make([]*header.Metric, 0)
	index := 0
	subindex := 0

	var metric *header.Metric
	// TODO1 rm
	var submetric *header.Metric

	metric = &header.Metric{}
	submetric = &header.Metric{}
	for execBotIndexs := range c {
		for _, execBotIndex := range execBotIndexs {
			if objId != execBotIndex.ObjId {
				submetric.ObjectCount += 1
				metric.ObjectCount += 1
				if objLead {
					submetric.LeadCount += 1
					metric.LeadCount += 1
				}
				objId = execBotIndex.ObjId
				objLead = false
			}

			if dateDime != execBotIndex.BeginTimeDay {
				metrics = append(metrics, &header.Metric{
					DateDim:    execBotIndex.BeginTimeDay,
					Submetrics: []*header.Metric{},
				})
				subindex = 0
				dateDime = execBotIndex.BeginTimeDay
				metric = metrics[index]
				index++
			}

			if objType != execBotIndex.ObjType {
				metric.Submetrics = append(metric.Submetrics, &header.Metric{
					ObjectType: execBotIndex.ObjType,
				})
				objType = execBotIndex.ObjType
				submetric = metric.Submetrics[subindex]
				subindex++
			}

			submetric.Count += 1
			metric.Count += 1
			if len(execBotIndex.LeadPayload) > 0 {
				objLead = true
			}
		}
	}

	submetric.ObjectCount += 1
	metric.ObjectCount += 1
	if objLead {
		submetric.LeadCount += 1
		metric.LeadCount += 1
	}

	return metrics
}

// TODO duplicate obj id
func (mgr *ExecBotMgr) ListObjectIds(botId string, beginTimeDayFrom, beginTimeDayTo int64) []*header.ListObjectId {
	c := make(chan []*cr.ExecBotIndex)
	go func() {
		// order by begin_time_day desc, id desc
		err := mgr.db.LoadExecBotIndexs(
			mgr.AccountId,
			botId,
			beginTimeDayFrom,
			beginTimeDayTo,
			[]string{"id"},
			c,
			100,
		)
		if err != nil {
			log.Error(err)
		}
	}()

	var objType string
	var objId string
	listObjIds := make([]*header.ListObjectId, 0)
	index := -1

	for execBotIndexs := range c {
		for _, execBotIndex := range execBotIndexs {
			if objType != execBotIndex.ObjType {
				listObjIds = append(listObjIds, &header.ListObjectId{
					ObjectType: execBotIndex.ObjType,
					ObjectIds:  make([]string, 0),
				})

				index++
				objType = execBotIndex.ObjType
			}

			if objId != execBotIndex.ObjId {
				listObjIds[index].ObjectIds = append(listObjIds[index].ObjectIds, execBotIndex.ObjId)
				objId = execBotIndex.ObjId
			}
		}
	}

	return listObjIds
}

func (mgr *ExecBotMgr) LoadExecBots(accountId string, pks []*cr.ExecBotPK) chan []*cr.ExecBot {
	botPKMap := make(map[string][]*cr.ExecBotPK)
	for _, pk := range pks {
		if pk == nil {
			continue
		}
		if pk.AccountId != accountId {
			continue
		}
		if _, has := botPKMap[pk.BotId]; !has {
			botPKMap[pk.BotId] = []*cr.ExecBotPK{pk}
		} else {
			botPKMap[pk.BotId] = append(botPKMap[pk.BotId], pk)
		}
	}
	c := make(chan []*cr.ExecBot)
	go func() {
		err := mgr.db.LoadExecBots(accountId, c, 100, botPKMap)
		if err != nil {
			log.Error(err)
		}
	}()
	return c
}
