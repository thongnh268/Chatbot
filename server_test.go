package main

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"github.com/subiz/perm"
	ggrpc "github.com/subiz/sgrpc"
)

func TestTryAction(t *testing.T) {
	server := &Server{}
	allperm := perm.MakeBase()
	accountId := "testacc"
	botId := "bot1"
	cred := ggrpc.ToGrpcCtx(&pb.Context{Credential: &pb.Credential{AccountId: accountId, Issuer: botId, Type: pb.Type_agent, Perm: &allperm}})
	inReq := proto.Clone(Test_fixtures.RunRequest7).(*header.RunRequest)
	inReq.AccountId = accountId
	inReq.BotId = botId
	inReq.Action = inReq.GetBot().GetAction().GetNexts()[0].GetAction()

	_, err := server.TryAction(cred, inReq)
	if err != nil {
		t.Error(err)
	}
}
