package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"github.com/subiz/idgen"
)

// ExecBot state: new, ready, running, waiting, end
// ExecBot Id include (account id, bot id, obj id), optional index
// Key of wait evts is field.operator.value
type ExecBot struct {
	AccountId   string            `json:"account_id"`
	BotId       string            `json:"bot_id"`
	ObjType     string            `json:"obj_type"`
	ObjId       string            `json:"obj_id"`
	LeadId      string            `json:"lead_id"`
	LeadPayload map[string]string `json:"lead_payload"`
	Id          string            `json:"id"`
	Bot         *header.Bot       // required, direct
	Created     int64             `json:"created"`
	Updated     int64             `json:"updated"`
	LastSeen    int64             `json:"last_seen"`
	ActionTree  *ActionNode       `json:"action_tree"`
	ActionTreeb []byte            // only raw
	Status      string            `json:"status"`
	StatusNote  string            `json:"status_note"`
	RunTimes    []int64           `json:"run_times"`
	SchedActs   map[string]int64  `json:"sched_acts"`

	beforeRun     func()
	afterRun      func(*ExecBot)
	onTerminate   func(*ExecBot) // TODO rm
	onStart       func(*ExecBot)
	onConvertLead func(*ExecBot)
	isSuspend     bool
	actionMgr     ActionMgr
	raw           *ExecBot
	mode          string
	realtime      header.PubsubClient
	reqQueue      *RunRequestQueue
	actionMaph    string
	actionMap     map[string]*header.Action
}

// TODO ref make, require bot
func NewExecBot(raw *ExecBot, live *ExecBot) *ExecBot {
	if live == nil {
		live = &ExecBot{
			AccountId: raw.AccountId,
			BotId:     raw.BotId,
			ObjType:   raw.ObjType,
			ObjId:     raw.ObjId,
			LeadId:    raw.LeadId,
			Id:        raw.Id,
		}
	}

	live.merge(raw)
	live.ActionTree.Reboot(live)

	return live
}

func MakeExecBot(bot *header.Bot, actionMaph string, objId, id string, objCtxs []*pb.KV) (*ExecBot, error) {
	if bot == nil {
		return nil, errors.New("bot is nil")
	}
	if bot.Action == nil {
		return nil, errors.New("actions is nil")
	}

	rootNode, err := MakeTree(bot.Action)
	rootNode.ActionHash = actionMaph
	if err != nil {
		return nil, err
	}

	execBot := &ExecBot{}
	execBot.AccountId = bot.AccountId
	execBot.BotId = bot.Id
	execBot.ObjType = bot.GetCategory()
	execBot.ObjId = objId
	var leadId string
	for _, objCtx := range objCtxs {
		if objCtx.GetKey() == "user_id" {
			leadId = objCtx.GetValue()
			break
		}
	}
	execBot.LeadId = leadId
	execBot.LeadPayload = make(map[string]string)
	execBot.Id = id
	execBot.Bot = bot // direct replace clone
	execBot.Created = time.Now().UnixNano() / 1e6
	rootNode.SetExecBot(execBot)
	rootNode.ObjCtxs = objCtxs
	execBot.ActionTree = rootNode
	execBot.Status = Exec_bot_state_new
	execBot.RunTimes = make([]int64, 0)
	execBot.LastSeen = time.Now().UnixNano() / 1e6
	execBot.SchedActs = make(map[string]int64)

	return execBot, nil
}

func (execBot *ExecBot) ToRaw() *ExecBot {
	raw := &ExecBot{
		AccountId:  execBot.AccountId,
		BotId:      execBot.BotId,
		ObjType:    execBot.ObjType,
		ObjId:      execBot.ObjId,
		LeadId:     execBot.LeadId,
		Id:         execBot.Id,
		Created:    execBot.Created,
		Updated:    execBot.Updated,
		LastSeen:   execBot.LastSeen,
		Status:     execBot.Status,
		StatusNote: execBot.StatusNote,
		RunTimes:   execBot.RunTimes,
		ActionTree: execBot.ActionTree,
	}
	actionTreeb, err := json.Marshal(execBot.ActionTree)
	if err != nil {
		log.Error(execBot.ObjId, "json.marshal", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("err=%s", err))
	}
	raw.ActionTreeb = actionTreeb
	raw.LeadPayload = make(map[string]string, len(execBot.LeadPayload))
	for k, v := range execBot.LeadPayload {
		raw.LeadPayload[k] = v
	}
	raw.SchedActs = make(map[string]int64, len(execBot.SchedActs))
	for k, v := range execBot.SchedActs {
		raw.SchedActs[k] = v
	}
	return raw
}

