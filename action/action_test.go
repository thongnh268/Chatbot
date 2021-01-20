package action

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	cr "git.subiz.net/bizbot/runner"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"github.com/subiz/header/conversation"

	"google.golang.org/grpc"
)

// basic case, exception case, fixed case
// input, expected output, actual output

func TestSendHttpResquest(t *testing.T) {
	method := "POST"
	url := "http://noaddress.subiz.com"
	headers := []*pb.KV{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "Authorization", Value: "0545612212"},
	}
	quilldelta := `{\"ops\":[{\"insert\":{\"dynamicField\":{\"id\":\"kmkhjkwvefhrqmpn\",\"value\":\"user\",\"key\":\"user\"}}},{\"insert\":\"\\n\\n\"}]}`
	var payload string
	formatTexts := unmarshalQuillDelta(quilldelta)
	mgr := NewActionMgr(nil, &TestUserMgrClient{}, nil, nil)
	userAttr := &UserAttr{
		AccountId: "testacc",
		BotId:     "bot1",
		UserId:    "user1",
		Mgr:       mgr,
	}
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
	err := sendHttpResquest(method, url, payload, headers)
	if err.Error() != "Post \"http://noaddress.subiz.com\": dial tcp: lookup noaddress.subiz.com: No address associated with hostname" {
		t.Errorf("%s", err)
	}
}

func TestAction_send_http(t *testing.T) {
	mgr := NewActionMgr(nil, &TestUserMgrClient{}, nil, nil)
	node := cr.NewActionNode(&header.Action{Id: "1"})
	node.Reboot(&cr.ExecBot{Bot: &header.Bot{AccountId: "testacc", Id: "bot1", Action: &header.Action{
		Id: "1",
		SendHttp: &header.ActionSendHttp{
			Method: "POST",
			Header: []*pb.KV{
				{
					Key:   "content-type",
					Value: "application/json",
				},
			},
			Url:        "https://enkhn7hqa7zf.x.pipedream.net/",
			QuillDelta: `{"ops":[{"insert":"{\n\t\"user\": "},{"insert":{"dynamicField":{"id":"jhfjqvfyifrxlotl","value":"user","key":"user"}}},{"insert":"\n}\n"}]}`,
		},
	}}})
	req := &header.RunRequest{}
	ids, _ := Action_send_http.Do(mgr, node, req)
	exo := []string{}
	aco := ids
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}

func TestAction_update_conversation(t *testing.T) {
	mgr := NewActionMgr(nil, nil, &TestConvoMgrClient{}, nil)
	actionUpdateConvo := &header.Action{
		Id: "act1",
		UpdateConversation: &header.ActionUpdateConversation{
			EndConversation: false,
			TagIds:          []string{"quan trọng", "khách hàng tiềm năng"},
			UntagIds:        []string{"quan trọng"},
		},
		Nexts: []*header.NextAction{
			{Action: &header.Action{Id: "next1"}},
		},
	}
	node := cr.NewActionNode(&header.Action{Id: "act1"})
	node.Reboot(&cr.ExecBot{Bot: &header.Bot{AccountId: "testacc", Id: "bot1", Action: actionUpdateConvo}})
	req := &header.RunRequest{}
	ids, _ := Action_update_conversation.Do(mgr, node, req)
	exo := []string{"next1"}
	aco := ids
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %v, actual %#v", exo, aco)
	}
}

func TestAction_sleepDo(t *testing.T) {
	var eo, ao []string
	var inNode *cr.ActionNode

	act1 := &header.Action{Id: "act1"}
	act2 := &header.Action{Id: "act2"}
	eo = []string{act1.GetId(), act2.GetId()}
	inNode = cr.NewActionNode(&header.Action{
		Id:   "act",
		Type: "sleep",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{Id: "act1"},
			},
			{
				Action: &header.Action{Id: "act2"},
			},
		},
	})
	ao, _ = Action_sleep.Do(nil, inNode, nil)
	if !reflect.DeepEqual(ao, eo) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
}

// 5.34ns
func BenchmarkParallelMapSelect(b *testing.B) {
	b.SetParallelism(5000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testMapSelect("func8")
		}
	})
}

// 5.81ns
func BenchmarkParallelSwitchCase(b *testing.B) {
	b.SetParallelism(5000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testSwitchCase("func8")
		}
	})
}

