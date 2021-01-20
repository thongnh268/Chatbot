package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"sync"
	"testing"
	"time"

	"git.subiz.net/bizbot/action"
	ca "git.subiz.net/bizbot/action"
	cr "git.subiz.net/bizbot/runner"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"github.com/subiz/header/conversation"
	"github.com/subiz/header/user"
	"google.golang.org/grpc"
)

// 60μs
func BenchmarkGetLogicAccounts(b *testing.B) {
	app := &Bizbot{}
	app.ShardNumber = 2048
	app.Shards = make([]*BizbotShard, app.ShardNumber)
	for i := 0; i < app.ShardNumber; i++ {
		app.Shards[i] = &BizbotShard{
			Mutex:         &sync.Mutex{},
			LogicAccounts: make(map[string]*LogicAccount),
		}
	}

	for n := 0; n < b.N; n++ {
		app.GetLogicAccounts()
	}
}

func BenchmarkFilterLogicAccounts(b *testing.B) {
	app := &Bizbot{}
	app.ShardNumber = 2048
	app.Shards = make([]*BizbotShard, app.ShardNumber)
	for i := 0; i < app.ShardNumber; i++ {
		app.Shards[i] = &BizbotShard{
			Mutex:         &sync.Mutex{},
			LogicAccounts: make(map[string]*LogicAccount),
		}
	}
	app.dbMutex = &sync.Mutex{}
	app.db = &DB{}
	app.dialLock = &sync.Mutex{}
	app.realtime = header.NewPubsubClient(nil)
	app.msg = header.NewConversationEventReaderClient(nil)
	app.user = header.NewUserMgrClient(nil)
	app.convo = header.NewConversationMgrClient(nil)
	app.event = header.NewEventMgrClient(nil)
	app.schedMutex = &sync.Mutex{}
	app.sched = &Schedule{}

	for n := 0; n < b.N; n++ {
		app.FilterLogicAccounts([]string{"a", "b", "c"})
	}
}

// 0. mock data & ...

// 1. black box/pure test
// basic case, exception case
// input, expected output, actual output

// 2. white box/pure test

// 3. integration test
// return if not enough env

func TestIntegrationFullBot(t *testing.T) {
	var inRunner *cr.BotRunner
	sched := NewSchedule()
	sched.BatchAsync()

	caMgr := &ActionMgrConv{caMgr: ca.NewActionMgr(
		&TestConversationEventReaderClient{},
		&TestUserMgrClient{},
		&TestConversationMgrClient{},
		&TestEventMgrClient{},
	)}

	inReq := proto.Clone(Test_fixtures.RunRequest6).(*header.RunRequest)
	accountId := inReq.GetBot().GetAccountId()
	objType := inReq.GetObjectType()
	objId := inReq.GetObjectId()
	execBotMgr := &TestExecBotMgr{}
	inRunner = cr.NewBotRunner(accountId, execBotMgr, sched, caMgr)

	sched.BatchFunc = func(idArr []string, schedNow int64) {
		idMap := make(map[string]struct{})
		for _, id := range idArr {
			idMap[id] = struct{}{}
		}
		for id := range idMap {
			if id == accountId {
				go inRunner.OnTime(schedNow)
			}
		}
	}

	inRunner.StartBot(inReq)
	time.Sleep(10 * time.Millisecond)
	time.Sleep(2100 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event1",
			Type:      header.RealtimeType_message_sent.String(),
			By:        &pb.By{Type: pb.Type_user.String()},
			Data:      &header.Event_Data{Message: &header.Message{Text: "Mua xe"}},
		}, Created: time.Now().UnixNano() / 1e6,
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(5100 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event2",
			Type:      header.RealtimeType_message_sent.String(),
			By:        &pb.By{Type: pb.Type_user.String()},
			Data:      &header.Event_Data{Message: &header.Message{Text: "abc"}},
		},
		Created: time.Now().UnixNano() / 1e6,
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(4100 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event3",
			Type:      header.RealtimeType_message_sent.String(),
			By:        &pb.By{Type: pb.Type_user.String()},
			Data:      &header.Event_Data{Message: &header.Message{Text: "Xe mới"}},
		}, Created: time.Now().UnixNano() / 1e6,
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(3100 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event5",
			Type:      header.RealtimeType_message_sent.String(),
			By:        &pb.By{Type: pb.Type_user.String()},
			Data:      &header.Event_Data{Message: &header.Message{Text: "0123456789"}},
		},
		Created: time.Now().UnixNano()/1e6 + 8*1e3,
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(2100 * time.Millisecond)
	if execBotMgr.ExecBot.Status != cr.Exec_bot_state_terminated {
		t.Error("action not work", execBotMgr.ExecBot.ActionTree.HeadActId)
	}
}

