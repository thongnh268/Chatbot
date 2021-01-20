package action

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/subiz/goutils/log"
	"github.com/subiz/goutils/loop"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"github.com/subiz/perm"
	ggrpc "github.com/subiz/sgrpc"
)

var Http_method = map[string]struct{}{
	"GET":    {},
	"POST":   {},
	"PUT":    {},
	"PATCH":  {},
	"DELETE": {},
}

var Action_send_http = &ActionFunc{
	Id: header.ActionType_send_http.String(),
	Do: func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		bot, action := node.GetBot(), node.GetAction()
		actionId, sendHttp := action.GetId(), node.GetAction().GetSendHttp()
		accountId, botId := bot.GetAccountId(), bot.GetId()
		url, quilldelta, method := sendHttp.GetUrl(), sendHttp.GetQuillDelta(), sendHttp.GetMethod()
		headers := sendHttp.GetHeader()
		if _, has := Http_method[method]; !has {
			method = "POST"
		}
		var payload string
		if method == "POST" || method == "PUT" || method == "PATCH" {
			_, _, kvs := node.GetObject()
			var userId string
			for _, kv := range kvs {
				if kv.GetKey() == "user_id" {
					userId = kv.GetValue()
					break
				}
			}
			bot := node.GetBot()
			userAttr := &UserAttr{
				Mgr:       mgr,
				AccountId: bot.GetAccountId(),
				BotId:     bot.GetId(),
				UserId:    userId,
			}
			formatTexts := unmarshalQuillDelta(quilldelta)
			var text string
			if len(formatTexts) > 0 {
				for _, format := range formatTexts {
					if format.Key == "user" {
						jsonAttrb, _ := json.Marshal(userAttr.GetTextMap())
						text += string(jsonAttrb)
						continue
					}
					if format.Key != "" && userAttr.GetAttrText(format.Key) != "" {
						jsonAttrb, _ := json.Marshal(userAttr.GetAttrText(format.Key))
						text += string(jsonAttrb)
					} else {
						text += format.Text
					}
				}
			}
			payload = text
		}
		log.Info("act-"+actionId, "send-http=begin", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, actionId))
		err := sendHttpResquest(method, url, payload, headers)
		log.Info("act-"+actionId, "send-http=end", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, actionId), fmt.Sprintf("err=%s", err))
		if err != nil {
			return nil, err // TODO ignore?
		}
		return FirstMatchCondition(node, nil), nil
	},
}

func sendHttpResquest(method, url, text string, headers []*pb.KV) error {
	reader := strings.NewReader(text)
	req, reqerr := http.NewRequest(method, url, reader)
	if reqerr != nil {
		return reqerr
	}
	for _, header := range headers {
		req.Header.Add(header.GetKey(), header.GetValue())
	}
	client := &http.Client{Timeout: 30 * time.Second}
	var err error
	var index int
	var resp *http.Response
	loop.LoopErr(func() error {
		if index > 3 {
			return nil
		}
		resp, err = client.Do(req)
		index++
		if err == nil {
			return nil
		}
		defer resp.Body.Close()
		return err
	})
	return err
}

var Action_assign = &ActionFunc{
	Id: header.ActionType_assign.String(),
	Do: func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		bot, action := node.GetBot(), node.GetAction()
		accountId, botId := bot.GetAccountId(), bot.GetId()
		actionId, assign := action.GetId(), action.GetAssign()

		var isAgentReply bool
		for _, req := range node.GetRunRequests() {
			if req.GetEvent().GetType() == header.RealtimeType_message_sent.String() &&
				req.GetEvent().GetBy().GetType() == pb.Type_agent.String() {
				isAgentReply = true
			}
		}
		if isAgentReply {
			return FirstMatchCondition(node, []*InCondition{{Key: "agent_reply", Type: "boolean", Value: true}}), nil
		}

		if node.GetRunType() == "onstart" {
			conversationId, _, kvs := node.GetObject()
			var userId string
			for _, kv := range kvs {
				if kv.GetKey() == "user_id" {
					userId = kv.GetValue()
					break
				}
			}
			log.Info(conversationId, "assign-rule", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, actionId), fmt.Sprintf("user-id=%s user-rule=%t strategy=%s to=%#v agentable-only=%t", userId, assign.GetUseRule(), assign.GetStrategy(), assign.GetAssignTos(), assign.GetAvailableAgentsOnly()))
			err := assignRule(mgr, accountId, botId, conversationId, userId, assign)
			if err != nil {
				return nil, err
			}
			assignAt := time.Now().UnixNano() / 1e6
			stateb, _ := json.Marshal(assignAt)
			node.SetInternalState(stateb)
		}

		var assignAt int64
		if node.GetInternalState() != nil {
			json.Unmarshal(node.GetInternalState(), &assignAt)
		}
		if assign.GetAgentReplyTimeout() > 0 && node.Delay(assign.GetAgentReplyTimeout()*1e3, assignAt) {
			return []string{node.GetAction().GetId()}, nil
		}

		return FirstMatchCondition(node, []*InCondition{{Key: "agent_reply", Type: "boolean", Value: false}}), nil
	},
}

