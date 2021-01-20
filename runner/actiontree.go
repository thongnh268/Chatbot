package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
)

type ActionNode struct {
	execBot  *ExecBot
	prev     *ActionNode
	ActionId string // required
	action   *header.Action
	Nexts    []*ActionNode
	ObjCtxs  []*pb.KV

	Events []*header.Event //

	// only root|exec-bot
	IndexIncre  int64
	ActionHash  string
	HeadActId   string
	RunRequests []*header.RunRequest
	_flatNodes  map[string]*ActionNode
	_justGrowed []*ActionNode

	_currentReq   *header.RunRequest
	PrevState     string
	State         string
	StateNote     string
	RunTimes      []*RunTime
	FirstSeen     int64
	FirstRequest  int64
	InternalState []byte
}

func (n *ActionNode) GetBot() *header.Bot {
	root := n.Root()
	return root.execBot.Bot
}

func (n *ActionNode) GetAction() *header.Action {
	if n == nil {
		return nil
	}

	// action map require action id
	if n.ActionId == "" {
		return n.action
	}
	root := n.Root()
	if root.execBot == nil || len(root.execBot.actionMap) == 0 {
		return n.action
	}
	if action, has := root.execBot.actionMap[n.ActionId]; has && action != nil {
		return action
	}

	return n.action
}

func (n *ActionNode) GetObject() (string, string, []*pb.KV) {
	root := n.Root()
	return root.execBot.ObjId, root.execBot.ObjType, root.ObjCtxs
}

// current is last req
func (n *ActionNode) GetRunRequest() *header.RunRequest {
	var req *header.RunRequest
	reqStack := n.GetRunRequestStack()
	if len(reqStack) > 0 {
		req = reqStack[len(reqStack)-1]
	}
	return req
}

func (n *ActionNode) GetRunRequests() []*header.RunRequest {
	reqArr := make([]*header.RunRequest, 0)
	root := n.Root()
	reqLen := len(root.RunRequests)
	for _, runTime := range n.RunTimes {
		// TODO ignore start bot req
		if runTime.ReqIndex <= 0 || runTime.ReqIndex >= reqLen {
			continue
		}
		reqArr = append(reqArr, root.RunRequests[runTime.ReqIndex])
	}

	return reqArr
}

func (n *ActionNode) GetRunRequestStack() []*header.RunRequest {
	runArr := make([]*RunTime, 0)
	iter := n
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return nil
		}
		if iter.RunTimes != nil && len(iter.RunTimes) > 0 {
			runTimes := iter.RunTimes
			// TODO optimize reverse
			for i, j := 0, len(runTimes)-1; i < j; i, j = i+1, j-1 {
				runTimes[i], runTimes[j] = runTimes[j], runTimes[i]
			}
			runArr = append(runArr, runTimes...)
		}
		if iter.prev == nil {
			break
		}
		iter = iter.prev
		deep++
	}

	reqArr := make([]*header.RunRequest, 0)
	root := n.Root()
	reqLen := len(root.RunRequests)
	for _, runTime := range runArr {
		if runTime.ReqIndex <= 0 || runTime.ReqIndex >= reqLen {
			continue
		}
		reqArr = append(reqArr, root.RunRequests[runTime.ReqIndex])
	}

	return reqArr
}

// TODO onfailure, onreset, onload
func (n *ActionNode) GetRunType() string {
	if n.PrevState == Node_state_new {
		return "onstart"
	}
	if n._currentReq != nil {
		return n._currentReq.GetBotRunType()
	}
	return ""
}

func (n *ActionNode) AppendSchedActs(schedAt int64) {
	root := n.Root()
	if root.execBot.SchedActs == nil {
		root.execBot.SchedActs = make(map[string]int64)
	}
	root.execBot.SchedActs[n.ActionId] = schedAt / 1e2
}