func TestIntegrationSendMessage(t *testing.T) {
	var inRunner *cr.BotRunner
	sched := NewSchedule()
	sched.BatchAsync()

	caMgr := &ActionMgrConv{caMgr: ca.NewActionMgr(
		&TestConversationEventReaderClient{},
		&TestUserMgrClient{},
		&TestConversationMgrClient{},
		&TestEventMgrClient{},
	)}

	inReq := proto.Clone(Test_fixtures.RunRequest5).(*header.RunRequest)
	accountId := inReq.GetBot().GetAccountId()
	objType := inReq.GetObjectType()
	objId := inReq.GetObjectId()
	execBotMgr := &TestExecBotMgr{}
	inRunner = cr.NewBotRunner(accountId, execBotMgr, sched, caMgr)

	sched.BatchFunc = func(idArr []string, schedNow int64) {
		idMap := make(map[string]struct{})
		for _, id := range idArr {
			idMap[id] = struct{}{}
		}
		for id := range idMap {
			if id == accountId {
				go inRunner.OnTime(schedNow)
			}
		}
	}

	inRunner.StartBot(inReq)
	time.Sleep(10 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event2",
			Type:      header.RealtimeType_message_sent.String(),
			By:        &pb.By{Type: pb.Type_user.String()},
			Data:      &header.Event_Data{Message: &header.Message{Text: "thong@gmail.com"}},
		},
		Created: time.Now().UnixNano() / 1e6,
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(5000 * time.Millisecond)
	if execBotMgr.ExecBot.Status != cr.Exec_bot_state_terminated {
		t.Error("action not work", execBotMgr.ExecBot.ActionTree.HeadActId)
	}
}

func TestIntegrationBotQuestion(t *testing.T) {
	var inRunner *cr.BotRunner
	sched := NewSchedule()
	sched.BatchAsync()

	caMgr := &ActionMgrConv{caMgr: ca.NewActionMgr(
		&TestConversationEventReaderClient{},
		&TestUserMgrClient{},
		&TestConversationMgrClient{},
		&TestEventMgrClient{},
	)}

	inReq := proto.Clone(Test_fixtures.RunRequest4).(*header.RunRequest)
	accountId := inReq.GetBot().GetAccountId()
	objType := inReq.GetObjectType()
	objId := inReq.GetObjectId()
	execBotMgr := &TestExecBotMgr{}
	inRunner = cr.NewBotRunner(accountId, execBotMgr, sched, caMgr)

	sched.BatchFunc = func(idArr []string, schedNow int64) {
		idMap := make(map[string]struct{})
		for _, id := range idArr {
			idMap[id] = struct{}{}
		}
		for id := range idMap {
			if id == accountId {
				go inRunner.OnTime(schedNow)
			}
		}
	}

	inRunner.StartBot(inReq)
	time.Sleep(10 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event1",
		},
		Created: time.Now().UnixNano() / 1e6,
	})
	time.Sleep(10 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event2",
			Type:      header.RealtimeType_message_sent.String(),
			By:        &pb.By{Type: pb.Type_user.String()},
			Data:      &header.Event_Data{Message: &header.Message{Text: "thong@gmail.com"}},
		},
		Created: time.Now().UnixNano()/1e6 + 5000,
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(8200 * time.Millisecond) // TODO 2200

	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event2",
			Type:      header.RealtimeType_message_sent.String(),
			By:        &pb.By{Type: pb.Type_user.String()},
			Data:      &header.Event_Data{Message: &header.Message{Text: "thong"}},
		},
		Created: time.Now().UnixNano() / 1e6,
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(3200 * time.Millisecond) // TODO 2200
	if execBotMgr.ExecBot.Status != cr.Exec_bot_state_terminated {
		t.Error("action not work", execBotMgr.ExecBot.ActionTree.HeadActId)
	}
}

