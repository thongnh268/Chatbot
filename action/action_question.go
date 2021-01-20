package action

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	"github.com/subiz/goutils/log"
	"github.com/subiz/goutils/loop"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"github.com/subiz/perm"
	ggrpc "github.com/subiz/sgrpc"
)

type AskQuestionState struct {
	Step     int32
	AskRange map[int][]int64
	AskIndex int

	InvalidResponse map[string]struct{}
	ResumeAt        int64
	QuestionAt      int64 // TODO conflict messages
}

func TypingDuration(text string) int64 {
	length := len([]rune(text))
	duration := (int64)(length) / 10
	if duration > 3 {
		return 3
	}
	return duration
}

var Action_ask_question_sample = &ActionFunc{
	Id: header.ActionType_ask_question.String(),
	Do: func(mgr *ActionMgr, node ActionNode, inReq *header.RunRequest) ([]string, error) {
		bot, action := node.GetBot(), node.GetAction()
		actionId, askQuestion := action.GetId(), action.GetAskQuestion()
		accountId, botId := bot.GetAccountId(), bot.GetId()
		conversationId, _, kvs := node.GetObject()
		var userId string
		for _, kv := range kvs {
			if kv.GetKey() == "user_id" {
				userId = kv.GetValue()
				break
			}
		}

		if !askQuestion.GetWaitForUserResponse() {
			log.Info(conversationId, "send-message", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, actionId))
			err := sendMessages(mgr, accountId, botId, conversationId, userId, askQuestion.GetMessages(), node.IsRoot())
			if err != nil {
				return nil, err
			}
			return FirstMatchCondition(node, nil), nil
		}

		var state AskQuestionState
		if node.GetInternalState() != nil {
			json.Unmarshal(node.GetInternalState(), &state)
		}
		if state.InvalidResponse == nil {
			state.InvalidResponse = make(map[string]struct{})
		}
		runRequests := node.GetRunRequests() // require req|event created asc
		userResponses := make([]*header.RunRequest, 0)
		// TODO ignore start action req
		for i := 1; state.QuestionAt > 0 && i < len(runRequests); i++ {
			if runRequests[i].GetCreated() > state.QuestionAt && isUserResponse(runRequests[i]) {
				userResponses = append(userResponses, runRequests[i])
			}
		}
		var lastResponseId, lastResponse string
		var lastResponseAt int64
		var validLastResponse, invalidMessage string
		if len(userResponses) > 0 {
			lastReq := userResponses[len(userResponses)-1]
			lastResponseId, lastResponse = lastReq.GetEvent().GetId(), lastReq.GetEvent().GetData().GetMessage().GetText()
			lastResponseAt = lastReq.GetCreated()
			textResponses := make([]string, len(userResponses))
			for i := 0; i < len(userResponses); i++ {
				textResponses[i] = userResponses[i].GetEvent().GetData().GetMessage().GetText()
			}
			validLastResponse, invalidMessage = getValidLast(textResponses, askQuestion.GetValidation())
		}

		if validLastResponse != "" && userId != "" {
			user := &header.User{Id: userId, AccountId: accountId, Attributes: []*header.Attribute{{Key: askQuestion.GetSaveToAttribute(), Text: validLastResponse}}}
			log.Info(conversationId, "update-user", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, actionId), fmt.Sprintf("user-id=%s %s=%s", userId, askQuestion.GetSaveToAttribute(), validLastResponse))
			err := upsertUser(mgr, accountId, botId, user)
			if err != nil {
				return nil, err
			}

			node.ConvertLead(askQuestion.GetSaveToAttribute(), validLastResponse)
		}
		if validLastResponse == "" && askQuestion.GetSkipIfAttributeAlreadyExisted() && userId != "" {
			log.Info(conversationId, "read-user", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, actionId), fmt.Sprintf("user-id=%s", userId))
			user, err := readUser(mgr, accountId, botId, userId)
			if err != nil {
				return nil, err
			}
			textAttrs := make([]string, 0)
			for _, attr := range user.GetAttributes() {
				if attr.GetKey() == askQuestion.GetSaveToAttribute() && attr.GetText() != "" {
					textAttrs = append(textAttrs, attr.GetText())
				}
			}
			if len(textAttrs) > 0 {
				validLastResponse, invalidMessage = getValidLast(textAttrs, askQuestion.GetValidation())
			}
		}

		if validLastResponse != "" {
			return FirstMatchCondition(node, []*InCondition{{Key: "last_response", Type: "string", Value: validLastResponse, Convs: Condition_conv_valid}}), nil
		}
		if askQuestion.GetRetry() != 1 && (isFirstResponseRefuse(userResponses) || len(userResponses) > 1) {
			return FirstMatchCondition(node, []*InCondition{{Key: "last_response", Type: "string", Value: lastResponse, Convs: Condition_conv_invalid}}), nil
		}

		// 1 time with on start
		if node.GetRunType() == "onstart" {
			log.Info(conversationId, "send-message", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, actionId))
			err := sendMessages(mgr, accountId, botId, conversationId, userId, askQuestion.GetMessages(), node.IsRoot())
			if err != nil {
				return nil, err
			}
			nowms := time.Now().UnixNano() / 1e6
			state.ResumeAt = nowms
			state.QuestionAt = nowms
		}
		// 1 time with user reponse
		_, isSent := state.InvalidResponse[lastResponseId]
		if lastResponseAt > 0 && !node.Delay(1*1e3, lastResponseAt) && !isSent {
			msg := &header.Message{ConversationId: conversationId, Text: invalidMessage}
			log.Info(conversationId, "send-message", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, actionId))
			err := sendMessages(mgr, accountId, botId, conversationId, userId, []*header.Message{msg}, false)
			if err != nil {
				return nil, err
			}
			state.InvalidResponse[lastResponseId] = struct{}{}
			state.ResumeAt = time.Now().UnixNano() / 1e6
		}
		// 1 time with last sent E.g. ask, invalid response
		if askQuestion.GetUseResumeMessage() && askQuestion.GetResumeMessage() != nil && state.ResumeAt > 0 && !node.Delay(askQuestion.GetResumeInterval()*1e3, state.ResumeAt) {
			log.Info(conversationId, "send-message", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, actionId))
			err := sendMessages(mgr, accountId, botId, conversationId, userId, []*header.Message{askQuestion.GetResumeMessage()}, false)
			if err != nil {
				return nil, err
			}
			state.ResumeAt = 0
		}
		stateb, _ := json.Marshal(state)
		node.SetInternalState(stateb)
		return []string{action.GetId()}, nil
	},
}

