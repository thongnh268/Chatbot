package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	cr "git.subiz.net/bizbot/runner"
	E "github.com/subiz/errors"
	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
	cpb "github.com/subiz/header/common"
	pb "github.com/subiz/header/common"
	"github.com/subiz/sgrpc"
)

type LogicAccountMgr interface {
	GetRunner(accid string) *cr.BotRunner
	GetActionMgr(accid string) *ActionMgrConv
	GetBotMgr(accid string) *BotMgr
	GetExecBotMgr(accid string) *ExecBotMgr
}

type Server struct {
	accMgr LogicAccountMgr
}

func (s *Server) ListBots(ctx context.Context, in *pb.Id) (*header.Bots, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return nil, nil
	}

	botMgr := s.accMgr.GetBotMgr(accountId)
	bots, err := botMgr.ListBots(in)
	if err != nil {
		return nil, err
	}

	return &header.Bots{Bots: bots}, nil
}

func (s *Server) GetBot(ctx context.Context, in *pb.Id) (*header.Bot, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return nil, nil
	}

	botMgr := s.accMgr.GetBotMgr(accountId)
	out, err := botMgr.GetBot(in)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Server) CreateBot(ctx context.Context, in *header.Bot) (*header.Bot, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return nil, nil
	}

	botMgr := s.accMgr.GetBotMgr(accountId)
	in.CreatedBy = cred.GetIssuer()
	out, err := botMgr.CreateBot(in)
	if err != nil {
		return nil, err
	}

	if out.GetBotState() == "active" {
		runner := s.accMgr.GetRunner(accountId)
		runner.NewBot(out, nil)
	}

	return out, nil
}

func (s *Server) UpdateBot(ctx context.Context, in *header.Bot) (*header.Bot, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return nil, nil
	}
	in.AccountId = accountId

	botMgr := s.accMgr.GetBotMgr(accountId)
	in.UpdatedBy = cred.GetIssuer()
	in.Updated = time.Now().UnixNano() / 1e6
	out, err := botMgr.UpdateBot(in)
	if err != nil {
		return nil, err
	}

	if out.GetBotState() == "active" {
		report, err := botMgr.GetBotReport(&pb.Id{AccountId: accountId, Id: out.GetId()})
		if err != nil {
			return nil, err
		}
		runner := s.accMgr.GetRunner(accountId)
		runner.NewBot(out, report)
	}

	return out, nil
}

func (s *Server) DeleteBot(ctx context.Context, in *cpb.Id) (*cpb.Empty, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return nil, nil
	}

	botMgr := s.accMgr.GetBotMgr(accountId)
	err := botMgr.DeleteBot(in)
	if err != nil {
		return nil, err
	}
	return &cpb.Empty{}, nil
}

func (s *Server) OnEvent(ctx context.Context, in *header.RunRequest) (*cpb.Empty, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return &cpb.Empty{}, nil
	}
	in.AccountId = accountId
	if in.GetObjectId() == "" {
		return nil, errors.New("obj id is required")
	}
	//  default obj type
	if in.GetObjectType() == "" {
		in.ObjectType = header.BotCategory_conversations.String()
	}
	if in.GetCreated() == 0 {
		in.Created = time.Now().UnixNano() / 1e6
	}

	runner := s.accMgr.GetRunner(accountId)
	log.Info(in.GetObjectId(), "runner.on-event=begin", fmt.Sprintf("account-id=%s", in.GetAccountId()), fmt.Sprintf("evt.type=%s evt.by-id=%s", in.GetEvent().GetType(), in.GetEvent().GetBy().GetId()))
	runner.OnEvent(in)
	log.Info(in.GetObjectId(), "runner.on-event=end", fmt.Sprintf("account-id=%s", in.GetAccountId()), fmt.Sprintf("evt.type=%s evt.by-id=%s", in.GetEvent().GetType(), in.GetEvent().GetBy().GetId()))
	return &cpb.Empty{}, nil
}

func (s *Server) UpdateBotRunState(ctx context.Context, p *header.Bot) (*header.Bot, error) {
	// TODO2 perm
	// cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	botMgr := s.accMgr.GetBotMgr(p.GetAccountId())

	oldbot, err := botMgr.GetBot(&cpb.Id{Id: p.GetId()})
	if err != nil {
		return nil, err
	}

	if oldbot == nil {
		return nil, E.New(400, E.E_setting_bot_not_found, p.GetAccountId(), p.GetId())
	}

	// intentionally skip update Updated and UpdatedBy
	// since this action do not change bot script
	oldbot.BotState = p.GetBotState()
	bot, err := botMgr.UpdateBot(oldbot)
	if err != nil {
		return nil, err
	}
	return bot, nil
}