func assignRule(mgr *ActionMgr, accountId, botId, conversationId, userId string, assign *header.AssignRequest) error {
	allperm := perm.MakeBase()
	cred := ggrpc.ToGrpcCtx(&pb.Context{Credential: &pb.Credential{AccountId: accountId, Issuer: botId, Type: pb.Type_agent, Perm: &allperm}})
	assignReq := &header.AssignRequest{
		AccountId:           accountId,
		ConversationId:      conversationId,
		UserId:              userId,
		UseRule:             assign.GetUseRule(),
		Strategy:            assign.GetStrategy(),
		AssignTos:           assign.GetAssignTos(),
		AvailableAgentsOnly: assign.GetAvailableAgentsOnly(),
	}

	var err error
	var index int
	loop.LoopErr(func() error {
		if index > 3 {
			return nil
		}
		_, err = mgr.convo.AssignRule(cred, assignReq)
		index++
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

var Action_update_conversation = &ActionFunc{
	Id: header.ActionType_update_conversation.String(),
	Do: func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		bot := node.GetBot()
		accid, botid := bot.GetAccountId(), bot.GetId()
		getUpdateConvo := node.GetAction().GetUpdateConversation()
		actionId := node.GetAction().GetId()
		allperm := perm.MakeBase()
		convoid, _, _ := node.GetObject()
		cred := ggrpc.ToGrpcCtx(&pb.Context{
			Credential: &pb.Credential{
				AccountId: accid,
				Issuer:    botid,
				Type:      pb.Type_agent,
				Perm:      &allperm,
			},
		})
		if getUpdateConvo.TagIds != nil {
			log.Info("act-"+actionId, "tag-conversation", fmt.Sprintf("account-id=%s bot-id=%s convo-id=%s tag-id=%#v", accid, botid, convoid, getUpdateConvo.GetTagIds()))
			for _, tagid := range getUpdateConvo.GetTagIds() {
				for i := 0; i < 3; i++ {
					_, err := mgr.convo.TagConversation(cred, &header.TagRequest{
						AccountId:      accid,
						ConversationId: convoid,
						Id:             tagid,
					})
					if err == nil {
						break
					}
				}
			}
		}
		if getUpdateConvo.UntagIds != nil {
			log.Info("act-"+actionId, "untag-conversation", fmt.Sprintf("account-id=%s bot-id=%s convo-id=%s untag-id=%#v", accid, botid, convoid, getUpdateConvo.GetUntagIds()))
			for _, untagId := range getUpdateConvo.GetUntagIds() {
				for i := 0; i < 3; i++ {
					_, err := mgr.convo.UntagConversation(cred, &header.TagRequest{
						AccountId:      accid,
						ConversationId: convoid,
						Id:             untagId,
					})
					if err == nil {
						break
					}
				}
			}
		}
		if getUpdateConvo.GetEndConversation() {
			return []string{}, nil
		}
		return FirstMatchCondition(node, nil), nil
	},
}

var Action_jump = &ActionFunc{
	Id: header.ActionType_jump.String(),
	Do: func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		actionId := node.GetAction().GetJump().GetActionId()
		if actionId == "" {
			return Str_empty_arr, nil
		}

		// if node.CanJump(actionId) {
		// 	return Str_empty_arr, nil
		// }
		return []string{actionId}, nil
	},
	Wait: func(node ActionNode) string {
		return Node_state_ready
	},
	Triggers: func(node ActionNode) []string {
		return []string{Trigger_event_exist}
	},
}

// Only run one time (E.g. origin + jump)
var Action_sleep = &ActionFunc{
	Id: header.ActionType_sleep.String(),
	Do: func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		delay := node.Delay(node.GetAction().GetSleep().GetDuration()*1e3, int64(0))
		if delay {
			return []string{node.GetAction().GetId()}, nil
		}

		nextActions := node.GetAction().GetNexts()
		actionIdArr := make([]string, len(nextActions))

		for i, nextAct := range nextActions {
			actionIdArr[i] = nextAct.GetAction().GetId()
		}

		return actionIdArr, nil
	},
	Wait: func(node ActionNode) string {
		return Node_state_ready
	},
	Triggers: func(node ActionNode) []string {
		return []string{Trigger_event_exist}
	},
}