// begin-end: latency1+runtime+sleep+latency2
func (n *ActionNode) Delay(durationMillisecond int64, from int64) bool {
	if n.FirstRequest == 0 && from == 0 {
		return false
	}

	var schedAt int64
	if from != 0 {
		schedAt = from + durationMillisecond
	} else {
		latency := n.FirstSeen - n.FirstRequest
		schedAt = n.FirstRequest + durationMillisecond - 2*latency
	}

	now := time.Now().UnixNano() / 1e6
	if now/1e2 >= schedAt/1e2 {
		return false
	}

	root := n.Root()
	if root.execBot.SchedActs == nil {
		root.execBot.SchedActs = make(map[string]int64)
	}
	root.execBot.SchedActs[n.ActionId] = schedAt / 1e2

	return true
}

func (n *ActionNode) CanJump(actionId string) bool {
	var jumpNode *ActionNode
	root := n.Root()
	if root._flatNodes != nil {
		jumpNode = root._flatNodes[actionId]
	}
	if jumpNode == nil {
		jumpNode = n.Find(actionId)
	}
	if jumpNode == nil || !jumpNode.beHead() {
		return false
	}
	return true
}

func (n *ActionNode) ConvertLead(key, value string) {
	root := n.Root()
	execBot := root.execBot
	if execBot.LeadPayload == nil {
		execBot.LeadPayload = make(map[string]string)
	}
	if key == "" || value == "" {
		return
	}
	key = strings.ToLower(strings.TrimSpace(key))
	if !strings.Contains(key, "email") && !strings.Contains(key, "phone") {
		return
	}
	execBot.LeadPayload[key] = value
	if execBot.onConvertLead != nil {
		execBot.onConvertLead(execBot)
	}
}

func (n *ActionNode) GetInternalState() []byte {
	if n != nil {
		return n.InternalState
	}
	return nil
}

func (n *ActionNode) SetInternalState(internalState []byte) {
	n.InternalState = internalState
}

func (n *ActionNode) beHead() bool {
	if n.State == Node_state_new ||
		n.State == Node_state_waiting ||
		n.State == Node_state_ready {
		return true
	}

	if n.beRenew() {
		return true
	}

	return false
}

func (n *ActionNode) beRenew() bool {
	if n.State == Node_state_done {
		return true
	}

	if (n.IsRoot() || n.prev.State == Node_state_done) &&
		(n.State == Node_state_error || n.State == Node_state_closed) {
		return true
	}

	return false
}

func (n *ActionNode) IsRoot() bool {
	if n.prev == nil {
		return true
	}
	return false
}

func (n *ActionNode) GetExecBot() *ExecBot {
	return n.Root().execBot
}

func (n *ActionNode) SetExecBot(execBot *ExecBot) {
	root := n.Root()
	root.execBot = execBot
}

func (root *ActionNode) JustGrowed() bool {
	return len(root._justGrowed) > 0
}