func TestIntegrationActionQuestion(t *testing.T) {
	var inRunner *cr.BotRunner
	sched := NewSchedule()
	sched.BatchAsync()

	caMgr := &ActionMgrConv{caMgr: ca.NewActionMgr(
		&TestConversationEventReaderClient{},
		&TestUserMgrClient{User: &header.User{Attributes: []*header.Attribute{
			{Key: "phone", Text: "123"},
			{Key: "emails", Text: "thong@gmail.com"},
		}}},
		&TestConversationMgrClient{},
		&TestEventMgrClient{},
	)}

	inReq := proto.Clone(Test_fixtures.RunRequest3).(*header.RunRequest)
	accountId := inReq.GetBot().GetAccountId()
	objType := inReq.GetObjectType()
	objId := inReq.GetObjectId()
	execBotMgr := &TestExecBotMgr{}
	inRunner = cr.NewBotRunner(accountId, execBotMgr, sched, caMgr)

	sched.BatchFunc = func(idArr []string, schedNow int64) {
		idMap := make(map[string]struct{})
		for _, id := range idArr {
			idMap[id] = struct{}{}
		}
		for id := range idMap {
			if id == accountId {
				go inRunner.OnTime(schedNow)
			}
		}
	}

	inRunner.StartBot(inReq)
	time.Sleep(10 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event1",
		},
		Created: time.Now().UnixNano() / 1e6,
	})
	time.Sleep(10 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event2",
			Type:      header.RealtimeType_message_sent.String(),
			By:        &pb.By{Type: pb.Type_user.String()},
			Data:      &header.Event_Data{Message: &header.Message{Text: "momo"}},
		},
		Created: time.Now().UnixNano()/1e6 + 5000,
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(7200 * time.Millisecond) // TODO 2200
	if execBotMgr.ExecBot.Status != cr.Exec_bot_state_terminated {
		t.Error("action not work", execBotMgr.ExecBot.ActionTree.HeadActId)
	}
}

func TestIntegrationAction(t *testing.T) {
	var inRunner *cr.BotRunner
	sched := NewSchedule()
	sched.BatchAsync()

	caMgr := &ActionMgrConv{caMgr: ca.NewActionMgr(
		&TestConversationEventReaderClient{},
		&TestUserMgrClient{User: &header.User{Attributes: []*header.Attribute{{Key: "phone", Text: "123"}}}},
		&TestConversationMgrClient{},
		&TestEventMgrClient{},
	)}

	inReq := proto.Clone(Test_fixtures.RunRequest2).(*header.RunRequest)
	accountId := inReq.GetBot().GetAccountId()
	objType := inReq.GetObjectType()
	objId := inReq.GetObjectId()
	execBotMgr := &TestExecBotMgr{}
	inRunner = cr.NewBotRunner(accountId, execBotMgr, sched, caMgr)

	sched.BatchFunc = func(idArr []string, schedNow int64) {
		idMap := make(map[string]struct{})
		for _, id := range idArr {
			idMap[id] = struct{}{}
		}
		for id := range idMap {
			if id == accountId {
				go inRunner.OnTime(schedNow)
			}
		}
	}

	inRunner.StartBot(inReq)
	time.Sleep(10 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event1",
		},
		Created: time.Now().UnixNano() / 1e6,
	})
	time.Sleep(10 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event2",
			Type:      header.RealtimeType_message_sent.String(),
			By:        &pb.By{Type: pb.Type_user.String()},
			Data:      &header.Event_Data{Message: &header.Message{Text: "Joe Biden"}},
		},
		Created: time.Now().UnixNano()/1e6 + 3000, // TODO AskRange
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(7200 * time.Millisecond) // TODO 2200
	if execBotMgr.ExecBot.Status != cr.Exec_bot_state_terminated {
		t.Error("action not work", execBotMgr.ExecBot.ActionTree.HeadActId)
	}
}

func TestActionMgrConvDo(t *testing.T) {
	var eo, ao []string
	var inNode *cr.ActionNode
	created := time.Now().UnixNano() / 1e6

	act1 := &header.Action{Id: "act1"}
	act2 := &header.Action{Id: "act2"}

	eo = []string{act1.GetId(), act2.GetId()}
	inNode = cr.NewActionNode(&header.Action{
		Id:   "act",
		Type: "sleep",
		Nexts: []*header.NextAction{
			{
				Action: act1,
			},
			{
				Action: act2,
			},
		},
	})
	inNode.Events = []*header.Event{
		{Id: "evt1", Created: created},
		{Id: "evt2", Created: created},
	}
	mgr := &ActionMgrConv{caMgr: action.NewActionMgr(nil, nil, nil, nil)}
	ao, _ = mgr.Do(inNode, nil)
	if !reflect.DeepEqual(ao, eo) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
}