// 9.45ns
func BenchmarkMapSelect(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testMapSelect("func8")
	}
}

// 11.3ns
func BenchmarkSwitchCase(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testSwitchCase("func8")
	}
}

var funcMap = map[string]func(){
	"func1":  func() {},
	"func2":  func() {},
	"func3":  func() {},
	"func4":  func() {},
	"func5":  func() {},
	"func6":  func() {},
	"func7":  func() {},
	"func8":  func() {},
	"func9":  func() {},
	"func10": func() {},
	"func11": func() {},
	"func12": func() {},
	"func13": func() {},
	"func14": func() {},
	"func15": func() {},
	"func16": func() {},
}

func testMapSelect(funcId string) {
	// funcIndex := rand.Intn(len(funcMap)) + 1
	// funcId := "func" + strconv.Itoa(funcIndex)
	actionFunc := funcMap[funcId]
	actionFunc()
}

func testSwitchCase(funcId string) {
	// funcIndex := rand.Intn(16) + 1
	// funcId := "func" + strconv.Itoa(funcIndex)
	var actionFunc func()
	switch funcId {
	case "func1":
		actionFunc = func() {}
	case "func2":
		actionFunc = func() {}
	case "func3":
		actionFunc = func() {}
	case "func4":
		actionFunc = func() {}
	case "func5":
		actionFunc = func() {}
	case "func6":
		actionFunc = func() {}
	case "func7":
		actionFunc = func() {}
	case "func8":
		actionFunc = func() {}
	case "func9":
		actionFunc = func() {}
	case "func10":
		actionFunc = func() {}
	case "func11":
		actionFunc = func() {}
	case "func12":
		actionFunc = func() {}
	case "func13":
		actionFunc = func() {}
	case "func14":
		actionFunc = func() {}
	case "func15":
		actionFunc = func() {}
	case "func16":
		actionFunc = func() {}
	}
	actionFunc()
}

type TestConvoMgrClient struct{}

func (mgr *TestConvoMgrClient) AssignRule(ctx context.Context, in *header.AssignRequest, opts ...grpc.CallOption) (*header.RouteResult, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) PongMessage(ctx context.Context, in *header.Event, opts ...grpc.CallOption) (*header.Event, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) StartConversation(ctx context.Context, in *header.StartRequest, opts ...grpc.CallOption) (*header.Conversation, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) EndConversation(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*header.Conversation, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) GetConversation(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*header.Conversation, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) ListConversations(ctx context.Context, in *header.ListConversationsRequest, opts ...grpc.CallOption) (*header.Conversations, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) MatchConversations(ctx context.Context, in *pb.Ids, opts ...grpc.CallOption) (*header.Conversations, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) TagConversation(ctx context.Context, in *header.TagRequest, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) UntagConversation(ctx context.Context, in *header.TagRequest, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) JoinConversation(ctx context.Context, in *header.ConversationMember, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) LeftConversation(ctx context.Context, in *header.ConversationMember, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) Typing(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) Ping(ctx context.Context, in *pb.PingRequest, opts ...grpc.CallOption) (*pb.Pong, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) UpdateConversationInfo(ctx context.Context, in *header.Conversation, opts ...grpc.CallOption) (*header.Conversation, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) SearchConversation(ctx context.Context, in *conversation.SearchConversationRequest, opts ...grpc.CallOption) (*conversation.SearchConversationResponse, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) ListConversations2(ctx context.Context, in *conversation.ConversationListRequest, opts ...grpc.CallOption) (*conversation.ConversationListResponse, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) UpdateMuteConversation(ctx context.Context, in *header.Conversation, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) UnwatchConversation(ctx context.Context, in *header.Conversation, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) MarkReadConversation(ctx context.Context, in *header.Conversation, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) UpdateEndchatSetting(ctx context.Context, in *header.EndchatSetting, opts ...grpc.CallOption) (*header.EndchatSetting, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) GetEndchatSetting(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*header.EndchatSetting, error) {
	return nil, nil
}
func (mgr *TestConvoMgrClient) TerminateBot(ctx context.Context, in *header.BotTerminated, opts ...grpc.CallOption) (*header.Event, error) {
	return nil, nil
}
