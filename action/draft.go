package action

import (
	"encoding/json"
	"regexp"
	"sort"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"github.com/subiz/perm"
	ggrpc "github.com/subiz/sgrpc"
)

var Action_send_email = &ActionFunc{Id: header.ActionType_send_email.String()}
var Action_convert_to_ticket = &ActionFunc{Id: header.ActionType_convert_to_ticket.String()}
var Action_send_webhook = &ActionFunc{Id: header.ActionType_send_webhook.String()}
var Action_question = &ActionFunc{Id: header.ActionType_question.String()}
var Action_update_user = &ActionFunc{Id: header.ActionType_update_user.String()}

var Action_ask_question = &ActionFunc{
	Id: header.ActionType_ask_question.String(),
	Do: func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		accid := node.GetBot().GetAccountId()
		botid := node.GetBot().GetId()

		action := node.GetAction()
		getAskQuestion := action.GetAskQuestion()

		allperm := perm.MakeBase()
		var state AskQuestionState

		internalState := node.GetInternalState()
		state.Step = 0
		if internalState != nil {
			json.Unmarshal(internalState, &state)
		}

		var err error
		userid := ""
		convoid, _, kvs := node.GetObject()
		for _, kv := range kvs {
			if kv.GetKey() == "user_id" {
				userid = kv.GetValue()
			}
			break
		}

		// only true if attr is text
		if getAskQuestion.GetSkipIfAttributeAlreadyExisted() {
			var readUser *header.User
			// ignore error if end loop
			for i := 0; i < 3; i++ {
				readUser, err = readUserAttr(mgr, allperm, botid, accid, userid)
				if err == nil {
					break
				}
			}
			for _, attrs := range readUser.GetAttributes() {
				if attrs.GetKey() == getAskQuestion.GetSaveToAttribute() {
					str := []*InCondition{
						{
							Key:   "last_response",
							Type:  "string",
							Value: attrs.GetText(),
							Convs: map[string]func(*InCondition) (string, interface{}){
								"valid": func(data *InCondition) (string, interface{}) {
									return "boolean", true
								},
							},
						},
					} // TODO other type
					return FirstMatchCondition(node, str), nil
				}
			}
		}

		//send message first
		if state.Step < 1 {
			//send event conversation_typing
			for _, row := range getAskQuestion.GetMessages() {
				msg := proto.Clone(row).(*header.Message)
				msg.ConversationId = convoid
				msg.QuillDelta = ""
				_, _ = typingEvent(mgr, allperm, accid, botid, convoid)
				time.Sleep(time.Duration(TypingDuration(msg.GetText())) * time.Second)
				for i := 0; i < 3; i++ {
					_, err = sendMessage(mgr, allperm, msg, accid, botid)
					if err == nil {
						break
					}
					e, ok := err.(*errors.Error)
					if ok && e.Class == 400 {
						return FirstMatchCondition(node, nil), nil
					}
				}
				if err != nil {
					return FirstMatchCondition(node, nil), nil
				}
			}
			if !getAskQuestion.GetWaitForUserResponse() {
				return FirstMatchCondition(node, nil), nil
			}
			state.Step = 1
			state.AskRange = map[int][]int64{}
			state.AskRange[0] = []int64{time.Now().UnixNano() / 1e6, 0}
			state.AskIndex = 0
			stateb, _ := json.Marshal(state)
			node.SetInternalState(stateb)
			return []string{action.GetId()}, nil
		}

		reqs := make([]*header.RunRequest, 0)
		var askfrom, askto int64
		askfrom = state.AskRange[state.AskIndex][0]
		askto = state.AskRange[state.AskIndex][1]
		if askto == 0 {
			askto = 9999999999999 // upper bound
		}
		for _, req := range node.GetRunRequests() {
			if req.GetEvent().GetType() != header.RealtimeType_message_sent.String() || req.GetEvent().GetBy().GetType() != pb.Type_user.String() {
				continue
			}
			log.Debug(askfrom, askto, req.GetCreated())
			if req.GetCreated() >= askfrom && req.GetCreated() < askto {
				reqs = append(reqs, req)
			}
			// reqs = append(reqs, req)

		}
		if len(reqs) == 0 {
			return []string{action.GetId()}, nil
		}
		sort.Slice(reqs, func(i, j int) bool { return reqs[i].GetCreated() > reqs[i].GetCreated() })
		//create event conversation_typing
		if getAskQuestion.GetValidation() != "" && getAskQuestion.GetValidation() != "none" {
			_, _ = typingEvent(mgr, allperm, accid, botid, convoid)
		}
		delay := node.Delay(1*1e3, reqs[0].GetCreated())
		if delay {
			return []string{action.GetId()}, nil
		}

		evt := reqs[0].GetEvent()
		val := evt.GetData().GetMessage().GetText()
		var r *regexp.Regexp
		validationErrorMess := ""
		switch getAskQuestion.GetValidation() {
		case "email":
			r, _ = regexp.Compile(RegexEmail)
			if !r.MatchString(val) {
				validationErrorMess = "Oops! You may input wrong email. Please input your email again"
			}
		case "phone":
			r, _ = regexp.Compile(RegexNumber)
			if !r.MatchString(val) {
				validationErrorMess = "Oops! You may input wrong number. Please input your number again"
			}
		case "number":
			r, _ = regexp.Compile(RegexNumber)
			if !r.MatchString(val) {
				validationErrorMess = "Oops! You may input wrong number. Please input your number again"
			}
		case "date":
			r, _ = regexp.Compile(RegexDate)
			if !r.MatchString(val) {
				validationErrorMess = "Oops! You may input wrong date format. Please enter according to the format dd/mm/yyyy"
			}
		default:
			validationErrorMess = ""
		}
		if validationErrorMess != "" {
			if state.Step > 3 {
				str := []*InCondition{
					{
						Key:   "last_response",
						Type:  "string",
						Value: val,
						Convs: map[string]func(*InCondition) (string, interface{}){
							"valid": func(data *InCondition) (string, interface{}) {
								return "boolean", false
							},
						},
					},
				}
				return FirstMatchCondition(node, str), nil
			}
			if state.Step <= 3 {
				msg := &header.Message{}
				msg.ConversationId = convoid
				msg.QuillDelta = ""
				msg.Text = validationErrorMess
				for i := 0; i < 3; i++ {
					_, err = sendMessage(mgr, allperm, msg, accid, botid)
					if err == nil {
						break
					}
					e, ok := err.(*errors.Error)
					if ok && e.Class == 400 {
						return FirstMatchCondition(node, nil), nil
					}
				}
				if err != nil {
					return FirstMatchCondition(node, nil), nil
				}
				state.Step++
				state.AskIndex++
				now := time.Now().UnixNano() / 1e6
				state.AskRange[state.AskIndex] = []int64{now, 0}
				state.AskRange[state.AskIndex-1][1] = now
				stateb, _ := json.Marshal(state)
				node.SetInternalState(stateb)
				return []string{action.GetId()}, nil
			}
		}

		//Update user attributes
		user := &header.User{
			Id:        evt.GetBy().GetId(),
			AccountId: accid,
		}
		attr := &header.Attribute{}
		attr.Key = getAskQuestion.GetSaveToAttribute()
		attr.Text = val
		user.Attributes = append(user.Attributes, attr)
		for i := 0; i < 3; i++ {
			_, err = updateUser(mgr, allperm, user, accid, botid)
			if err == nil {
				break
			}
		}

		node.ConvertLead(attr.GetKey(), attr.GetText())

		state.Step = 0
		stateb, _ := json.Marshal(state)
		node.SetInternalState(stateb)

		str := []*InCondition{
			{
				Key:   "last_response",
				Type:  "string",
				Value: val,
				Convs: map[string]func(*InCondition) (string, interface{}){
					"valid": func(data *InCondition) (string, interface{}) {
						return "boolean", true
					},
				},
			},
		}
		return FirstMatchCondition(node, str), nil
	},
}