func TestIntegrationBotRunner(t *testing.T) {
	var inRunner *cr.BotRunner
	sched := NewSchedule()
	sched.BatchAsync()

	caMgr := &ActionMgrConv{caMgr: ca.NewActionMgr(
		&TestConversationEventReaderClient{},
		&TestUserMgrClient{User: &header.User{Attributes: []*header.Attribute{{Key: "phone", Text: "123"}}}},
		&TestConversationMgrClient{},
		&TestEventMgrClient{},
	)}

	inReq := proto.Clone(Test_fixtures.RunRequest1).(*header.RunRequest)
	accountId := inReq.GetBot().GetAccountId()
	objType := inReq.GetObjectType()
	objId := inReq.GetObjectId()
	inRunner = cr.NewBotRunner(accountId, &TestExecBotMgr{}, sched, caMgr)

	sched.BatchFunc = func(idArr []string, schedNow int64) {
		idMap := make(map[string]struct{})
		for _, id := range idArr {
			idMap[id] = struct{}{}
		}
		for id := range idMap {
			if id == accountId {
				go inRunner.OnTime(schedNow)
			}
		}
	}

	inRunner.StartBot(inReq)
	time.Sleep(10 * time.Millisecond)
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  accountId,
		ObjectType: objType,
		ObjectId:   objId,
		Event: &header.Event{
			AccountId: accountId,
			Id:        "event1",
		},
		Created: time.Now().UnixNano() / 1e6,
	})
	time.Sleep(10 * time.Millisecond)
	time.Sleep(2200 * time.Millisecond)
}

type TestExecBotMgr struct {
	ExecBot *cr.ExecBot
}

func (mgr *TestExecBotMgr) ReadExecBot(botId, id string) (*cr.ExecBot, error) {
	return nil, nil
}
func (mgr *TestExecBotMgr) UpdateExecBot(execBot *cr.ExecBot) (*cr.ExecBot, error) {
	mgr.ExecBot = execBot
	return nil, nil
}
func (mgr *TestExecBotMgr) CreateExecBotIndex(execBotIndex *cr.ExecBotIndex) (*cr.ExecBotIndex, error) {
	return nil, nil
}
func (mgr *TestExecBotMgr) LoadExecBots(accountId string, pks []*cr.ExecBotPK) chan []*cr.ExecBot {
	return nil
}

type TestUserMgrClient struct {
	User *header.User
}

func (mgr *TestUserMgrClient) SearchUsers(ctx context.Context, in *user.UserSearchRequest, opts ...grpc.CallOption) (*header.UserSearchResult, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) SearchLeads(ctx context.Context, in *header.LeadSearchRequest, opts ...grpc.CallOption) (*header.LeadSearchResult, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) ListLeads(ctx context.Context, in *header.LeadSearchRequest, opts ...grpc.CallOption) (*header.LeadSearchResult, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) CreateUser(ctx context.Context, in *header.User, opts ...grpc.CallOption) (*pb.Id, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) UpdateUser(ctx context.Context, in *header.User, opts ...grpc.CallOption) (*pb.Id, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) ReadUser(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*header.User, error) {
	return mgr.User, nil
}
func (mgr *TestUserMgrClient) ReportUsers(ctx context.Context, in *header.UserReportRequest, opts ...grpc.CallOption) (*header.UserReportResult, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) CountTotal(ctx context.Context, in *header.CountTotalRequest, opts ...grpc.CallOption) (*header.CountTotalResponse, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) Ping(ctx context.Context, in *pb.PingRequest, opts ...grpc.CallOption) (*pb.Pong, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) SearchNote(ctx context.Context, in *user.SearchNoteRequest, opts ...grpc.CallOption) (*user.SearchNoteResponse, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) MatchUsers(ctx context.Context, in *pb.Ids, opts ...grpc.CallOption) (*header.Users, error) {
	return nil, nil
}

type TestConversationEventReaderClient struct{}