// ignore if action type not found
var Action_nil = &ActionFunc{
	Id: header.ActionType_nil.String(),
	Do: func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		nextActions := node.GetAction().GetNexts()
		actionIdArr := make([]string, len(nextActions))

		for i, nextAct := range nextActions {
			actionIdArr[i] = nextAct.GetAction().GetId()
		}
		return actionIdArr, nil
	},
	Wait: func(node ActionNode) string {
		return Node_state_ready
	},
	Triggers: func(node ActionNode) []string {
		return []string{}
	},
}

type ActionFunc struct {
	Id       string                                                                          // require
	Do       func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) // require
	Wait     func(node ActionNode) string
	Triggers func(node ActionNode) []string // belongs to wait
}

// action channels: subiz chat, zalo, messenger, page comment
// action groups: blocks, update data, conversation
// action types: E.g condition, sleep, ...
var Act_func_arr []*ActionFunc
var Act_func_map map[string]*ActionFunc

// Other structure?
func init() {
	testingActs := []*ActionFunc{
		Action_send_email,
		Action_convert_to_ticket,
		Action_send_webhook,
		Action_update_user,
		Action_send_message, // replace
		Action_question,     // replace
		Action_ask_question, // replace
		Action_condition,    // replace
	}
	stableActs := []*ActionFunc{
		Action_update_conversation,
		Action_send_http,
		Action_ask_question_sample,
		Action_assign,
		Action_jump,
		Action_sleep,
	}

	// read-only
	Act_func_arr = []*ActionFunc{}
	Act_func_arr = append(Act_func_arr, testingActs...)
	Act_func_arr = append(Act_func_arr, stableActs...) // priority 1
	Act_func_arr = append(Act_func_arr, Action_nil)    // priority 0
	for _, act := range Act_func_arr {
		if act.Id == "" {
			continue
		}
		if act.Do == nil {
			act.Do = Do(act.Id)
		}
		if act.Wait == nil {
			act.Wait = Wait(act.Id)
		}
		if act.Triggers == nil {
			act.Triggers = Triggers(act.Id)
		}
	}
	Act_func_map = make(map[string]*ActionFunc)
	for _, act := range Act_func_arr {
		if act.Id == "" {
			continue
		}
		if _, has := Act_func_map[act.Id]; has {
			log.Warn("act func is duplicated", act.Id)
		}
		Act_func_map[act.Id] = act
	}
}

// Can return default from state
func Triggers(id string) func(node ActionNode) []string {
	return func(node ActionNode) []string {
		return []string{}
	}
}

func Do(id string) func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
	return func(mgr *ActionMgr, node ActionNode, req *header.RunRequest) ([]string, error) {
		// if wait, return self action
		// if conditional, return valid actions
		// else return next actions of action
		// optional, update triggers

		// return error not from ActionNode mean give up then, closed branch
		return Str_empty_arr, nil
	}
}

// optional
func Wait(id string) func(node ActionNode) string {
	return func(node ActionNode) string {
		return Node_state_ready
	}
}

// Action state: new, ready, running, waiting, end
// Action state base on wait events: empty, notempty, full
// Action wait in do: delay & ability to do infinity

// TODO remove with wait and trigger func
const (
	Node_state_closed   = "closed"
	Node_state_waiting  = "waiting"
	Node_state_ready    = "ready"
	Trigger_event_exist = "event_exist"
)

const (
	Bizbot_limit_cond_level = 8
)

const (
	Bizbot_channel_zalo     = "zalo"
	Bizbot_channel_facebook = "facebook"
)

var Str_empty_arr = []string{}