// Strs should not empty
func getValidLast(strs []string, validationName string) (string, string) {
	if len(strs) == 0 {
		return "", "nil"
	}
	if validationName == "" {
		return strs[len(strs)-1], ""
	}
	validation, has := RegexpValidation_map[validationName]
	if !has || validation.LastMatch == nil {
		return strs[len(strs)-1], ""
	}

	validStr := validation.LastMatch(validation, strings.Join(strs[:], ","))
	if validStr == "" {
		rand.Seed(time.Now().UnixNano())
		return "", validation.InvalidMessage[rand.Intn(len(validation.InvalidMessage))]
	}

	return validStr, ""
}

func isUserResponse(req *header.RunRequest) bool {
	if req.GetBotRunType() != "onevent" {
		return false
	}
	if req.GetEvent().GetType() != header.RealtimeType_message_sent.String() {
		return false
	}
	if req.GetEvent().GetBy().GetType() != pb.Type_user.String() {
		return false
	}
	return true
}

var Refuse_string = []string{"no", "NO", "No", "không", "Không", "KHÔNG", "khong", "KO", "ko", "Ko", "Nope", "nope", "not"}

func isFirstResponseRefuse(requests []*header.RunRequest) bool {
	if len(requests) == 0 {
		return false
	}
	text := requests[0].GetEvent().GetData().GetMessage().GetText()
	for _, str := range Refuse_string {
		if strings.Contains(text, str) {
			return true
		}
	}
	return false
}