func (root *ActionNode) Grow(inReq *header.RunRequest) error {
	execBot := root.execBot
	// TODO rm
	if execBot == nil {
		return errors.New("exec bot not exist")
	}
	if root.RunRequests == nil {
		root.RunRequests = []*header.RunRequest{nil}
	}
	inReqIndex := len(root.RunRequests)
	root.RunRequests = append(root.RunRequests, inReq)

	seenCreated := time.Now().UnixNano() / 1e6
	reqCreated := root.GetRunRequest().GetCreated()
	if reqCreated == 0 {
		reqCreated = seenCreated
	}

	headNode := root.HeadActNode()
	var headNodeErr error
	justGrowed := make([]*ActionNode, 0)
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return nil
		}
		widthNodes := make([]*ActionNode, 0)

		node := headNode
		// TODO1
		if node == nil {
			log.Warn("NILLLLLLL", fmt.Sprintf("account-id=%s bot-id=%s obj-id=%s", execBot.AccountId, execBot.BotId, execBot.ObjId))
			break
		}
		if node.beRenew() {
			node.renew()
		}

		node._currentReq = inReq
		node.PrevState = node.State
		if node.FirstRequest == 0 {
			node.FirstRequest = reqCreated
		}
		if node.FirstSeen == 0 {
			node.FirstSeen = seenCreated
		}
		runTime := &RunTime{ReqIndex: inReqIndex, PrevState: node.State}
		node.RunTimes = append(node.RunTimes, runTime)

		if node.State == Node_state_new {
			node.State = Node_state_waiting
		}

		if node.State == Node_state_waiting {
			log.Info(execBot.ObjId, "action-mgr.wait=begin", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", execBot.AccountId, execBot.BotId, node.GetAction().GetId()))
			newState := root.execBot.actionMgr.Wait(node)
			log.Info(execBot.ObjId, "action-mgr.wait=end", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", execBot.AccountId, execBot.BotId, node.GetAction().GetId()), fmt.Sprintf("new-state=%s", newState))
			if newState == Node_state_ready || newState == Node_state_waiting {
				node.State = newState
			} else {
				node.Close()
			}
		}

		if node.State == Node_state_ready {
			node.State = Node_state_running
			execBot.BeforeAction(node)
			runTime.DoBeginAt = time.Now().UnixNano() / 1e6
			log.Info(execBot.ObjId, "action-mgr.do=begin", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", execBot.AccountId, execBot.BotId, node.GetAction().GetId()), fmt.Sprintf("req.run-type=%s prev-state=%s", inReq.GetBotRunType(), runTime.PrevState))
			actIds, err := root.execBot.actionMgr.Do(node, inReq)
			log.Info(execBot.ObjId, "action-mgr.do=end", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", execBot.AccountId, execBot.BotId, node.GetAction().GetId()), fmt.Sprintf("act-ids=%#v err=%s sched=%#v", actIds, err, execBot.SchedActs))
			runTime.NextActionIds = actIds
			runTime.DoEndAt = time.Now().UnixNano() / 1e6

			if err != nil {
				headNodeErr = err
				node.State = Node_state_error
				node.StateNote = err.Error()
				for _, n := range node.Nexts {
					n.Close()
				}
				justGrowed = append(justGrowed, node)
			} else if len(actIds) == 1 && actIds[0] == node.ActionId {
				node.State = Node_state_waiting
			} else {
				node.State = Node_state_done
				for _, id := range actIds {
					if _, has := root._flatNodes[id]; has {
						widthNodes = append(widthNodes, root._flatNodes[id])
					}
				}
				justGrowed = append(justGrowed, node)
			}
		}

		runTime.NextState = node.State
		runTime.NextStateNote = node.StateNote

		if node.State == Node_state_waiting {
			root.HeadActId = node.ActionId
		} else {
			root.HeadActId = "" // end
		}

		if len(widthNodes) == 0 {
			break
		}
		headNode = widthNodes[0] // first match
		headNodeErr = nil
		root.HeadActId = headNode.GetAction().GetId()
		deep++
	}
	root._justGrowed = justGrowed

	return headNodeErr
}

// Share data between actions by wait events and other services
// Do action belongs to channel if exist, else do default
func (root *ActionNode) Grows(inReq *header.RunRequest) error {
	execBot := root.execBot
	// TODO rm
	if execBot == nil {
		return errors.New("exec bot not exist")
	}
	if root.RunRequests == nil {
		root.RunRequests = []*header.RunRequest{nil}
	}
	inReqIndex := len(root.RunRequests)
	root.RunRequests = append(root.RunRequests, inReq)

	seenCreated := time.Now().UnixNano() / 1e6
	reqCreated := root.GetRunRequest().GetCreated()
	if reqCreated == 0 {
		reqCreated = seenCreated
	}

	headNodes := root.HeadActNodes()
	justGrowed := make([]*ActionNode, 0)
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return nil
		}
		widthNodes := make([]*ActionNode, 0)
		for _, node := range headNodes {
			if node.beRenew() {
				node.renew()
			}

			node._currentReq = inReq
			node.PrevState = node.State
			if node.FirstRequest == 0 {
				node.FirstRequest = reqCreated
			}
			if node.FirstSeen == 0 {
				node.FirstSeen = seenCreated
			}
			runTime := &RunTime{ReqIndex: inReqIndex, PrevState: node.State}
			node.RunTimes = append(node.RunTimes, runTime)

			if node.State == Node_state_new {
				node.State = Node_state_waiting
			}

			if node.State == Node_state_waiting {
				newState := root.execBot.actionMgr.Wait(node)
				if newState == Node_state_ready || newState == Node_state_waiting {
					node.State = newState
				} else {
					node.Close()
				}
			}

			if node.State == Node_state_ready {
				node.State = Node_state_running
				execBot.BeforeAction(node)
				runTime.DoBeginAt = time.Now().UnixNano() / 1e6
				actIds, err := root.execBot.actionMgr.Do(node, inReq)
				runTime.NextActionIds = actIds
				runTime.DoEndAt = time.Now().UnixNano() / 1e6

				if err != nil {
					node.State = Node_state_error
					node.StateNote = err.Error()
					for _, n := range node.Nexts {
						n.Close()
					}
					justGrowed = append(justGrowed, node)
				} else if len(actIds) == 1 && actIds[0] == node.ActionId {
					node.State = Node_state_waiting
				} else {
					node.State = Node_state_done
					for _, id := range actIds {
						if _, has := root._flatNodes[id]; has {
							widthNodes = append(widthNodes, root._flatNodes[id])
						}
					}
					justGrowed = append(justGrowed, node)
				}
			}
			runTime.NextState = node.State
			runTime.NextStateNote = node.StateNote
		}
		if len(widthNodes) == 0 {
			break
		}
		headNodes = widthNodes
		deep++
	}
	root._justGrowed = justGrowed

	return nil
}