func sendMessage(mgr *ActionMgr, allperm pb.Permission, msg *header.Message, accid, botid string) (*header.Event, error) {
	return mgr.msg.SendMessage(ggrpc.ToGrpcCtx(&pb.Context{
		Credential: &pb.Credential{
			AccountId: accid,
			Issuer:    botid,
			Type:      pb.Type_agent,
			Perm:      &allperm,
		},
	}), &header.Event{
		AccountId: accid,
		Type:      header.RealtimeType_message_sent.String(),
		Data:      &header.Event_Data{Message: msg},
	})
}

func typingEvent(mgr *ActionMgr, allperm pb.Permission, accid, botid, convoid string) (*pb.Empty, error) {
	return mgr.convo.Typing(ggrpc.ToGrpcCtx(&pb.Context{
		Credential: &pb.Credential{
			AccountId: accid,
			Issuer:    botid,
			Type:      pb.Type_agent,
			Perm:      &allperm,
			ClientId:  "clisubizapi",
		},
	}), &pb.Id{
		AccountId: accid,
		Id:        convoid,
	})
}

func readUserAttr(mgr *ActionMgr, allperm pb.Permission, botid, accid, userid string) (*header.User, error) {
	return mgr.user.ReadUser(ggrpc.ToGrpcCtx(&pb.Context{
		Credential: &pb.Credential{
			AccountId: accid,
			Issuer:    botid,
			Type:      pb.Type_agent,
			Perm:      &allperm,
		},
	}), &pb.Id{
		Id: userid,
	})
}

