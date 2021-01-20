package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
	"github.com/subiz/header/common"
	pb "github.com/subiz/header/common"
	"github.com/subiz/idgen"
	"github.com/urfave/cli"
)

// TODO update new
func seedDB(clictx *cli.Context) {
	pub := bizbot.GetPub()
	accountId := "testacc"
	number, _ := strconv.Atoi(strings.TrimSpace(clictx.Args().Get(0)))
	number2, _ := strconv.Atoi(strings.TrimSpace(clictx.Args().Get(1)))
	convos := make([]*header.Conversation, number+number2)
	evtMap := make(map[string][]*header.Event, len(convos))
	convoIndex := 0
	tagNum := 100
	tagArr := make([]*header.Tag, tagNum)
	for i := range tagArr {
		tagArr[i] = &header.Tag{
			AccountId: accountId,
			Id:        idgen.NewTagID(),
			Created:   time.Now().UnixNano()/1e6 - rand.Int63n(90*24*3600*1e3),
		}
	}
	memberNum := 20
	memberArr := make([]*header.ConversationMember, memberNum)
	for i := range memberArr {
		var memberId, memberType string
		if i%10 < 3 {
			memberType = pb.Type_agent.String()
			memberId = idgen.NewAgentID()
		} else {
			memberType = pb.Type_user.String()
			memberId = idgen.NewUserID()
		}
		memberArr[i] = &header.ConversationMember{
			AccountId: accountId,
			Id:        memberId,
			Type:      memberType,
		}
	}

	// sematic data
	for i := 0; i < number; i++ {
		tags := make([]*header.Tag, rand.Intn(5))
		for i := range tags {
			tagIndex := rand.Intn(tagNum)
			tags[i] = tagArr[tagIndex]
		}
		members := make([]*header.ConversationMember, rand.Intn(5)+1)
		for i := range members {
			memberIndex := rand.Intn(memberNum)
			members[i] = memberArr[memberIndex]
		}
		created := time.Now().UnixNano()/1e6 - rand.Int63n(90*24*3600*1e3)
		var convoState string
		if i%10 < 2 {
			convoState = header.ConvoState_pending.String()
		} else {
			convoState = header.ConvoState_ended.String()
		}
		convo := &header.Conversation{
			AccountId: accountId,
			Id:        idgen.NewConversationID(),
			Created:   created,
			Tags:      tags,
			Members:   members,
			Ended:     created - rand.Int63n(1*24*3600*1e9),
			State:     convoState,
		}

		evts := make([]*header.Event, len(members)+rand.Intn(5))
		evtIndex := 0
		for _, member := range members {
			evts[evtIndex] = &header.Event{
				AccountId: accountId,
				Id:        idgen.NewEventID(),
				Created:   time.Now().UnixNano()/1e6 - rand.Int63n(90*24*3600*1e3),
				Type:      header.RealtimeType_conversation_invited.String(),
				Data: &header.Event_Data{
					Conversation: &header.Conversation{
						Members: []*header.ConversationMember{member},
					},
				},
			}
			evtIndex++
		}
		for i := len(members); i < len(evts); i++ {
			memberIndex := rand.Intn(len(members))
			evts[evtIndex] = &header.Event{
				AccountId: accountId,
				Id:        idgen.NewEventID(),
				Created:   time.Now().UnixNano()/1e6 - rand.Int63n(90*24*3600*1e3),
				Type:      header.RealtimeType_message_sent.String(),
				By: &pb.By{
					Type: members[memberIndex].GetType(),
					Id:   members[memberIndex].GetId(),
				},
			}
			evtIndex++
		}
		evtMap[convo.GetId()] = evts

		convos[convoIndex] = convo
		convoIndex++
	}
	// TODO syntactic data
	for i := 0; i < number2; i++ {
		convos[convoIndex] = &header.Conversation{AccountId: accountId}
		convoIndex++
	}
	for _, evts := range evtMap {
		for _, evt := range evts {
			evt.Ctx = &common.Context{
				SubTopic: header.E_StartBot.String(),
			}
			pub.Publish(
				header.E_BotSynced.String(),
				evt,
				-1,
				evt.GetId(),
			)
		}
	}

	db := bizbot.GetDB()
	now := time.Now().UnixNano() / 1e6
	bot := &header.Bot{
		AccountId: accountId,
		Id:        "bot1",
		Category:  header.BotCategory_users.String(),
		Created:   now,
		Updated:   now,
		Fullname:  "Conversation Bot",
		Action: &header.Action{
			Id:   "acc1",
			Type: header.ActionType_send_message.String(),
			AskQuestion: &header.ActionAskQuestion{
				Messages: []*header.Message{
					{Text: "hello, i am x"},
				},
			},
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{
						Id:    "acc1.1",
						Type:  header.ActionType_sleep.String(),
						Sleep: &header.ActionSleep{Duration: 5},
						Nexts: []*header.NextAction{
							{Action: &header.Action{Id: "acc1.1.1"}},
							{Action: &header.Action{Id: "acc1.1.2"}},
							{Action: &header.Action{Id: "acc1.1.3"}},
						},
					},
				},
				{
					Action: &header.Action{
						Id:   "acc1.2",
						Type: header.ActionType_condition.String(),
						Nexts: []*header.NextAction{
							{
								Action: &header.Action{
									Id: "acc1.2.1",
									Nexts: []*header.NextAction{
										{Action: &header.Action{Id: "acc1.2.1.1"}},
										{Action: &header.Action{Id: "acc1.2.1.2"}},
										{Action: &header.Action{Id: "acc1.2.1.3"}},
									},
								},
								Condition: &header.Condition{
									Group: header.Condition_all.String(),
									Conditions: []*header.Condition{
										{
											Key:  "data.content.url",
											Type: "string", Group: header.Condition_single.String(),
											String_: &pb.StringParams{Contain: "d"},
										},
										{
											Key:    "created",
											Type:   "number",
											Group:  header.Condition_single.String(),
											Number: &pb.NumberParams{Gte: 0},
										},
									},
								},
							},
							{
								Action: &header.Action{
									Id: "acc1.2.2",
									Nexts: []*header.NextAction{
										{Action: &header.Action{Id: "acc1.2.2.1"}},
										{Action: &header.Action{Id: "acc1.2.2.2"}},
										{Action: &header.Action{Id: "acc1.2.2.3"}},
									},
								},
								Condition: &header.Condition{
									Group: header.Condition_any.String(),
									Conditions: []*header.Condition{
										{
											Key:  "data.content.url",
											Type: "string", Group: header.Condition_single.String(),
											String_: &pb.StringParams{Contain: "d"},
										},
										{
											Key:    "created",
											Type:   "number",
											Group:  header.Condition_single.String(),
											Number: &pb.NumberParams{Gte: 0},
										},
									},
								},
							},
						},
					},
				},
				{
					Action: &header.Action{
						Id:   "acc1.3",
						Type: header.ActionType_jump.String(),
						Jump: &header.ActionJump{
							ActionId: "acc1.1",
						},
					},
				},
			},
		},
	}

	err := db.UpsertBot(bot)
	if err != nil {
		log.Error(err)
	}
}

