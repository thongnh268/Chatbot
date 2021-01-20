package action

import (
	"context"
	"testing"

	"github.com/subiz/header/common"
	"github.com/subiz/header/user"

	"github.com/subiz/header"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// basic case, exception case, fixed case
// input, expected output, actual output

func TestRenderMessages(t *testing.T) {
	var aco, exo, inMessage []*header.Message
	var inUserAttr *UserAttr
	quillDelta := "{\"ops\":[{\"insert\":\"hello \"},{\"insert\":{\"dynamicField\":{\"id\":\"oaqjdmekllkgpclo\",\"value\":\"Họ và tên\",\"key\":\"user.fullname\"}}},{\"insert\":\"  \"},{\"insert\":{\"dynamicField\":{\"id\":\"rqeghvmhqagtapfu\",\"value\":\"Địa chỉ email\",\"key\":\"user.emails\"}}},{\"insert\":\" hi\"}]}"
	mgr := NewActionMgr(nil, &TestUserMgrClient{}, nil, nil)
	inMessage = []*header.Message{
		{Text: "Hello Test"},
		{QuillDelta: quillDelta},
		{Text: "Hello Test", QuillDelta: quillDelta},
	}
	inUserAttr = &UserAttr{
		AccountId: "testacc",
		BotId:     "bot1",
		UserId:    "user1",
		Mgr:       mgr,
	}
	exo = []*header.Message{
		{Text: "Hello Test", ConversationId: "convo1"},
		{Text: "hello dan  d@n.com hi", ConversationId: "convo1"},
		{Text: "hello dan  d@n.com hi", ConversationId: "convo1"},
	}
	aco = renderMessages(inMessage, inUserAttr, "convo1")
	if len(aco) != len(exo) {
		t.Errorf("exo len is not equal aco len")
		return
	}
	for i := 0; i < len(exo); i++ {
		if !proto.Equal(aco[i], exo[i]) {
			t.Errorf("want %s, actual %s", exo[i], aco[i])
		}
	}
}

func TestGetValidLast(t *testing.T) {
	var aco, exo string
	var inStr []string
	inStr = []string{"email cua toi la ducck92@gmail.com"}
	exo = "ducck92@gmail.com"
	aco, _ = getValidLast(inStr, "email")
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}

type TestUserMgrClient struct{}

func (mgr *TestUserMgrClient) SearchUsers(ctx context.Context, in *user.UserSearchRequest, opts ...grpc.CallOption) (*header.UserSearchResult, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) SearchLeads(ctx context.Context, in *header.LeadSearchRequest, opts ...grpc.CallOption) (*header.LeadSearchResult, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) ListLeads(ctx context.Context, in *header.LeadSearchRequest, opts ...grpc.CallOption) (*header.LeadSearchResult, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) CreateUser(ctx context.Context, in *header.User, opts ...grpc.CallOption) (*common.Id, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) UpdateUser(ctx context.Context, in *header.User, opts ...grpc.CallOption) (*common.Id, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) ReadUser(ctx context.Context, in *common.Id, opts ...grpc.CallOption) (*header.User, error) {
	return &header.User{
		Attributes: []*header.Attribute{
			{Key: "phone", Text: "123"},
			{Key: "emails", Text: "d@n.com"},
			{Key: "fullname", Text: "dan"},
			{Key: "undefined", Text: "thong"},
		},
	}, nil
}
func (mgr *TestUserMgrClient) ReportUsers(ctx context.Context, in *header.UserReportRequest, opts ...grpc.CallOption) (*header.UserReportResult, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) CountTotal(ctx context.Context, in *header.CountTotalRequest, opts ...grpc.CallOption) (*header.CountTotalResponse, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) Ping(ctx context.Context, in *common.PingRequest, opts ...grpc.CallOption) (*common.Pong, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) SearchNote(ctx context.Context, in *user.SearchNoteRequest, opts ...grpc.CallOption) (*user.SearchNoteResponse, error) {
	return nil, nil
}
func (mgr *TestUserMgrClient) MatchUsers(ctx context.Context, in *common.Ids, opts ...grpc.CallOption) (*header.Users, error) {
	return nil, nil
}