func updateUser(mgr *ActionMgr, allperm pb.Permission, user *header.User, accid, botid string) (*pb.Id, error) {
	return mgr.user.UpdateUser(ggrpc.ToGrpcCtx(&pb.Context{
		Credential: &pb.Credential{
			AccountId: accid,
			Issuer:    botid,
			Type:      pb.Type_agent,
			Perm:      &allperm,
		},
	}), user)
}

var Action_send_message = &ActionFunc{
	Id: header.ActionType_send_message.String(),
	Do: func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		accid, botid := node.GetBot().GetAccountId(), node.GetBot().GetId()
		convoid, _, _ := node.GetObject()
		allperm := perm.MakeBase()
		var err error
		for _, msg := range node.GetAction().GetAskQuestion().GetMessages() {
			msg = proto.Clone(msg).(*header.Message)
			msg.ConversationId = convoid
			msg.QuillDelta = ""
			_, _ = typingEvent(mgr, allperm, accid, botid, convoid)
			time.Sleep(1000 * time.Millisecond)
			for i := 0; i < 3; i++ {
				_, err = sendMessage(mgr, allperm, msg, accid, botid)
				if err == nil {
					break
				}
				e, ok := err.(*errors.Error)
				if ok && e.Class == 400 {
					return FirstMatchCondition(node, nil), nil
				}
			}
			if err != nil {
				return FirstMatchCondition(node, nil), nil
			}

		}
		return FirstMatchCondition(node, nil), nil
	},
}

var Action_condition = &ActionFunc{
	Id: header.ActionType_condition.String(),
	Do: func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		nextActions := node.GetAction().GetNexts()
		if nextActions == nil || len(nextActions) == 0 {
			return Str_empty_arr, nil
		}

		evts := make([]*header.Event, 0)
		for _, req := range node.GetRunRequestStack() {
			if req.GetEvent() != nil {
				evts = append(evts, req.GetEvent())
			}
		}
		actionIdArr := make([]string, 0)
		for _, n := range nextActions {
			if n.GetCondition() == nil {
				continue
			}
			if doConditionalEvents(n.GetCondition(), evts) {
				actionIdArr = append(actionIdArr, n.GetAction().GetId())
			}
		}

		return actionIdArr, nil
	},
	Wait: func(node ActionNode) string {
		evts := make([]*header.Event, 0)
		for _, req := range node.GetRunRequestStack() {
			if req.GetEvent() != nil {
				evts = append(evts, req.GetEvent())
			}
		}
		actWaitState := Act_cond_empty
		for _, nextAct := range node.GetAction().GetNexts() {
			var waitState string
			waitState, _ = waitConditionalEvents(nextAct.GetCondition(), evts, Bizbot_limit_cond_level)
			if waitState == Act_cond_full {
				return Node_state_ready
			}

			if actWaitState == Act_cond_empty && waitState == Act_cond_notempty {
				actWaitState = waitState
			}
		}

		if actWaitState == Act_cond_empty {
			return Node_state_closed
		}

		// should wait
		return Node_state_waiting
	},
	Triggers: func(node ActionNode) []string {
		if node == nil {
			return Str_empty_arr
		}
		triggers := make([]string, 0)
		for _, nextAct := range node.GetAction().GetNexts() {
			triggers = append(triggers, getCondTriggers(nextAct.GetCondition())...)
		}

		return triggers
	},
}

// Default group is single
func getCondTriggers(condition *header.Condition) []string {
	ids := make([]string, 0)
	if condition.Group == "" || condition.Group == header.Condition_single.String() {
		ids = append(ids, condition.Key)
	}

	if condition.Group == header.Condition_all.String() || condition.Group == header.Condition_any.String() {
		for _, cond := range condition.Conditions {
			subIds := getCondTriggers(cond)
			ids = append(ids, subIds...)
		}
	}

	return ids
}