func (mgr *TestConversationEventReaderClient) SendMessage(ctx context.Context, in *header.Event, opts ...grpc.CallOption) (*header.Event, error) {
	return nil, nil
}
func (mgr *TestConversationEventReaderClient) UpdateMessage(ctx context.Context, in *header.Event, opts ...grpc.CallOption) (*header.Event, error) {
	return nil, nil
}
func (mgr *TestConversationEventReaderClient) ListEvents(ctx context.Context, in *header.ListConversationEventsRequest, opts ...grpc.CallOption) (*header.Events, error) {
	return nil, nil
}
func (mgr *TestConversationEventReaderClient) SearchEvents(ctx context.Context, in *header.SearchMessageRequest, opts ...grpc.CallOption) (*header.Events, error) {
	return nil, nil
}

type TestEventMgrClient struct{}

func (mgr *TestEventMgrClient) SearchEvents(ctx context.Context, in *header.ListUserEventsRequest, opts ...grpc.CallOption) (*header.Events, error) {
	return nil, nil
}
func (mgr *TestEventMgrClient) CreateEvent(ctx context.Context, in *header.UserEvent, opts ...grpc.CallOption) (*header.Event, error) {
	return nil, nil
}

type TestConversationMgrClient struct{}

func (mgr *TestConversationMgrClient) AssignRule(ctx context.Context, in *header.AssignRequest, opts ...grpc.CallOption) (*header.RouteResult, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) PongMessage(ctx context.Context, in *header.Event, opts ...grpc.CallOption) (*header.Event, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) StartConversation(ctx context.Context, in *header.StartRequest, opts ...grpc.CallOption) (*header.Conversation, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) EndConversation(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*header.Conversation, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) GetConversation(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*header.Conversation, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) ListConversations(ctx context.Context, in *header.ListConversationsRequest, opts ...grpc.CallOption) (*header.Conversations, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) MatchConversations(ctx context.Context, in *pb.Ids, opts ...grpc.CallOption) (*header.Conversations, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) TagConversation(ctx context.Context, in *header.TagRequest, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) UntagConversation(ctx context.Context, in *header.TagRequest, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) JoinConversation(ctx context.Context, in *header.ConversationMember, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) LeftConversation(ctx context.Context, in *header.ConversationMember, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) Typing(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) Ping(ctx context.Context, in *pb.PingRequest, opts ...grpc.CallOption) (*pb.Pong, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) UpdateConversationInfo(ctx context.Context, in *header.Conversation, opts ...grpc.CallOption) (*header.Conversation, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) SearchConversation(ctx context.Context, in *conversation.SearchConversationRequest, opts ...grpc.CallOption) (*conversation.SearchConversationResponse, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) ListConversations2(ctx context.Context, in *conversation.ConversationListRequest, opts ...grpc.CallOption) (*conversation.ConversationListResponse, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) UpdateMuteConversation(ctx context.Context, in *header.Conversation, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) UnwatchConversation(ctx context.Context, in *header.Conversation, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) MarkReadConversation(ctx context.Context, in *header.Conversation, opts ...grpc.CallOption) (*pb.Empty, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) UpdateEndchatSetting(ctx context.Context, in *header.EndchatSetting, opts ...grpc.CallOption) (*header.EndchatSetting, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) GetEndchatSetting(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*header.EndchatSetting, error) {
	return nil, nil
}
func (mgr *TestConversationMgrClient) TerminateBot(ctx context.Context, in *header.BotTerminated, opts ...grpc.CallOption) (*header.Event, error) {
	return nil, nil
}

func TestInitFixtures(t *testing.T) {
	if proto.Equal(Test_fixtures.RunRequest1, &header.RunRequest{}) {
		t.Error("test fixtures is nil")
	}
}

type TestFixtures struct {
	RunRequest1 *header.RunRequest `json:"run_request_1"`
	RunRequest2 *header.RunRequest `json:"run_request_2"`
	RunRequest3 *header.RunRequest `json:"run_request_3"`
	RunRequest4 *header.RunRequest `json:"run_request_4"`
	RunRequest5 *header.RunRequest `json:"run_request_5"`
	RunRequest6 *header.RunRequest `json:"run_request_6"`
	RunRequest7 *header.RunRequest `json:"run_request_7"`
}

var Test_fixtures *TestFixtures

func init() {
	var err error
	fixturesb, err := ioutil.ReadFile("bizbot_test.json")
	if err != nil {
		panic(err)
	}
	Test_fixtures = &TestFixtures{}
	err = json.Unmarshal(fixturesb, Test_fixtures)
	if err != nil {
		panic(err)
	}
}