func MakeTree(action *header.Action) (*ActionNode, error) {
	if action == nil {
		return nil, errors.New("action is nil")
	}

	actionMap := make(map[*header.Action]struct{})
	actionIdMap := make(map[string]struct{})
	deep := 0
	nodeIndex := 0
	if action.GetId() == "" {
		action.Id = strconv.Itoa(nodeIndex)
		nodeIndex++
	}
	root := NewActionNode(action)
	root.HeadActId = action.GetId()
	flatActions := map[string]*header.Action{}
	flatActions[action.GetId()] = action
	root._flatNodes = map[string]*ActionNode{}
	nodes := []*ActionNode{root}
	root._flatNodes[root.ActionId] = root
	var flatAction *header.Action
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return nil, errors.New("too deep")
		}

		widthNodes := make([]*ActionNode, 0)
		for _, node := range nodes {
			// TODO depend on wait of action type (not guarantee)
			if _, has := actionMap[node.action]; has {
				return nil, errors.New("loop")
			}
			if _, has := actionIdMap[node.ActionId]; has {
				return nil, errors.New("duplicated id")
			}

			actionMap[node.action] = struct{}{}
			actionIdMap[node.ActionId] = struct{}{}

			flatAction = flatActions[node.ActionId]
			if flatAction == nil || len(flatAction.GetNexts()) == 0 {
				continue
			}

			node.Nexts = make([]*ActionNode, 0)
			for _, actNext := range flatAction.GetNexts() {
				if actNext.GetAction() == nil {
					continue
				}
				if actNext.GetAction().GetId() == "" {
					actNext.Action.Id = strconv.Itoa(nodeIndex) // TODO1 duplicate
					nodeIndex++
				}
				flatActions[actNext.GetAction().GetId()] = actNext.GetAction()
				nextNode := NewActionNode(actNext.GetAction())
				nextNode.prev = node
				node.Nexts = append(node.Nexts, nextNode)
			}

			widthNodes = append(widthNodes, node.Nexts...)
		}

		if len(widthNodes) == 0 {
			break
		}
		for _, n := range widthNodes {
			root._flatNodes[n.ActionId] = n
		}

		nodes = widthNodes
		deep++
	}

	return root, nil
}

func (root *ActionNode) Reboot(execBot *ExecBot) {
	root.execBot = execBot
	actionMap := actionToMap(execBot.Bot.GetAction())
	root._flatNodes = map[string]*ActionNode{}
	nodes := []*ActionNode{root}
	root._flatNodes[root.ActionId] = root
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return
		}
		widthNodes := make([]*ActionNode, 0)
		for _, node := range nodes {
			if _, has := actionMap[node.ActionId]; !has {
				log.Warn("NILLLLLLL", fmt.Sprintf("account-id=%s bot-id=%s obj-id=%s", execBot.AccountId, execBot.BotId, execBot.ObjId))
			}
			node.action = actionMap[node.ActionId]
			if node.Nexts == nil || len(node.Nexts) == 0 {
				continue
			}
			for _, n := range node.Nexts {
				n.prev = node
			}
			widthNodes = append(widthNodes, node.Nexts...)
			root._flatNodes[node.ActionId] = node
		}

		if len(widthNodes) == 0 {
			break
		}
		nodes = widthNodes
		deep++
	}
}