func seedBot(ctx *cli.Context) {
	number, _ := strconv.Atoi(strings.TrimSpace(ctx.Args().Get(0)))
	db := bizbot.GetDB()
	for i := 0; i < number; i++ {
		bots := seedBots()
		for _, bot := range bots {
			fmt.Println(bot)
			err := db.UpsertBot(bot)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

var active = []string{"active", "inactive"}
var fullname = []string{"Template_Bot Giáo Dục", "Template_Tư vấn ô tô", "Subot", "Template_Bệnh viện/Phòng khám", "Dược phẩm", "Công nghệ - Điện tử", "Template_Bot hỗ trợ website Nội thất"}

func seedBots() []*header.Bot {
	bots := make([]*header.Bot, 10)
	now := time.Now().UnixNano()
	created := now - rand.Int63n(90*24*3600*1e9)
	accId := idgen.NewAccountID()
	for i := 0; i < len(bots); i++ {
		bots[i] = &header.Bot{}
		bots[i].AccountId = accId
		bots[i].Id = idgen.NewBizbotID()
		bots[i].Created = created
		bots[i].State = active[rand.Intn(len(active))]
		bots[i].Fullname = fullname[rand.Intn(len(fullname))]
		bots[i].Action = &header.Action{
			Id:   "acc1.2",
			Type: header.ActionType_condition.String(),
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{
						Id: "acc1.2.1",
						Nexts: []*header.NextAction{
							{Action: &header.Action{Id: "acc1.2.1.1"}},
							{Action: &header.Action{Id: "acc1.2.1.2"}},
							{Action: &header.Action{Id: "acc1.2.1.3"}},
						},
					},
					Condition: &header.Condition{
						Group: header.Condition_all.String(),
						Conditions: []*header.Condition{
							{
								Key:  "data.content.url",
								Type: "string", Group: header.Condition_single.String(),
								String_: &pb.StringParams{Contain: "d"},
							},
							{
								Key:    "created",
								Type:   "number",
								Group:  header.Condition_single.String(),
								Number: &pb.NumberParams{Gte: 0},
							},
						},
					},
				},
				{
					Action: &header.Action{
						Id: "acc1.2.2",
						Nexts: []*header.NextAction{
							{Action: &header.Action{Id: "acc1.2.2.1"}},
							{Action: &header.Action{Id: "acc1.2.2.2"}},
							{Action: &header.Action{Id: "acc1.2.2.3"}},
						},
					},
					Condition: &header.Condition{
						Group: header.Condition_any.String(),
						Conditions: []*header.Condition{
							{
								Key:  "data.content.url",
								Type: "string", Group: header.Condition_single.String(),
								String_: &pb.StringParams{Contain: "d"},
							},
							{
								Key:    "created",
								Type:   "number",
								Group:  header.Condition_single.String(),
								Number: &pb.NumberParams{Gte: 0},
							},
						},
					},
				},
			},
		}
	}
	return bots
}

func fakeBot() *header.Bot {
	bot := &header.Bot{}
	now := time.Now().UnixNano()
	created := now - rand.Int63n(90*24*3600*1e9)
	bot.AccountId = idgen.NewAccountID()
	bot.Id = idgen.NewBizbotID()
	bot.Created = created
	bot.State = active[rand.Intn(len(active))]
	bot.Fullname = fullname[rand.Intn(len(fullname))]
	bot.Action = &header.Action{
		Id:   "acc1.2",
		Type: header.ActionType_condition.String(),
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{
					Id: "acc1.2.1",
					Nexts: []*header.NextAction{
						{Action: &header.Action{Id: "acc1.2.1.1"}},
						{Action: &header.Action{Id: "acc1.2.1.2"}},
						{Action: &header.Action{Id: "acc1.2.1.3"}},
					},
				},
				Condition: &header.Condition{
					Group: header.Condition_all.String(),
					Conditions: []*header.Condition{
						{
							Key:  "data.content.url",
							Type: "string", Group: header.Condition_single.String(),
							String_: &pb.StringParams{Contain: "d"},
						},
						{
							Key:    "created",
							Type:   "number",
							Group:  header.Condition_single.String(),
							Number: &pb.NumberParams{Gte: 0},
						},
					},
				},
			},
			{
				Action: &header.Action{
					Id: "acc1.2.2",
					Nexts: []*header.NextAction{
						{Action: &header.Action{Id: "acc1.2.2.1"}},
						{Action: &header.Action{Id: "acc1.2.2.2"}},
						{Action: &header.Action{Id: "acc1.2.2.3"}},
					},
				},
				Condition: &header.Condition{
					Group: header.Condition_any.String(),
					Conditions: []*header.Condition{
						{
							Key:  "data.content.url",
							Type: "string", Group: header.Condition_single.String(),
							String_: &pb.StringParams{Contain: "d"},
						},
						{
							Key:    "created",
							Type:   "number",
							Group:  header.Condition_single.String(),
							Number: &pb.NumberParams{Gte: 0},
						},
					},
				},
			},
		},
	}
	return bot
}