func (s *Server) StartBot(ctx context.Context, in *header.RunRequest) (*cpb.Empty, error) {
	// TODO2 perm
	// cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := in.GetAccountId()
	if accountId == "" {
		return &cpb.Empty{}, nil
	}
	in.AccountId = accountId
	if in.GetObjectId() == "" {
		return nil, errors.New("obj id is required")
	}

	botId := in.GetBotId()
	if botId == "" && in.GetBot().GetId() != "" {
		botId = in.GetBot().GetId()
		in.BotId = in.GetBot().GetId()
	}

	if botId == "" {
		return nil, errors.New("bot is required")
	}

	if in.GetBot() == nil {
		var err error
		botMgr := s.accMgr.GetBotMgr(accountId)
		// cached
		bot, err := botMgr.GetBot(&pb.Id{AccountId: accountId, Id: in.GetBotId()})
		if err != nil {
			return &cpb.Empty{}, err
		}
		in.Bot = bot
	}

	if in.GetBot() == nil || in.GetBot().GetId() != botId {
		return nil, errors.New("bot is nil or id is invalid")
	}

	//  default obj type
	if in.GetObjectType() == "" {
		in.ObjectType = header.BotCategory_conversations.String()

	}
	if in.GetBot().GetCategory() == "" {
		in.Bot.Category = header.BotCategory_conversations.String()
	}

	if in.GetBot().GetCategory() != in.GetObjectType() {
		return nil, errors.New("bot category is not equal object type")
	}

	if in.GetBot().GetBotState() != "active" && in.GetMode() == "" {
		return nil, errors.New("bot isn't active and mode is nil")
	}

	runner := s.accMgr.GetRunner(accountId)
	log.Info(in.GetObjectId(), "runner.start-bot", fmt.Sprintf("account-id=%s bot-id=%s", in.GetAccountId(), in.GetBotId()), fmt.Sprintf("mode=%s", in.GetMode()))
	err := runner.StartBot(in)
	if err != nil {
		return &cpb.Empty{}, err
	}
	return &cpb.Empty{}, err
}

func (s *Server) StopBot(ctx context.Context, in *header.RunRequest) (*cpb.Empty, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return &cpb.Empty{}, nil
	}
	in.AccountId = accountId
	if in.GetObjectId() == "" {
		return nil, errors.New("obj id is required")
	}
	if in.GetBotId() == "" {
		return nil, errors.New("bot is required")
	}

	//  default obj type
	if in.GetObjectType() == "" {
		in.ObjectType = header.BotCategory_conversations.String()
	}

	runner := s.accMgr.GetRunner(accountId)
	runner.TerminateExecBot(in)
	return &cpb.Empty{}, nil
}

func (s *Server) TryAction(ctx context.Context, in *header.RunRequest) (*cpb.Empty, error) {
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return &cpb.Empty{}, nil
	}
	in.AccountId = accountId
	if in.GetAction() == nil {
		return nil, errors.New("action is required")
	}

	try, has := Try_map[in.GetAction().GetType()]
	if !has {
		try = Try_nil
	}
	log.Info("act-"+in.GetAction().GetId(), "try.do=begin", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, in.GetBotId(), in.GetAction().GetId()), fmt.Sprintf("action-type=%s", in.GetAction().GetType()))
	err := try.Do(accountId, in.GetBotId(), in.GetAction())
	log.Info("act-"+in.GetAction().GetId(), "try.do=end", fmt.Sprintf("account-id=%s bot-id=%s action-id=%s", accountId, in.GetBotId(), in.GetAction().GetId()), fmt.Sprintf("err=%s", err))
	if err != nil {
		return nil, err
	}

	return &cpb.Empty{}, nil
}

func (s *Server) DoAction(ctx context.Context, in *header.RunRequest) (*header.Actions, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return nil, nil
	}

	runner := s.accMgr.GetActionMgr(accountId)
	node := cr.NewActionNode(in.GetAction())
	node.Events = []*header.Event{}
	if in.GetEvent() != nil {
		node.Events = append(node.Events, in.GetEvent())
	}
	actIds, err := runner.Do(node, in)
	if err != nil {
		return nil, err
	}
	out := &header.Actions{
		Actions: make([]*header.Action, len(actIds)),
	}
	for i, actId := range actIds {
		out.Actions[i] = &header.Action{Id: actId}
	}
	return out, nil
}

func (s *Server) ReportBot(ctx context.Context, in *header.ReportBotRequest) (*header.ReportBotResponse, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return nil, nil
	}

	execBotMgr := s.accMgr.GetExecBotMgr(accountId)
	// TODO1 replace
	metrics := execBotMgr.AggregateExecBotIndex(in.GetBotId(), in.GetDayFrom(), in.GetDayTo())
	return &header.ReportBotResponse{Metrics: metrics}, nil
}

func (s *Server) ListObjects(ctx context.Context, in *header.ListObjectsRequest) (*header.ListObjectsResponse, error) {
	// TODO2 perm
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountId := cred.GetAccountId()
	if accountId == "" {
		return nil, nil
	}

	execBotMgr := s.accMgr.GetExecBotMgr(accountId)
	list := execBotMgr.ListObjectIds(in.GetBotId(), in.GetDayFrom(), in.GetDayTo())
	return &header.ListObjectsResponse{List: list}, nil
}

func (s *Server) CreateBotRevision(ctx context.Context, in *header.Bot) (*header.Bot, error) {
	// TODO
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountid := cred.GetAccountId()
	if accountid == "" {
		return nil, nil
	}
	botMgr := s.accMgr.GetBotMgr(accountid)
	out, err := botMgr.CreateBotTag(in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Server) ListBotRevisions(ctx context.Context, in *pb.Id) (*header.Bots, error) {
	// TODO
	cred := sgrpc.FromGrpcCtx(ctx).GetCredential()
	accountid := cred.GetAccountId()
	if accountid == "" {
		return nil, nil
	}
	botMgr := s.accMgr.GetBotMgr(accountid)
	bots, err := botMgr.ListBotTags(in)
	if err != nil {
		return nil, err
	}
	return &header.Bots{Bots: bots}, nil
}