func (execBot *ExecBot) GetExecBotIndex() *ExecBotIndex {
	if len(execBot.RunTimes) == 0 {
		return nil
	}

	return &ExecBotIndex{
		AccountId:    execBot.AccountId,
		BotId:        execBot.BotId,
		ObjType:      execBot.Bot.GetCategory(),
		ObjId:        execBot.ObjId,
		LeadId:       execBot.LeadId,
		LeadPayload:  execBot.LeadPayload,
		Id:           execBot.Id,
		IndexIncre:   execBot.ActionTree.IndexIncre,
		BeginTimeDay: execBot.RunTimes[0] / (24 * 3600 * 1e3),
		BeginTime:    execBot.RunTimes[0],
		EndTime:      execBot.RunTimes[len(execBot.RunTimes)-1],
		ExecBot: &ExecBot{
			AccountId:  execBot.AccountId,
			BotId:      execBot.BotId,
			ObjId:      execBot.ObjId,
			Id:         execBot.Id,
			Bot:        nil, // TODO bot tag
			Created:    execBot.Created,
			Updated:    execBot.Updated,
			LastSeen:   execBot.LastSeen,
			ActionTree: nil, // cut
			Status:     execBot.Status,
			StatusNote: execBot.StatusNote,
			RunTimes:   execBot.RunTimes,
		},
	}
}

func (execBot *ExecBot) Run(reqs []*header.RunRequest) {
	defer func() {
		if execBot.afterRun != nil {
			execBot.afterRun(execBot)
		}
	}()
	if execBot.beforeRun != nil {
		execBot.beforeRun()
	}

	if execBot.Status == Exec_bot_state_terminated {
		return
	}
	// slow terminate
	if execBot.actionMaph != execBot.ActionTree.ActionHash {
		log.Info(execBot.ObjId, "exec-bot.terminate", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("bot-h=%s tree-h=%s", execBot.actionMaph, execBot.ActionTree.ActionHash))
		execBot.Terminate(header.BotTerminated_self.String())
		return
	}

	now := time.Now().UnixNano() / 1e6
	execBot.LastSeen = now
	execBot.RunTimes = append(execBot.RunTimes, now)
	for _, req := range reqs {
		log.Info(execBot.ObjId, "action-tree.grow=begin", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("reqs.len=%d", len(reqs)))
		err := execBot.ActionTree.Grow(req) // or Grows
		log.Info(execBot.ObjId, "action-tree.grow=end", fmt.Sprintf("account-id=%s bot-id=%s", execBot.AccountId, execBot.BotId), fmt.Sprintf("action-tree.head-act-id=%s err=%s", execBot.ActionTree.HeadActId, err))
		if err != nil && execBot.ActionTree.HeadActNode() == nil {
			execBot.Terminate(header.BotTerminated_error.String())
			return
		}
	}

	if execBot.Status == Exec_bot_state_new && execBot.onStart != nil {
		execBot.onStart(execBot)
	}

	if execBot.ActionTree.HeadActNode() == nil { // or HeadActNodes
		execBot.Terminate(header.BotTerminated_complete.String())
		return
	}

	// actions continue
	execBot.Status = Exec_bot_state_waiting
}

func (execBot *ExecBot) DebuggingMode(mode string, realtime header.PubsubClient) {
	if mode == Exec_bot_mode_debug {
		execBot.mode = mode
		execBot.realtime = realtime
	}
}