func actionToMap(rootAction *header.Action) map[string]*header.Action {
	actionMap := make(map[string]*header.Action)
	actions := []*header.Action{rootAction}
	actionMap[rootAction.GetId()] = rootAction
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return actionMap
		}
		widthActions := make([]*header.Action, 0)
		for _, action := range actions {
			if action.GetId() != "" {
				actionMap[action.GetId()] = action
			}
			for _, nxt := range action.GetNexts() {
				widthActions = append(widthActions, nxt.GetAction())
			}
		}
		if len(widthActions) == 0 {
			break
		}
		actions = widthActions
		deep++
	}
	return actionMap
}

// Only save pure action (exclude act.Nexts)
// Proto clone too deep
// Require update if header.Action change
func NewActionNode(action *header.Action) *ActionNode {
	return &ActionNode{
		ActionId: action.GetId(),
		action:   action,
		State:    Node_state_new,
		RunTimes: make([]*RunTime, 0),
	}
}

func (n *ActionNode) renew() {
	n.PrevState = ""
	n.State = Exec_bot_state_new
	n.StateNote = ""
	n.RunTimes = make([]*RunTime, 0)
	n.FirstSeen = 0
	n.FirstRequest = 0
}

func (root *ActionNode) HeadActNode() *ActionNode {
	if root == nil {
		return nil
	}
	return root.Find(root.HeadActId)
}

func (root *ActionNode) HeadActNodes() []*ActionNode {
	head := make([]*ActionNode, 0)
	nodes := []*ActionNode{root}
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return nil
		}
		widthNodes := make([]*ActionNode, 0)
		for _, node := range nodes {
			if node.State == Node_state_new ||
				node.State == Node_state_waiting ||
				node.State == Node_state_ready {
				head = append(head, node)
			}
			if node.State == Node_state_done {
				widthNodes = append(widthNodes, node.Nexts...)
			}
		}
		if len(widthNodes) == 0 {
			break
		}
		nodes = widthNodes
		deep++
	}
	return head
}

func (n *ActionNode) Close() {
	nodes := []*ActionNode{n}
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return
		}
		widthNodes := make([]*ActionNode, 0)
		for _, node := range nodes {
			node.State = Node_state_closed
			widthNodes = append(widthNodes, node.Nexts...)
		}
		if len(widthNodes) == 0 {
			break
		}
		nodes = widthNodes
		deep++
	}
}

func (n *ActionNode) Root() *ActionNode {
	root := n
	if root.prev == nil {
		return root
	}

	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return nil
		}
		root = root.prev
		if root.prev == nil {
			return root
		}
		deep++
	}
}

func (n *ActionNode) Find(actionId string) *ActionNode {
	nodeMap := make(map[string]struct{})
	root := n.Root()
	nodes := root.trunkNodes()
	nodes = append(nodes, root)
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return nil
		}
		widthNodes := make([]*ActionNode, 0)
		for _, node := range nodes {
			if _, has := nodeMap[node.ActionId]; has {
				continue
			}
			if node.ActionId == actionId {
				return node
			}
			nodeMap[node.ActionId] = struct{}{}
			if node.prev != nil {
				widthNodes = append(widthNodes, node.prev)
			}
			if node.Nexts != nil && len(node.Nexts) > 0 {
				widthNodes = append(widthNodes, node.Nexts...)
			}
		}
		if len(widthNodes) == 0 {
			return nil
		}
		nodes = widthNodes
		deep++
	}
}

func (root *ActionNode) trunkNodes() []*ActionNode {
	trunk := make([]*ActionNode, 0)
	nodes := []*ActionNode{root}
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return nil
		}
		widthNodes := make([]*ActionNode, 0)
		for _, node := range nodes {
			if node.State == Node_state_done {
				trunk = append(trunk, node)
				widthNodes = append(widthNodes, node.Nexts...)
			}
		}
		if len(widthNodes) == 0 {
			break
		}
		nodes = widthNodes
		deep++
	}
	return trunk
}