// ignore duplicated wait events
func waitConditionalEvents(cond *header.Condition, events []*header.Event, limitlevel int) (string, []string) {
	if limitlevel <= 0 {
		return Act_cond_full, []string{} // ignore too deep level
	}

	condState := Act_cond_empty

	if cond.Group == header.Condition_single.String() {
		evtFunc := Evt_func_map[cond.Key]

		if evtFunc == nil {
			return Act_cond_full, []string{} // valid with every event
		}
		// valid with first event (not latest event)
		for _, evt := range events {
			if evtFunc.GetType(evt) != "" {
				return Act_cond_full, []string{evt.GetId()}
			}
		}
	}

	evtChain := make([]string, 0)
	if cond.Group == header.Condition_any.String() {
		for _, c := range cond.Conditions {
			subCondState, subEvtChain := waitConditionalEvents(c, events, limitlevel-1)
			evtChain = append(evtChain, subEvtChain...)
			if subCondState == Act_cond_full {
				return Act_cond_full, evtChain
			}

			if condState == Act_cond_empty && subCondState != Act_cond_empty {
				condState = subCondState
			}
		}
	}

	if cond.Group == header.Condition_all.String() {
		allTrue := true
		for _, c := range cond.Conditions {
			subCondState, subEvtChain := waitConditionalEvents(c, events, limitlevel-1)
			evtChain = append(evtChain, subEvtChain...)
			if subCondState != Act_cond_full {
				allTrue = false
			}

			if condState == Act_cond_empty && subCondState != Act_cond_empty {
				condState = Act_cond_notempty
			}
		}

		if allTrue {
			return Act_cond_full, evtChain
		}
	}

	return condState, evtChain
}

func doConditionalEvents(cond *header.Condition, events []*header.Event) bool {
	if cond.Group == header.Condition_single.String() {
		// valid with first event (not latest event), ref func wait
		for _, evt := range events {
			if compareEvent(evt, cond) {
				return true
			}
		}
	}

	if cond.Group == header.Condition_any.String() {
		// valid with first cond
		for _, c := range cond.Conditions {
			subIs := doConditionalEvents(c, events)
			if subIs {
				return true
			}
		}
	}

	if cond.Group == header.Condition_all.String() {
		for _, c := range cond.Conditions {
			subIs := doConditionalEvents(c, events)
			if !subIs {
				return false
			}
		}
		return true
	}

	return false
}

func compareEvent(evt *header.Event, cond *header.Condition) bool {
	evtFunc := Evt_func_map[cond.Key]
	if evtFunc == nil {
		return false
	}

	is := false
	value := evtFunc.GetValue(evt)

	switch cond.Type {
	case "string":
		str := value.(string)
		is = compareStringParams(cond.GetString_(), str)
	case "number":
		num := value.(float32)
		is = compareNumberParams(cond.GetNumber(), num)
	case "boolean":
	}

	return is
}

// II. Conditional
// 1. just exist event require
// 2. conditional (single, any, all)

// State of action base on conditions of events
const (
	Act_cond_full     = "full"     // totally true
	Act_cond_empty    = "empty"    // false
	Act_cond_notempty = "notempty" // maybe true
)

func Condition(node ActionNode) []string {
	nextActions := node.GetAction().GetNexts()
	evts := make([]*header.Event, 0)
	for _, req := range node.GetRunRequestStack() {
		if req.GetEvent() != nil {
			evts = append(evts, req.GetEvent())
		}
	}
	actionIdArr := make([]string, 0)
	for _, n := range nextActions {
		if n.GetAction().GetId() == "" {
			continue
		}
		if n.GetCondition() == nil {
			actionIdArr = append(actionIdArr, n.GetAction().GetId())
			continue
		}
		if doConditionalEvents(n.GetCondition(), evts) {
			actionIdArr = append(actionIdArr, n.GetAction().GetId())
		}
	}

	return actionIdArr
}

func doConditionalNode(cond *header.Condition, node ActionNode, data []*InCondition) bool {
	if cond.Group == header.Condition_single.String() {
		dataFunc := transformFunc(cond.GetTransformFunction(), data)
		if compareNode(cond, node, dataFunc) {
			return true
		}
	}

	if cond.Group == header.Condition_any.String() {
		// valid with first cond
		for _, c := range cond.Conditions {
			subIs := doConditionalNode(c, node, data)
			if subIs {
				return true
			}
		}
	}

	if cond.Group == header.Condition_all.String() {
		for _, c := range cond.Conditions {
			subIs := doConditionalNode(c, node, data)
			if !subIs {
				return false
			}
		}
		return true
	}
	return false
}