func (execBot *ExecBot) Terminate(statusNote string) {
	if execBot.Status == Exec_bot_state_terminated {
		return
	}
	if execBot.ActionTree == nil {
		log.Error(execBot.ObjId, "exec-bot.terminate", fmt.Sprintf("account-id=%s bot-id=%s status-note=%s", execBot.AccountId, execBot.BotId, statusNote), "err=NILLLLLLL")
	}

	execBot.Status = Exec_bot_state_terminated
	execBot.StatusNote = statusNote

	if execBot.mode == Exec_bot_mode_debug && execBot.realtime != nil {
		evt := &header.Event{
			AccountId: execBot.AccountId,
			Id:        idgen.NewEventID(),
			Created:   time.Now().UnixNano() / 1e6,
			Type:      header.RealtimeType_bot_debug_end.String(),
			Data: &header.Event_Data{BotRunResponse: &header.RunResponse{
				AccountId:    execBot.AccountId,
				BotId:        execBot.BotId,
				ObjectType:   execBot.ObjType,
				ObjectId:     execBot.ObjId,
				ExecBotId:    execBot.Id,
				ExecBotState: execBot.Status,
			}},
		}
		payload, _ := json.Marshal(evt)
		go execBot.realtime.Publish(
			context.Background(),
			&header.PublishMessage{
				AccountId: execBot.AccountId,
				Payload:   payload,
				Topics:    []string{header.RealtimeType_bot_debug_end.String() + ".account." + execBot.AccountId},
			})
	}
	if execBot.onTerminate != nil {
		execBot.onTerminate(execBot)
	}
}

func (execBot *ExecBot) BeforeAction(node *ActionNode) {
	delete(execBot.SchedActs, node.GetAction().GetId())

	if execBot.mode == Exec_bot_mode_debug && execBot.realtime != nil {
		evt := &header.Event{
			AccountId: execBot.AccountId,
			Id:        idgen.NewEventID(),
			Created:   time.Now().UnixNano() / 1e6,
			Type:      header.RealtimeType_bot_debug_begin_action.String(),
			Data: &header.Event_Data{BotRunResponse: &header.RunResponse{
				AccountId:    execBot.AccountId,
				BotId:        execBot.BotId,
				ObjectType:   execBot.ObjType,
				ObjectId:     execBot.ObjId,
				ExecBotId:    execBot.Id,
				ExecBotState: execBot.Status,
				ActionId:     node.ActionId,
				ActionState:  node.State,
			}},
		}
		payload, _ := json.Marshal(evt)
		go func(ctx context.Context, publishMess *header.PublishMessage) {
			_, err := execBot.realtime.Publish(ctx, publishMess)
			if err != nil {
				log.Error(err)
				return
			}
		}(
			context.Background(),
			&header.PublishMessage{
				AccountId: execBot.AccountId,
				Payload:   payload,
				Topics:    []string{header.RealtimeType_bot_debug_begin_action.String() + ".account." + execBot.AccountId},
			},
		)
	}
}

func (execBot *ExecBot) Suspend() *ExecBot {
	execBot.isSuspend = true
	execBot.ActionTree = nil
	execBot.RunTimes = nil
	return execBot
}

func (execBot *ExecBot) Resume(raw *ExecBot) error {
	execBot.isSuspend = false
	execBot.merge(raw)
	execBot.ActionTree.Reboot(execBot)
	return nil
}

func (execBot *ExecBot) merge(raw *ExecBot) {
	if raw.AccountId != execBot.AccountId ||
		raw.BotId != execBot.BotId ||
		raw.ObjId != execBot.ObjId ||
		raw.Id != execBot.Id {
		return
	}
	if raw.Bot != nil {
		execBot.Bot = raw.Bot
	}

	if execBot.Created == 0 {
		execBot.Created = raw.Created
	}
	if execBot.Updated == 0 {
		execBot.Updated = raw.Updated
	}
	if execBot.LastSeen == 0 {
		execBot.LastSeen = raw.LastSeen
	}

	if execBot.ActionTree == nil {
		execBot.ActionTree = raw.ActionTree
	}

	if execBot.Status == "" {
		execBot.Status = raw.Status
	}
	if execBot.SchedActs == nil {
		execBot.SchedActs = raw.SchedActs
	}
	mergeRunTimes(execBot.RunTimes, raw.RunTimes)
	mergeLeadPayload(execBot.LeadPayload, raw.LeadPayload)
	execBot.raw = raw
}

