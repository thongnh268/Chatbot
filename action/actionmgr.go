package action

import (
	"sync"

	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
)

type ActionNode interface {
	GetBot() *header.Bot
	GetAction() *header.Action // just action, not tree
	GetObject() (string, string, []*pb.KV)
	GetRunRequests() []*header.RunRequest
	GetRunRequestStack() []*header.RunRequest
	GetRunType() string
	Delay(durationMillisecond int64, from int64) bool
	AppendSchedActs(schedAt int64)
	CanJump(actionId string) bool
	IsRoot() bool
	ConvertLead(key, value string)
	GetInternalState() []byte
	SetInternalState([]byte)
}

type ActionMgr struct {
	*sync.Mutex

	msg   header.ConversationEventReaderClient
	user  header.UserMgrClient
	convo header.ConversationMgrClient
	event header.EventMgrClient
}

func NewActionMgr(
	msg header.ConversationEventReaderClient,
	user header.UserMgrClient,
	convo header.ConversationMgrClient,
	event header.EventMgrClient,
) *ActionMgr {
	return &ActionMgr{
		Mutex: &sync.Mutex{},
		msg:   msg,
		user:  user,
		convo: convo,
		event: event,
	}
}

// require
func (mgr *ActionMgr) Do(node ActionNode, req *header.RunRequest) ([]string, error) {
	actFunc := mgr.getActFunc(node.GetAction().GetType())
	arr, err := actFunc.Do(mgr, node, req)
	return arr, err
}

func (mgr *ActionMgr) Wait(node ActionNode) string {
	actFunc := mgr.getActFunc(node.GetAction().GetType())
	return actFunc.Wait(node)
}

func (mgr *ActionMgr) Triggers(node ActionNode) []string {
	actFunc := mgr.getActFunc(node.GetAction().GetType())
	return actFunc.Triggers(node)
}

func (mgr *ActionMgr) GetEvtCondTriggers(evt *header.Event) []string {
	triggers := make([]string, 0)
	triggers = append(triggers, Trigger_event_exist)

	for _, evtFunc := range Evt_func_arr {
		if evtFunc.GetType(evt) != "" {
			triggers = append(triggers, evtFunc.Id)
		}
	}

	return triggers
}

func (mgr *ActionMgr) getActFunc(actType string) *ActionFunc {
	actFunc := Act_func_map[actType]
	if actFunc == nil {
		// invalid action
		actFunc = Action_nil
	}

	return actFunc
}