func sendMessages(mgr *ActionMgr, accountId, botId, conversationId, userId string, messages []*header.Message, isRoot bool) error {
	if len(messages) == 0 {
		return nil
	}
	allperm := perm.MakeBase()
	cred := ggrpc.ToGrpcCtx(&pb.Context{Credential: &pb.Credential{AccountId: accountId, Issuer: botId, Type: pb.Type_agent, Perm: &allperm}})
	idReq := &pb.Id{AccountId: accountId, Id: conversationId}
	userAttr := &UserAttr{
		Mgr:       mgr,
		AccountId: accountId,
		BotId:     botId,
		UserId:    userId,
	}
	for i, message := range messages {
		formatTexts := unmarshalQuillDelta(message.GetQuillDelta())
		var text string
		if len(formatTexts) > 0 {
			for _, format := range formatTexts {
				if format.Key != "" && userAttr.GetAttrText(format.Key) != "" {
					text += userAttr.GetAttrText(format.Key)
				} else {
					text += format.Text
				}
			}
		}
		if text == "" {
			text = message.GetText()
		}
		fmessage := proto.Clone(message).(*header.Message)
		fmessage.Text = text
		fmessage.ConversationId = conversationId
		fmessage.QuillDelta = ""
		evt := &header.Event{
			AccountId: accountId,
			Type:      header.RealtimeType_message_sent.String(),
			Data:      &header.Event_Data{Message: fmessage},
		}
		var err error
		var index int
		loop.LoopErr(func() error {
			if index > 3 {
				return nil
			}
			// TODO error
			if i != 0 || !isRoot {
				mgr.convo.Typing(cred, idReq)
				time.Sleep(time.Duration(TypingDuration(fmessage.GetText())) * time.Second)
			}
			_, err = mgr.msg.SendMessage(cred, evt)
			index++
			if err == nil {
				return nil
			}
			e, ok := err.(*errors.Error)
			if ok && e.Class == 400 {
				err = nil // TODO ignore error (? if ignore action)
			}
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func readUser(mgr *ActionMgr, accountId, botId, userId string) (*header.User, error) {
	allperm := perm.MakeBase()
	cred := ggrpc.ToGrpcCtx(&pb.Context{Credential: &pb.Credential{AccountId: accountId, Issuer: botId, Type: pb.Type_agent, Perm: &allperm}})
	reqId := &pb.Id{Id: userId}
	var user *header.User
	var err error
	var index int
	loop.LoopErr(func() error {
		if index > 3 {
			return nil
		}
		user, err = mgr.user.ReadUser(cred, reqId)
		index++
		return err
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func upsertUser(mgr *ActionMgr, accountId, botId string, user *header.User) error {
	allperm := perm.MakeBase()
	cred := ggrpc.ToGrpcCtx(&pb.Context{Credential: &pb.Credential{AccountId: accountId, Issuer: botId, Type: pb.Type_agent, Perm: &allperm}})
	var err error
	var index int
	loop.LoopErr(func() error {
		if index > 3 {
			return nil
		}
		_, err = mgr.user.UpdateUser(cred, user)
		index++
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

var Regex_attr_name, _ = regexp.Compile(`([a-zA-Z0-9_-]+)`)

type UserAttr struct {
	Mgr       *ActionMgr
	AccountId string
	BotId     string
	UserId    string
	user      *header.User
	isFetch   bool
	fetchAt   int64
	textMap   map[string]string
}

func (u *UserAttr) GetAttrText(key string) string {
	if key == "" {
		return ""
	}
	partKeys := Regex_attr_name.FindAllString(key, -1)
	if len(partKeys) != 2 || partKeys[0] != "user" {
		return ""
	}
	attrKey := partKeys[1]

	if !u.isFetch {
		err := u.fetch()
		if err != nil {
			return "" // TODO err
		}
	}
	return u.textMap[attrKey]
}

func (u *UserAttr) GetTextMap() map[string]string {
	if !u.isFetch {
		err := u.fetch()
		if err != nil {
			return nil
		}
	}
	return u.textMap
}

func (u *UserAttr) fetch() error {
	u.fetchAt = time.Now().UnixNano() / 1e6
	user, err := readUser(u.Mgr, u.AccountId, u.BotId, u.UserId)
	if err != nil {
		return err
	}
	u.user = user
	u.textMap = make(map[string]string)
	attrs := u.user.GetAttributes()
	// last
	for i := 0; i < len(attrs); i++ {
		u.textMap[attrs[i].GetKey()] = attrs[i].GetText()
	}
	u.isFetch = true
	return nil
}

func renderMessages(messages []*header.Message, userAttr *UserAttr, conversationId string) []*header.Message {
	out := make([]*header.Message, len(messages))
	for i, message := range messages {
		formatTexts := unmarshalQuillDelta(message.GetQuillDelta())
		var text string
		if len(formatTexts) > 0 {
			for _, format := range formatTexts {
				if format.Key != "" && userAttr.GetAttrText(format.Key) != "" {
					text += userAttr.GetAttrText(format.Key)
				} else {
					text += format.Text
				}
			}
		}
		if text == "" {
			text = message.GetText()
		}
		out[i] = proto.Clone(message).(*header.Message)
		out[i].Text = text
		out[i].ConversationId = conversationId
		out[i].QuillDelta = ""
	}
	return out
}