func mergeRunTimes(dst []int64, src []int64) []int64 {
	dstMap := make(map[int64]struct{}, len(dst))
	for _, number := range dst {
		dstMap[number] = struct{}{}
	}

	for _, number := range src {
		if _, has := dstMap[number]; !has {
			dst = append(dst, number)
		}
	}

	sort.Slice(dst, func(i, j int) bool { return dst[i] > dst[j] })

	return dst
}

func mergeLeadPayload(dst map[string]string, src map[string]string) {
	if dst == nil {
		dst = make(map[string]string)
	}
	for key, value := range src {
		if _, has := dst[key]; !has {
			dst[key] = value
		}
	}
}

const (
	Exec_bot_state_new                 = "new"
	Exec_bot_state_runnable            = "runnable" // optional
	Exec_bot_state_waiting             = "waiting"
	Exec_bot_state_waiting_event       = "waiting_event"
	Exec_bot_state_waiting_time        = "waiting_time"
	Exec_bot_state_waiting_lock        = "waiting_lock"
	Exec_bot_state_running             = "running" // optional
	Exec_bot_state_terminated          = "terminated"
	Exec_bot_state_terminated_force    = "terminated_force"
	Exec_bot_state_terminated_complete = "terminated_complete"
	Exec_bot_state_terminated_expire   = "terminated_expire"
	Exec_bot_state_terminated_self     = "terminated_self"
	Exec_bot_state_terminated_start    = "terminated_start"
	Exec_bot_state_terminated_error    = "terminated_error"

	Exec_bot_mode_debug = "debugging"
)

// ExecBot strategy/limitation: 1 time (cached, update), new every time
// Cut tree if strategy is 1 time

type ExecBotIndex struct {
	AccountId    string            `json:"account_id"`
	BotId        string            `json:"bot_id"`
	ObjType      string            `json:"obj_type"`
	ObjId        string            `json:"obj_id"`
	LeadId       string            `json:"lead_id"`
	LeadPayload  map[string]string `json:"lead_payload"` // TODO replace
	Id           string            `json:"id"`
	IndexIncre   int64             `json:"index_incre"`
	BeginTimeDay int64             `json:"begin_time_day"`
	BeginTime    int64             `json:"begin_time"`
	EndTime      int64             `json:"end_time"`
	Created      int64             `json:"created"`
	Updated      int64             `json:"updated"`
	ExecBot      *ExecBot          `json:"exec_bot"`
}

func NewExecBotID(accountId, botId, objId string) string {
	// return accountId + "." + botId + "." + objId
	return objId + "." + botId // better for sort
}

func ObjIdOfExecBotID(id string) string {
	arr := strings.Split(id, ".")
	if len(arr) != 3 {
		return ""
	}
	return arr[2]
}

type RunRequestQueue struct {
	head   int
	Length int
	values []*header.RunRequest
}

func NewQueue() *RunRequestQueue {
	return &RunRequestQueue{head: 0, Length: 0, values: make([]*header.RunRequest, PKG_SIZE)}
}

func (r *RunRequestQueue) Peek(size int) []*header.RunRequest {
	size = min(size, r.Length)
	arr := make([]*header.RunRequest, size)
	for i := 0; i < size; i++ {
		arr[i] = r.values[(r.head+i)%len(r.values)]
	}

	// dequeue
	r.head = r.head + size
	r.Length -= size

	return arr
}

func (r *RunRequestQueue) Enqueue(payload *header.RunRequest) {
	// extend ring
	if r.Length+1 > len(r.values) {
		// copy reseted value
		arr := make([]*header.RunRequest, r.Length+PKG_SIZE)
		for i := 0; i < r.Length; i++ {
			arr[i] = r.values[(r.head+i)%len(r.values)]
		}

		r.values = arr
		r.head = 0
	}
	// TODO narrow ring
	r.Length = r.Length + 1
	r.values[(r.head+r.Length-1)%len(r.values)] = payload
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const PKG_SIZE = 16 // half dozen

// TODO exec bot is alive
