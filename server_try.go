package main

import (
	"context"
	"errors"
	"fmt"

	ca "git.subiz.net/bizbot/action"
	cr "git.subiz.net/bizbot/runner"
	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	upb "github.com/subiz/header/user"
	"google.golang.org/grpc"
)

var Try_send_http = &ActionFuncTry{
	Id:         ca.Action_send_http.Id,
	ActionFunc: ca.Action_send_http,
	ActionMgr:  ca.NewActionMgr(nil, &TryUserMgrClient{User: User_sample}, nil, nil),
}

var User_sample = &header.User{Attributes: []*header.Attribute{
	{Key: "phone", Text: "0123456789"},
	{Key: "emails", Text: "dd2020@example.com"},
	{Key: "fullname", Text: "dd"},
	{Key: "undefined", Text: "dieutest"},
}}

var Try_nil = &ActionFuncTry{
	Id:         ca.Action_nil.Id,
	ActionFunc: ca.Action_nil,
	ActionMgr:  ca.NewActionMgr(nil, nil, nil, nil),
}

type ActionFuncTry struct {
	Id         string
	ActionFunc *ca.ActionFunc
	ActionMgr  *ca.ActionMgr
}

func (try *ActionFuncTry) Do(accountId string, botId string, action *header.Action) error {
	if try.ActionFunc == nil {
		return errors.New("action func is nil")
	}
	if try.ActionMgr == nil {
		return errors.New("action mgr is nil")
	}

	node := cr.NewActionNode(&header.Action{Id: action.GetId()})
	node.Reboot(&cr.ExecBot{Bot: &header.Bot{AccountId: accountId, Id: botId, Action: action}})
	req := &header.RunRequest{}
	log.Info("act-"+action.GetId(), "action-func.do=begin", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, action.GetId()), fmt.Sprintf("action-type=%s", action.GetType()))
	_, err := try.ActionFunc.Do(try.ActionMgr, node, req)
	log.Info("act-"+action.GetId(), "action-func.do=end", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, botId, action.GetId()), fmt.Sprintf("err=%s", err))
	return err
}

var Try_arr []*ActionFuncTry
var Try_map map[string]*ActionFuncTry

func init() {
	Try_arr = []*ActionFuncTry{
		Try_send_http,
	}
	Try_map = make(map[string]*ActionFuncTry)
	for _, act := range Try_arr {
		Try_map[act.Id] = act
	}
}

type TryUserMgrClient struct {
	User *header.User
}

func (mgr *TryUserMgrClient) SearchUsers(ctx context.Context, in *upb.UserSearchRequest, opts ...grpc.CallOption) (*header.UserSearchResult, error) {
	return nil, nil
}
func (mgr *TryUserMgrClient) SearchLeads(ctx context.Context, in *header.LeadSearchRequest, opts ...grpc.CallOption) (*header.LeadSearchResult, error) {
	return nil, nil
}
func (mgr *TryUserMgrClient) ListLeads(ctx context.Context, in *header.LeadSearchRequest, opts ...grpc.CallOption) (*header.LeadSearchResult, error) {
	return nil, nil
}
func (mgr *TryUserMgrClient) CreateUser(ctx context.Context, in *header.User, opts ...grpc.CallOption) (*pb.Id, error) {
	return nil, nil
}
func (mgr *TryUserMgrClient) UpdateUser(ctx context.Context, in *header.User, opts ...grpc.CallOption) (*pb.Id, error) {
	return nil, nil
}
func (mgr *TryUserMgrClient) ReadUser(ctx context.Context, in *pb.Id, opts ...grpc.CallOption) (*header.User, error) {
	return mgr.User, nil
}
func (mgr *TryUserMgrClient) ReportUsers(ctx context.Context, in *header.UserReportRequest, opts ...grpc.CallOption) (*header.UserReportResult, error) {
	return nil, nil
}
func (mgr *TryUserMgrClient) CountTotal(ctx context.Context, in *header.CountTotalRequest, opts ...grpc.CallOption) (*header.CountTotalResponse, error) {
	return nil, nil
}
func (mgr *TryUserMgrClient) Ping(ctx context.Context, in *pb.PingRequest, opts ...grpc.CallOption) (*pb.Pong, error) {
	return nil, nil
}
func (mgr *TryUserMgrClient) SearchNote(ctx context.Context, in *upb.SearchNoteRequest, opts ...grpc.CallOption) (*upb.SearchNoteResponse, error) {
	return nil, nil
}
func (mgr *TryUserMgrClient) MatchUsers(ctx context.Context, in *pb.Ids, opts ...grpc.CallOption) (*header.Users, error) {
	return nil, nil
}