// TODO use event map on root
// TODO save event id to node.events in grow
func (n *ActionNode) GetEvents() []*header.Event {
	return n.Events
}

// TODO before, after events
func (n *ActionNode) EventStack() []*header.Event {
	evtArr := make([]*header.Event, 0)
	iter := n
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return nil
		}
		if iter.Events != nil && len(iter.Events) > 0 {
			evtArr = append(evtArr, iter.Events...)
		}
		if iter.prev == nil {
			break
		}
		iter = iter.prev
		deep++
	}
	return evtArr
}

// TODO use in grows, disable need? too complicate
func (root *ActionNode) branchTrigger(nodeTriggerMap map[string]map[string]struct{}) {
	nodes := []*ActionNode{root}
	objType := root.execBot.Bot.GetCategory()
	objId := root.execBot.ObjId
	deep := 0
	for {
		if deep >= Bizbot_limit_act_deep {
			log.Warn("TOODEEPPP")
			return
		}
		widthNodes := make([]*ActionNode, 0)
		for _, node := range nodes {
			if node.State == Node_state_done {
				widthNodes = append(widthNodes, node.Nexts...)
			}

			nodeId := node.ActionId
			if _, has := nodeTriggerMap[nodeId]; has {
				continue
			}
			var parentTrigger map[string]struct{}
			var evtTrigger map[string]struct{}

			if node.prev != nil {
				parentTrigger = nodeTriggerMap[node.prev.ActionId]
			}
			if node.Events != nil && len(node.Events) > 0 {
				evtTrigger = make(map[string]struct{})
				for _, evt := range node.Events {
					triggerArr := TriggerOnEvent(objType, objId, evt, root.execBot.actionMgr)
					for _, trigger := range triggerArr {
						evtTrigger[trigger] = struct{}{}
					}
				}
			}

			length := 0
			if parentTrigger != nil {
				length += len(parentTrigger)
			}
			if evtTrigger != nil {
				length += len(evtTrigger)
			}
			nodeTriggerMap[nodeId] = make(map[string]struct{}, length)
			if parentTrigger != nil && len(parentTrigger) > 0 {
				for trigger := range parentTrigger {
					nodeTriggerMap[nodeId][trigger] = struct{}{}
				}
			}
			if evtTrigger != nil && len(evtTrigger) > 0 {
				for trigger := range evtTrigger {
					nodeTriggerMap[nodeId][trigger] = struct{}{}
				}
			}
		}
		if len(widthNodes) == 0 {
			break
		}
		nodes = widthNodes
		deep++
	}
}

// TODO use in grows, disable need? too complicate
func appendEvents(objType, objId string, nodes []*ActionNode, evt *header.Event, actionMgr ActionMgr) {
	evtTriggerArr := TriggerOnEvent(objType, objId, evt, actionMgr)
	evtTriggerMap := make(map[string]struct{}, len(evtTriggerArr))
	for _, id := range evtTriggerArr {
		evtTriggerMap[id] = struct{}{}
	}

	for _, node := range nodes {
		triggerArr := TriggerOfActionNode(node, actionMgr)
		found := false
		for _, trigger := range triggerArr {
			if _, has := evtTriggerMap[trigger]; has {
				found = true
				break
			}
		}
		if found {
			if node.Events == nil {
				node.Events = []*header.Event{evt}
			} else {
				node.Events = append(node.Events, evt)
			}
		}
	}
}

const (
	Bizbot_limit_act_deep = 128
	Node_state_new        = "new"
	Node_state_closed     = "closed"
	Node_state_waiting    = "waiting"
	Node_state_ready      = "ready"
	Node_state_running    = "running"
	Node_state_error      = "error"
	Node_state_done       = "done"
)

type RunTime struct {
	DoBeginAt     int64
	DoEndAt       int64
	ReqIndex      int
	PrevState     string
	NextState     string
	NextStateNote string
	NextActionIds []string
}

func (run *RunTime) String() string {
	out, _ := json.Marshal(run)

	return string(out)
}

// TODO flat replace for tree structure
// TODO set timeout for actions