func compareNode(cond *header.Condition, node ActionNode, data []*InCondition) bool {
	var value interface{}

	for _, row := range data {
		if row.Key == cond.Key {
			if row.Type != cond.Type {
				return false
			}
			value = row.Value
			break
		}
	}

	if value == nil && node != nil {
		reqs := make([]*header.RunRequest, 0)
		for _, row := range data {
			if row.Req != nil {
				reqs = append(reqs, row.Req)
			}
		}
		condFunc := Action_cond_func_map[cond.Key+"."+cond.Type]
		if condFunc == nil {
			return false
		}

		value = condFunc.GetValue(node, reqs)
	}

	if value == nil {
		return false
	}

	is := false
	switch cond.Type {
	case "string":
		str, ok := value.(string)
		is = ok && compareStringParams(cond.GetString_(), str)
	case "number":
		num, ok := value.(float32)
		is = ok && compareNumberParams(cond.GetNumber(), num)
	case "boolean":
		b, ok := value.(bool)
		is = ok && b == cond.GetBoolean().GetIs()
	}
	return is
}

func transformFunc(f string, src []*InCondition) []*InCondition {
	if f == "" {
		return src
	}
	dst := make([]*InCondition, len(src))
	for i, srcRow := range src {
		dst[i] = &InCondition{}
		// TODO
		dst[i].Type = srcRow.Type
		dst[i].Value = srcRow.Value
	}

	return src
}

type EventFunc struct {
	Id       string
	GetType  func(evt *header.Event) string
	GetValue func(evt *header.Event) interface{}
}

var Evt_func_arr []*EventFunc
var Evt_func_map map[string]*EventFunc

type ActionConditionFunc struct {
	Id       string
	Type     string
	GetValue func(node ActionNode, inReqs []*header.RunRequest) interface{}
}

var Action_cond_func_arr []*ActionConditionFunc
var Action_cond_func_map map[string]*ActionConditionFunc

func init() {
	Evt_func_arr = []*EventFunc{
		{
			Id: "created",
			GetType: func(evt *header.Event) string {
				if evt.Created > 0 {
					return "number"
				}
				return ""
			},
			GetValue: func(evt *header.Event) interface{} {
				return float32(evt.GetCreated()) // type number
			},
		},
		{
			Id: "data.content.url",
			GetType: func(evt *header.Event) string {
				if evt.Type == header.RealtimeType_content_viewed.String() {
					return "string"
				}
				return ""
			},
			GetValue: func(evt *header.Event) interface{} {
				return evt.GetData().GetContent().GetUrl()
			},
		},
	}

	Evt_func_map = make(map[string]*EventFunc)
	for _, evtFunc := range Evt_func_arr {
		Evt_func_map[evtFunc.Id] = evtFunc
	}

	Action_cond_func_arr = []*ActionConditionFunc{
		{
			Id:   "conversation_id",
			Type: "string",
			GetValue: func(node ActionNode, inReqs []*header.RunRequest) interface{} {
				objId, objType, _ := node.GetObject()
				if objType == header.BotCategory_conversations.String() {
					return objId
				}
				return ""
			},
		},
		{
			Id:   "last_response",
			Type: "string",
			GetValue: func(node ActionNode, inReqs []*header.RunRequest) interface{} {
				reqs := node.GetRunRequestStack()
				reqs = append(reqs, inReqs...)
				evts := make([]*header.Event, 0)
				for _, req := range reqs {
					if req.GetEvent() == nil {
						continue
					}
					if req.GetEvent().GetType() != header.RealtimeType_message_sent.String() {
						continue
					}
					if req.GetEvent().GetBy().GetType() != pb.Type_user.String() {
						continue
					}
					evts = append(evts, req.GetEvent())
				}

				lastEvt := &header.Event{}
				for _, evt := range evts {
					if evt.GetCreated() >= lastEvt.GetCreated() {
						lastEvt = evt
					}
				}

				return lastEvt.GetData().GetMessage().GetText()
			},
		},
	}

	Action_cond_func_map = make(map[string]*ActionConditionFunc)
	for _, condFunc := range Action_cond_func_arr {
		Action_cond_func_map[condFunc.Id+"."+condFunc.Type] = condFunc
	}
}
