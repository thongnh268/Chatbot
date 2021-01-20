package runner

import (
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	ca "git.subiz.net/bizbot/action"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"google.golang.org/protobuf/proto"
)

// basic case, exception case, fixed caseCompareEvent
// input, expected output, actual output

func TestBot(t *testing.T) {
	var inBot *header.Bot
	var inRunner *BotRunner
	inBot = &header.Bot{
		AccountId: "testacc",
		Id:        "bbqvehnmuqpxjkvdrq",
		Category:  "conversations",
		Action: &header.Action{
			Name: "Gửi tin nhắn",
			Type: "send_message",
		},
	}
	inRunner = NewBotRunner("testacc", nil, nil, &TestActionMgr{})
	inRunner.StartBot(&header.RunRequest{
		AccountId:  "testacc",
		Bot:        inBot,
		Mode:       "debugging",
		ObjectType: "conversations",
		ObjectId:   "conversation.1",
	})
	inRunner.OnEvent(&header.RunRequest{
		AccountId:  "testacc",
		ObjectType: "conversations",
		ObjectId:   "conversation.1",
		Event: &header.Event{
			AccountId: "testacc",
			Id:        "event1",
		},
	})
	time.Sleep(10 * time.Millisecond)
	execBotId := NewExecBotID(inBot.GetAccountId(), inBot.GetId(), "conversation.1")

	if _, has := inRunner.execBots[execBotId]; has {
		t.Error("run go wrong")
	}
}

func TestAddSchedExecBots(t *testing.T) {
	var inRunner *BotRunner
	var exo, aco map[int64][]string
	var inExecBot *ExecBot

	now := time.Now().UnixNano() / 1e6

	inRunner = NewBotRunner("testacc", nil, &SchedTest{}, nil)
	inExecBot = &ExecBot{
		Id: "execbot1",
		SchedActs: map[string]int64{
			"act1": (now + 2*1e3) / 1e2,
			"act2": (now + 3*1e3) / 1e2,
		},
	}
	inRunner.addSchedExecBots(inExecBot)
	exo = map[int64][]string{(now + 2*1e3) / 1e2: {"execbot1"}}
	aco = inRunner.schedExecBots
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inRunner = NewBotRunner("testacc", nil, &SchedTest{}, nil)
	inExecBot1 := &ExecBot{
		Id:        "execbot1",
		SchedActs: map[string]int64{"act1": (now + 2*1e3) / 1e2},
	}
	inExecBot2 := &ExecBot{
		Id:        "execbot2",
		SchedActs: map[string]int64{"act1": (now + 2*1e3) / 1e2},
	}
	inRunner.addSchedExecBots(inExecBot1)
	inRunner.addSchedExecBots(inExecBot2)
	exo = map[int64][]string{(now + 2*1e3) / 1e2: {"execbot1", "execbot2"}}
	aco = inRunner.schedExecBots
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inRunner = NewBotRunner("testacc", nil, &SchedTest{}, nil)
	inExecBot = &ExecBot{
		Id: "execbot1",
		SchedActs: map[string]int64{
			"act1": (now - 2*1e3) / 1e2,
			"act2": (now - 1*1e3) / 1e2,
			"act3": now / 1e2,
		},
	}
	inRunner.addSchedExecBots(inExecBot)
	exo = map[int64][]string{now / 1e2: {"execbot1"}}
	aco = inRunner.schedExecBots
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}

func TestCleanExecBots(t *testing.T) {
	var inRunner *BotRunner
	now := time.Now().UnixNano() / 1e6
	inExecBot := &ExecBot{
		Id:       "execbot1",
		LastSeen: now - Runner_lock_time_expire - 1,
		ObjId:    "obj1",
	}
	inRunner = NewBotRunner("testacc", &TestExecBotMgr{}, nil, nil)
	inRunner.suspendExecBots[inExecBot.Id] = inExecBot
	inRunner.objExecBots = map[string][]string{
		inExecBot.ObjId: {inExecBot.Id},
	}
	inRunner.CleanExecBots()
	if _, has := inRunner.suspendExecBots[inExecBot.Id]; has {
		t.Error("cant remove from exec bots")
	}
	if _, has := inRunner.objExecBots[inExecBot.ObjId]; has {
		t.Error("cant remove from obj exec bots")
	}
}

func TestOnLock(t *testing.T) {
	var inRunner *BotRunner

	now := time.Now().UnixNano() / 1e6
	inExecBot := &ExecBot{
		Id:       "execbot1",
		LastSeen: now - Runner_lock_time_out - 1,
	}
	inRunner = NewBotRunner("testacc", &TestExecBotMgr{}, nil, nil)
	inRunner.execBots[inExecBot.Id] = inExecBot
	inRunner.triggerExecBots = map[string][]string{
		"trigger1": {inExecBot.Id},
	}
	inRunner.execBotTriggers = map[string][]string{
		inExecBot.Id: {"trigger1"},
	}
	inRunner.OnLock()
	if _, has := inRunner.execBots[inExecBot.Id]; has {
		t.Error("cant remove from exec bots")
	}
	for _, id := range inRunner.triggerExecBots["trigger1"] {
		if id == inExecBot.Id {
			t.Error("cant remove old trigger from index")
		}
	}
	if _, has := inRunner.suspendExecBots[inExecBot.Id]; !has {
		t.Error("exec bot not found in suspend exec bots")
	}
}

func TestOnForgot(t *testing.T) {
	var inRunner *BotRunner
	now := time.Now().UnixNano() / 1e6
	inRunner = NewBotRunner("testacc", nil, &SchedTest{}, Test_action_mgr)
	inRunner.schedExecBots = make(map[int64][]string)
	inRunner.schedExecBots[now/1e2-123] = []string{"execbot1"}
	inRunner.OnForgot()
	if inRunner.schedAt != now/1e2 {
		t.Error("onforgot not working")
	}
}

func TestOnTime(t *testing.T) {
	var inRunner *BotRunner
	var inExecBot *ExecBot
	now := time.Now().UnixNano() / 1e6
	node1 := &ActionNode{
		ActionId: "act2",
		action: &header.Action{
			Id:   "act2",
			Type: header.ActionType_sleep.String(),
			Sleep: &header.ActionSleep{
				Duration: 10,
			},
		},
		State:        Node_state_waiting,
		FirstSeen:    now,
		FirstRequest: now,
		HeadActId:    "act2",
	}
	inExecBot = &ExecBot{
		AccountId:  "acc123",
		Id:         "execbot1",
		ObjId:      "obj123",
		Bot:        &header.Bot{Category: "obj"},
		ActionTree: node1,
		actionMgr:  Test_action_mgr,
	}
	node1.SetExecBot(inExecBot)
	inRunner = NewBotRunner("testacc", nil, &SchedTest{}, Test_action_mgr)
	inRunner.execBots[inExecBot.Id] = inExecBot
	inRunner.schedExecBots = map[int64][]string{now / 1e2: {inExecBot.Id}}
	inExecBot.afterRun = inRunner.AfterRunExecBot
	inRunner.OnTime(now / 1e2)
	time.Sleep(1 * time.Millisecond)
	for _, id := range inRunner.schedExecBots[now/1e2] {
		if id == inExecBot.Id {
			t.Error("cant remove exec bot from sched")
		}
	}
	isScheduled := false
	for schedAt, ids := range inRunner.schedExecBots {
		for _, id := range ids {
			if id == inExecBot.Id && schedAt > now/1e2 {
				isScheduled = true
			}
		}
	}
	if !isScheduled {
		t.Error("cant add exec bot to sched")
	}
}

func TestOnEvent(t *testing.T) {
	var inRunner *BotRunner
	var inExecBot *ExecBot
	inRunner = NewBotRunner("testacc", nil, nil, Test_action_mgr)
	act2 := &header.Action{
		Id: "act2",
	}
	cond := &header.Condition{
		Group: header.Condition_all.String(),
		Conditions: []*header.Condition{
			{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
			{Key: "created", Type: "number", Group: header.Condition_single.String()},
		},
	}
	act1 := &header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{Action: act2, Condition: cond},
		},
	}
	inExecBot = &ExecBot{
		AccountId: "acc123",
		Id:        "execbot1",
		ObjId:     "obj123",
		BotId:     "bot1",
		Bot:       &header.Bot{Category: "obj", Id: "bot1"},
		ActionTree: &ActionNode{
			ActionId: act1.Id,
			action:   act1,
			State:    Node_state_waiting,
			Nexts: []*ActionNode{
				{
					action: act2,
				},
			},
			HeadActId: act1.GetId(),
		},
	}
	inRunner.execBots[inExecBot.Id] = inExecBot
	inRunner.triggerExecBots = map[string][]string{
		"trigger1":                      {inExecBot.Id},
		"acc123.obj.obj123.event_exist": {inExecBot.Id},
	}
	inRunner.execBotTriggers = map[string][]string{
		inExecBot.Id: {"trigger1", "acc123.obj.obj123.event_exist"},
	}
	inExecBot.afterRun = inRunner.AfterRunExecBot
	inRunner.OnEvent(&header.RunRequest{
		ObjectId:   "obj123",
		ObjectType: "obj",
		Event: &header.Event{
			AccountId: "acc123",
		},
	})
	time.Sleep(1 * time.Millisecond)
	if _, has := inRunner.runningExecBots[inExecBot.Id]; has {
		t.Error("cant remove from running exec bots")
	}
	for _, id := range inRunner.triggerExecBots["trigger1"] {
		if id == inExecBot.Id {
			t.Error("cant remove old trigger from index")
		}
	}
	if _, has := inRunner.execBots[inExecBot.Id]; !has {
		t.Error("exec bot not found in exec bots")
	}
	exoExecBotTriggers := []string{"acc123.obj.obj123.data.content.url", "acc123.obj.obj123.created"}
	if !reflect.DeepEqual(inRunner.execBotTriggers[inExecBot.Id], exoExecBotTriggers) {
		t.Error("cant add exec bot to indexs")
		t.Errorf("exec bot triggers is %#v", inRunner.execBotTriggers[inExecBot.Id])
	}
}

func TestAfterRunExecBot(t *testing.T) {
	var inRunner *BotRunner
	var inExecBot *ExecBot
	inRunner = NewBotRunner("testacc", nil, nil, Test_action_mgr)
	act2 := &header.Action{
		Id: "act2",
	}
	cond := &header.Condition{
		Group: header.Condition_all.String(),
		Conditions: []*header.Condition{
			{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
			{Key: "created", Type: "number", Group: header.Condition_single.String()},
		},
	}
	act1 := &header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{Action: act2, Condition: cond},
		},
	}
	inExecBot = &ExecBot{
		AccountId: "acc123",
		Id:        "execbot1",
		ObjId:     "obj123",
		Bot:       &header.Bot{Category: "obj"},
		ActionTree: &ActionNode{
			ActionId: act1.Id,
			action:   act1,
			State:    Node_state_waiting,
			Nexts: []*ActionNode{
				{
					ActionId: act2.Id, action: act2,
				},
			},
			HeadActId: act1.GetId(),
		},
	}
	inRunner.execBots[inExecBot.Id] = inExecBot
	inRunner.triggerExecBots = map[string][]string{
		"trigger1": {inExecBot.Id},
	}
	inRunner.execBotTriggers = map[string][]string{
		inExecBot.Id: {"trigger1"},
	}
	inRunner.runningExecBots = map[string]int64{inExecBot.Id: 123}
	inRunner.AfterRunExecBot(inExecBot)
	if _, has := inRunner.runningExecBots[inExecBot.Id]; has {
		t.Error("cant remove from running exec bots")
	}
	for _, id := range inRunner.triggerExecBots["trigger1"] {
		if id == inExecBot.Id {
			t.Error("cant remove old trigger from index")
		}
	}
	if _, has := inRunner.execBots[inExecBot.Id]; !has {
		t.Error("exec bot not found in exec bots")
	}
	exoExecBotTriggers := []string{"acc123.obj.obj123.data.content.url", "acc123.obj.obj123.created"}
	if !reflect.DeepEqual(inRunner.execBotTriggers[inExecBot.Id], exoExecBotTriggers) {
		t.Error("cant add exec bot to indexs")
		t.Errorf("exec bot triggers is %#v", inRunner.execBotTriggers[inExecBot.Id])
	}
}

func TestTerminateExecBot(t *testing.T) {
	var inRunner *BotRunner
	var inExecBot *ExecBot
	inRunner = NewBotRunner("testacc", nil, nil, nil)
	inExecBot = &ExecBot{Id: "obj1.bot1"}
	inRunner.execBots[inExecBot.Id] = inExecBot
	inRunner.triggerExecBots = map[string][]string{
		"trigger1": {inExecBot.Id},
	}
	inRunner.execBotTriggers = map[string][]string{
		inExecBot.Id: {"trigger1"},
	}
	inRunner.TerminateExecBot(&header.RunRequest{
		ObjectId: "obj1",
		BotId:    "bot1",
	})
	if _, has := inRunner.execBots[inExecBot.Id]; has {
		t.Error("cant remove from exec bots")
	}
	if _, has := inRunner.execBotTriggers[inExecBot.Id]; has {
		t.Error("cant remove from index")
	}
	if inExecBot.Status != Exec_bot_state_terminated {
		t.Error("status was wrong")
	}

	inRunner = NewBotRunner("testacc", nil, nil, nil)
	inExecBot = &ExecBot{Id: "obj1.bot1"}
	inRunner.execBots[inExecBot.Id] = inExecBot
	inRunner.triggerExecBots = map[string][]string{
		"trigger1": {inExecBot.Id},
	}
	inRunner.execBotTriggers = map[string][]string{
		inExecBot.Id: {"trigger1"},
	}
	inRunner.runningExecBots = map[string]int64{inExecBot.Id: 123}
	inRunner.TerminateExecBot(&header.RunRequest{
		ObjectId: "obj1",
		BotId:    "bot1",
	})
	inExecBot.afterRun = inRunner.AfterRunExecBot
	inExecBot.afterRun(inExecBot)
	if _, has := inRunner.execBots[inExecBot.Id]; has {
		t.Error("cant remove from exec bots")
	}
	if _, has := inRunner.execBotTriggers[inExecBot.Id]; has {
		t.Error("cant remove from index")
	}
	if inExecBot.Status != Exec_bot_state_terminated {
		t.Error("status was wrong")
	}
}

func TestLoadExecBots(t *testing.T) {
	var inRunner *BotRunner
	inRunner = NewBotRunner("testacc", nil, nil, Test_action_mgr)
	inRunner.logicBots = map[string]*LogicBot{"bot1": {}}
	c := make(chan []*ExecBot)
	act2 := &header.Action{
		Id: "act2",
	}
	cond := &header.Condition{
		Group: header.Condition_all.String(),
		Conditions: []*header.Condition{
			{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
			{Key: "created", Type: "number", Group: header.Condition_single.String()},
		},
	}
	act1 := &header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{Action: act2, Condition: cond},
		},
	}
	execBot := &ExecBot{
		AccountId: "acc123",
		ObjId:     "obj123",
		Id:        "execbot1",
		BotId:     "bot1",
		Bot:       &header.Bot{Category: "obj"},
		ActionTree: &ActionNode{
			ActionId: act1.Id,
			action:   act1,
			State:    Node_state_waiting,
			Nexts: []*ActionNode{
				{
					ActionId: act2.Id, action: act2,
				},
			},
		},
		LastSeen: time.Now().UnixNano() / 1e6,
	}
	go func() {
		c <- []*ExecBot{execBot}
		close(c)
	}()
	for srcExecBots := range c {
		inRunner.LoadExecBots(srcExecBots)
	}

	if _, has := inRunner.execBots[execBot.Id]; !has {
		t.Error("exec bot not found")
	}
}

func TestCleanIndexings(t *testing.T) {
	var inRunner *BotRunner
	var exo, aco map[string][]string
	inRunner = NewBotRunner("testacc", nil, nil, nil)
	inRunner.triggerBots = map[string][]string{
		"trigger1": {},
		"trigger2": {"bot2"},
	}
	inRunner.triggerExecBots = map[string][]string{
		"trigger1.execbot1": {},
		"trigger2.execbot2": {"execbot2"},
	}
	inRunner.execBotTriggers = map[string][]string{
		"execbot1": {},
		"execbot2": {"trigger2.execbot2"},
	}
	inRunner.objExecBots = map[string][]string{
		"obj1": {},
		"obj2": {"execbot2"},
	}
	inRunner.schedExecBots = map[int64][]string{
		123: {},
		124: {"execbot3"},
	}
	inRunner.CleanIndexings()

	exo = map[string][]string{
		"trigger2": {"bot2"},
	}
	aco = inRunner.triggerBots
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	exo = map[string][]string{
		"trigger2.execbot2": {"execbot2"},
	}
	aco = inRunner.triggerExecBots
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	exo = map[string][]string{
		"execbot2": {"trigger2.execbot2"},
	}
	aco = inRunner.execBotTriggers
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	exo = map[string][]string{
		"obj2": {"execbot2"},
	}
	aco = inRunner.objExecBots
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	exoInt64 := map[int64][]string{
		124: {"execbot3"},
	}
	acoInt64 := inRunner.schedExecBots
	if !reflect.DeepEqual(acoInt64, exoInt64) {
		t.Errorf("want %#v, actual %#v", exoInt64, acoInt64)
	}
}

func TestNewBot(t *testing.T) {
	var inRunner *BotRunner
	var inBot *header.Bot
	var exoIndexTriggers map[string][]string
	var acoIndexTriggers map[string][]string

	inRunner = NewBotRunner("testacc", nil, nil, Test_action_mgr)
	inBot = &header.Bot{
		AccountId: "testacc",
		Category:  header.BotCategory_users.String(),
		Id:        "bot1",
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{Id: "acc123.1"},
					Condition: &header.Condition{
						Key:   "data.content.url",
						Type:  "string",
						Group: header.Condition_single.String(),
					},
				},
			},
		},
	}
	exoIndexTriggers = map[string][]string{"testacc.users.data.content.url": {"bot1"}}
	inRunner.NewBot(inBot, nil)
	acoIndexTriggers = inRunner.triggerBots
	if !reflect.DeepEqual(acoIndexTriggers, exoIndexTriggers) {
		t.Errorf("want %#v, actual %#v", exoIndexTriggers, acoIndexTriggers)
	}
}

func TestRemoveBotFromIndexTriggers(t *testing.T) {
	// var bot *header.Bot
	bot1 := &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
		Id:        "bot1",
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{Id: "acc123.1"},
					Condition: &header.Condition{
						Group: header.Condition_all.String(),
						Conditions: []*header.Condition{
							{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
							{Key: "created", Type: "number", Group: header.Condition_single.String()},
						},
					},
				},
			},
		},
	}
	var inRunner *BotRunner
	var eoIndexTriggers map[string][]string
	var aoIndexTriggers map[string][]string

	eoIndexTriggers = map[string][]string{
		"acc123.users.created": {"bot2"},
	}
	inRunner = &BotRunner{
		triggerBots: map[string][]string{
			"acc123.users.data.content.url": {"bot1"},
			"acc123.users.created":          {"bot1", "bot2"},
		},
	}
	inRunner.actionMgr = Test_action_mgr
	inRunner.rmTriggerBots(bot1)
	aoIndexTriggers = inRunner.triggerBots
	if !reflect.DeepEqual(eoIndexTriggers, aoIndexTriggers) {
		t.Errorf("want %v, actual %v", eoIndexTriggers, aoIndexTriggers)
	}
}

func TestRemoveExecBotFromIndexTriggers(t *testing.T) {
	var inExecBotId string
	var inRunner *BotRunner
	var eoFgIndexExecTriggers, eoFgIndexTriggerExecs map[string][]string
	var aoFgIndexExecTriggers, aoFgIndexTriggerExecs map[string][]string

	eoFgIndexExecTriggers = map[string][]string{"execbot2": {"created"}}
	eoFgIndexTriggerExecs = map[string][]string{"data.content.url": {}, "created": {"execbot2"}}
	inExecBotId = "execbot1"
	inRunner = &BotRunner{
		execBotTriggers: map[string][]string{"execbot1": {"data.content.url", "created"}, "execbot2": {"created"}},
		triggerExecBots: map[string][]string{"data.content.url": {"execbot1"}, "created": {"execbot1", "execbot2"}},
	}
	inRunner.actionMgr = Test_action_mgr
	inRunner.rmTriggerExecBots(inExecBotId)
	aoFgIndexExecTriggers = inRunner.execBotTriggers
	aoFgIndexTriggerExecs = inRunner.triggerExecBots
	if !reflect.DeepEqual(eoFgIndexExecTriggers, aoFgIndexExecTriggers) {
		t.Errorf("want %v, actual %v", eoFgIndexExecTriggers, aoFgIndexExecTriggers)
	}
	if !reflect.DeepEqual(eoFgIndexTriggerExecs, aoFgIndexTriggerExecs) {
		t.Errorf("want %v, actual %v", eoFgIndexTriggerExecs, aoFgIndexTriggerExecs)
	}

	eoFgIndexExecTriggers = map[string][]string{}
	eoFgIndexTriggerExecs = map[string][]string{"data.content.url": {}, "created": {}}
	inExecBotId = "execbot1"
	inRunner = &BotRunner{
		execBotTriggers: map[string][]string{"execbot1": {"data.content.url", "created"}},
		triggerExecBots: map[string][]string{"data.content.url": {"execbot1"}, "created": {"execbot1"}},
	}
	inRunner.actionMgr = Test_action_mgr
	inRunner.rmTriggerExecBots(inExecBotId)
	aoFgIndexExecTriggers = inRunner.execBotTriggers
	aoFgIndexTriggerExecs = inRunner.triggerExecBots
	if !reflect.DeepEqual(eoFgIndexExecTriggers, aoFgIndexExecTriggers) {
		t.Errorf("want %v, actual %v", eoFgIndexExecTriggers, aoFgIndexExecTriggers)
	}
	if !reflect.DeepEqual(eoFgIndexTriggerExecs, aoFgIndexTriggerExecs) {
		t.Errorf("want %v, actual %v", eoFgIndexTriggerExecs, aoFgIndexTriggerExecs)
	}
}

func TestRegisterTrigger(t *testing.T) {
	var inExecBot *ExecBot
	var inRunner *BotRunner
	var eoFgIndexExecTriggers, eoFgIndexTriggerExecs map[string][]string
	var aoFgIndexExecTriggers, aoFgIndexTriggerExecs map[string][]string

	act2 := &header.Action{
		Id: "act2",
	}
	cond := &header.Condition{
		Group: header.Condition_all.String(),
		Conditions: []*header.Condition{
			{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
			{Key: "created", Type: "number", Group: header.Condition_single.String()},
		},
	}
	act1 := &header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{Action: act2, Condition: cond},
		},
	}

	eoFgIndexExecTriggers = map[string][]string{"execbot1": {"acc123.obj.obj123.data.content.url", "acc123.obj.obj123.created"}}
	eoFgIndexTriggerExecs = map[string][]string{"acc123.obj.obj123.data.content.url": {"execbot1"}, "acc123.obj.obj123.created": {"execbot1"}}
	inExecBot = &ExecBot{
		AccountId: "acc123",
		ObjId:     "obj123",
		Id:        "execbot1",
		Bot:       &header.Bot{Category: "obj"},
		ActionTree: &ActionNode{
			ActionId: act1.Id,
			action:   act1,
			State:    Node_state_waiting,
			Nexts: []*ActionNode{
				{
					ActionId: act2.Id, action: act2,
				},
			},
			HeadActId: act1.GetId(),
		},
	}
	inRunner = &BotRunner{Mutex: &sync.Mutex{}}
	inRunner.actionMgr = Test_action_mgr
	inRunner.triggerExecBots = make(map[string][]string)
	inRunner.execBotTriggers = make(map[string][]string)
	inRunner.addTriggerExecBots(inExecBot)
	aoFgIndexExecTriggers = inRunner.execBotTriggers
	aoFgIndexTriggerExecs = inRunner.triggerExecBots
	if !reflect.DeepEqual(eoFgIndexExecTriggers, aoFgIndexExecTriggers) {
		t.Errorf("want %v, actual %v", eoFgIndexExecTriggers, aoFgIndexExecTriggers)
	}
	if !reflect.DeepEqual(eoFgIndexTriggerExecs, aoFgIndexTriggerExecs) {
		t.Errorf("want %v, actual %v", eoFgIndexTriggerExecs, aoFgIndexTriggerExecs)
	}
}

func TestLoadBots(t *testing.T) {
	var eoBots, aoBots map[string]*header.Bot
	var eoIndexTriggers, aoIndexTriggers map[string][]string
	var inBots []*header.Bot
	var inRunner *BotRunner

	// const
	bot1 := &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
		Id:        "bot1",
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{Id: "acc123.1"},
					Condition: &header.Condition{
						Key:   "data.content.url",
						Type:  "string",
						Group: header.Condition_single.String(),
					},
				},
			},
		},
	}
	bot2 := &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
		Id:        "bot2",
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{Id: "acc123.1"},
					Condition: &header.Condition{
						Group: header.Condition_all.String(),
						Conditions: []*header.Condition{
							{
								Key:   "data.content.url",
								Type:  "string",
								Group: header.Condition_single.String(),
							},
							{
								Key:   "created",
								Type:  "number",
								Group: header.Condition_single.String(),
							},
						},
					},
				},
			},
		},
	}
	var bot3 *header.Bot

	eoBots = map[string]*header.Bot{bot1.Id: bot1, bot2.Id: bot2}
	eoIndexTriggers = map[string][]string{
		"acc123.users.data.content.url": {"bot1", "bot2"},
		"acc123.users.created":          {"bot2"},
	}
	inBots = []*header.Bot{bot1, bot2, bot3}
	inRunner = &BotRunner{Mutex: &sync.Mutex{}}
	inRunner.actionMgr = Test_action_mgr
	inRunner.logicBots = make(map[string]*LogicBot)
	inRunner.triggerBots = make(map[string][]string)
	inRunner.LoadBots(inBots, nil)
	aoBots = inRunner.GetBotMap()
	aoIndexTriggers = inRunner.triggerBots
	if !reflect.DeepEqual(eoBots, aoBots) {
		t.Errorf("want %v, actual %v", eoBots, aoBots)
	}
	if !reflect.DeepEqual(eoIndexTriggers, aoIndexTriggers) {
		t.Errorf("want %v, actual %v", eoIndexTriggers, aoIndexTriggers)
	}

	eoBots = map[string]*header.Bot{bot1.Id: bot1, bot2.Id: bot2}
	eoIndexTriggers = map[string][]string{
		"acc123.users.data.content.url": {"bot1", "bot2"},
		"acc123.users.created":          {"bot2"},
	}
	bot1i := proto.Clone(bot1).(*header.Bot)
	bot1i.Created = bot1.Created + 1
	bot1i.Action = &header.Action{
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Condition: &header.Condition{
					Group: header.Condition_all.String(),
					Conditions: []*header.Condition{
						{
							Key:   "data.content.url",
							Type:  "string",
							Group: header.Condition_single.String(),
						},
						{
							Key:   "created",
							Type:  "number",
							Group: header.Condition_single.String(),
						},
					},
				},
			},
		},
	}
	inBots = []*header.Bot{bot1i, bot2}
	inRunner = &BotRunner{Mutex: &sync.Mutex{}}
	inRunner.actionMgr = Test_action_mgr
	inRunner.logicBots = map[string]*LogicBot{bot1.Id: {Bot: bot1}}
	inRunner.triggerBots = map[string][]string{
		"acc123.users.data.content.url": {"bot1"},
	}
	inRunner.LoadBots(inBots, nil)
	aoBots = inRunner.GetBotMap()
	aoIndexTriggers = inRunner.triggerBots
	if !reflect.DeepEqual(eoBots, aoBots) {
		t.Errorf("want %v, actual %v", eoBots, aoBots)
	}
	if !reflect.DeepEqual(eoIndexTriggers, aoIndexTriggers) {
		t.Errorf("want %v, actual %v", eoIndexTriggers, aoIndexTriggers)
	}
}

// benchmark test sequence for private func
// benchmark test parallel for public func (use mutex)

// unstable
func BenchmarkRemoveIndex(b *testing.B) {
	strarr := make([]string, 5000)
	for n := 0; n < 1000; n++ {
		removeIndex(strarr, 0)
	}
}

func TestRemoveIndex(t *testing.T) {
	var inArr []string
	var index int
	var exo, aco []string

	inArr = []string{"a"}
	index = 0
	aco = removeIndex(inArr, index)
	exo = []string{}
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}

// 700μs
func BenchmarkRemoveBotFromIndexTriggers(b *testing.B) {
	bot1 := &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
		Id:        "bot1",
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Condition: &header.Condition{
						Group: header.Condition_all.String(),
						Conditions: []*header.Condition{
							{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
							{Key: "created", Type: "string", Group: header.Condition_single.String()},
						},
					},
				},
			},
		},
	}
	inRunner := &BotRunner{
		triggerBots: map[string][]string{
			"acc123.users.data.content.url": {"bot1"},
			"acc123.users.created":          {"bot1", "bot2"},
		},
	}
	inRunner.actionMgr = &TestActionMgr{}
	triggers := []string{"data.content.url", "data.content.path", "data.content.ip", "created", "updated"}
	ids := make([]string, 10000)
	for i := 0; i < len(ids); i++ {
		ids[i] = strconv.Itoa(i)
		inRunner.triggerBots[ids[i]] = triggers
	}
	for _, trigger := range triggers {
		inRunner.triggerBots[trigger] = ids
	}

	for n := 0; n < b.N; n++ {
		inRunner.rmTriggerBots(bot1)
	}
}

// 80ms
func BenchmarkRemoveExecBotFromIndexTriggers(b *testing.B) {
	inRunner := &BotRunner{
		execBotTriggers: map[string][]string{"execbot1": {"data.content.url", "created"}},
		triggerExecBots: map[string][]string{"data.content.url": {"execbot1"}, "created": {"execbot1"}},
	}

	triggers := []string{"data.content.url", "data.content.path", "data.content.ip", "created", "updated"}
	ids := make([]string, 10000)
	for i := 0; i < len(ids); i++ {
		ids[i] = strconv.Itoa(i)
		inRunner.execBotTriggers[ids[i]] = triggers
	}
	for _, trigger := range triggers {
		inRunner.triggerExecBots[trigger] = ids
	}

	for n := 0; n < b.N; n++ {
		inRunner.rmTriggerExecBots("1")
	}
}

func BenchmarkParallelLoadBots(b *testing.B) {
	bot1 := &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
		Id:        "bot1",
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Condition: &header.Condition{
						Key:   "data.content.url",
						Type:  "string",
						Group: header.Condition_single.String(),
					},
				},
			},
		},
	}
	bot2 := &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
		Id:        "bot2",
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Condition: &header.Condition{
						Group: header.Condition_all.String(),
						Conditions: []*header.Condition{
							{
								Key:   "data.content.url",
								Type:  "string",
								Group: header.Condition_single.String(),
							},
							{
								Key:   "created",
								Type:  "number",
								Group: header.Condition_single.String(),
							},
						},
					},
				},
			},
		},
	}
	inBots := []*header.Bot{bot1, bot2}
	inRunner := &BotRunner{Mutex: &sync.Mutex{}}
	inRunner.logicBots = make(map[string]*LogicBot)
	inRunner.triggerBots = make(map[string][]string)

	bots := make([]*header.Bot, 10000)
	for i := 0; i < len(bots); i++ {
		bot := proto.Clone(bot2).(*header.Bot)
		bot.Id = strconv.Itoa(i)
		bots[i] = bot

	}
	inRunner.LoadBots(bots, nil)

	b.SetParallelism(5000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			inRunner.LoadBots(inBots, nil)
		}
	})
}

func BenchmarkOnEvent(b *testing.B) {
	evt1 := &header.Event{
		UserId:    "hasdkasdjkas",
		Id:        "djhfdjshfdjshf",
		AccountId: "dfkfdksjfds",
	}
	var inRunner *BotRunner
	inRunner = &BotRunner{
		Mutex: &sync.Mutex{},
		triggerBots: map[string][]string{
			"acc123.users.data.content.url": {"bot1"},
			"acc123.users.created":          {"bot1", "bot2"},
		},
		execBots: map[string]*ExecBot{
			"execbot1": {
				AccountId: "djfdhjfd",
				Id:        "jshfjdhj",
			},
			"execbot2": {
				AccountId: "thonghoang",
				Id:        "dkfjdskfjkds",
				LastSeen:  10,
			},
		},
	}
	inRunner.actionMgr = Test_action_mgr
	for n := 0; n < b.N; n++ {
		inRunner.OnEvent(&header.RunRequest{
			ObjectType: "users",
			ObjectId:   "obj1",
			Event:      evt1,
		})
	}
}

func TestStartBot(t *testing.T) {
	var inRunner *BotRunner
	var inBot *header.Bot
	var exoTriggerBots map[string][]string
	var acoTriggerBots map[string][]string
	var exoTriggerExecBots map[string][]string
	var acoTriggerExecBots map[string][]string

	inRunner = NewBotRunner("acc123", nil, nil, Test_action_mgr)
	inBot = &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
		Id:        "bot1",
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{Id: "acc123.1"},
					Condition: &header.Condition{
						Key:   "data.content.url",
						Type:  "string",
						Group: header.Condition_single.String(),
					},
				},
			},
		},
	}
	inRunner.NewBot(inBot, nil)
	exoTriggerBots = map[string][]string{"acc123.users.data.content.url": {"bot1"}}
	acoTriggerBots = inRunner.triggerBots
	if !reflect.DeepEqual(acoTriggerBots, exoTriggerBots) {
		t.Errorf("want %#v, actual %#v", exoTriggerBots, acoTriggerBots)
	}

	inRunner.actionMgr = &TestActionMgr{
		Action: &TestAction{
			Wait: func(node *ActionNode) string {
				return Node_state_waiting
			},
		},
	}
	err := inRunner.StartBot(&header.RunRequest{
		BotId:      inBot.GetId(),
		Bot:        inBot,
		ObjectType: pb.Type_user.String(),
		ObjectId:   "testobj",
	})
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(10 * time.Millisecond)
	exoTriggerExecBots = map[string][]string{"acc123.users.testobj.data.content.url": {"testobj.bot1"}}
	acoTriggerExecBots = inRunner.triggerExecBots
	if !reflect.DeepEqual(acoTriggerExecBots, exoTriggerExecBots) {
		t.Errorf("want %#v, actual %#v", exoTriggerExecBots, acoTriggerExecBots)
	}

	inRunner = NewBotRunner("acc123", nil, nil, Test_action_mgr)
	inRunner.actionMgr = &TestActionMgr{
		Action: &TestAction{
			Wait: func(node *ActionNode) string {
				return Node_state_ready
			},
		},
	}
	inRunner.triggerExecBots = map[string][]string{"acc123.users.testobj.data.content.url": {}}
	err = inRunner.StartBot(&header.RunRequest{
		BotId:      inBot.GetId(),
		Bot:        inBot,
		Mode:       "debugging",
		ObjectType: pb.Type_user.String(),
		ObjectId:   "testobj",
	})
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(10 * time.Millisecond)
	exoTriggerExecBots = map[string][]string{"acc123.users.testobj.data.content.url": {}}
	acoTriggerExecBots = inRunner.triggerExecBots
	if !reflect.DeepEqual(acoTriggerExecBots, exoTriggerExecBots) {
		t.Errorf("want %#v, actual %#v", exoTriggerExecBots, acoTriggerExecBots)
	}
}

// TODO replace, ref from bizbot.go
type ActionMgrConv struct {
	caMgr *ca.ActionMgr
}

func (mgr *ActionMgrConv) GetEvtCondTriggers(evt *header.Event) []string {
	return mgr.caMgr.GetEvtCondTriggers(evt)
}

func (mgr *ActionMgrConv) Do(node *ActionNode, req *header.RunRequest) ([]string, error) {
	var crNode interface{}
	crNode = node
	var caNode ca.ActionNode
	var ok bool
	if caNode, ok = crNode.(ca.ActionNode); !ok {
		panic("trigger func connect failure")
	}
	return mgr.caMgr.Do(caNode, req)
}
func (mgr *ActionMgrConv) Wait(node *ActionNode) string {
	var crNode interface{}
	crNode = node
	var caNode ca.ActionNode
	var ok bool
	if caNode, ok = crNode.(ca.ActionNode); !ok {
		panic("trigger func connect failure")
	}
	return mgr.caMgr.Wait(caNode)
}
func (mgr *ActionMgrConv) Triggers(node *ActionNode) []string {
	var crNode interface{}
	crNode = node
	var caNode ca.ActionNode
	var ok bool
	if caNode, ok = crNode.(ca.ActionNode); !ok {
		panic("trigger func connect failure")
	}
	return mgr.caMgr.Triggers(caNode)
}

var Test_action_mgr = &ActionMgrConv{ca.NewActionMgr(nil, nil, nil, nil)}

type SchedTest struct{}

func (sched *SchedTest) Push(schedAt int64, accountId string) bool {
	return true
}

type TestAction struct {
	Do       func(node *ActionNode) ([]string, error) // require
	Wait     func(node *ActionNode) string
	Triggers func(node *ActionNode) []string // belongs to wait
}

type TestActionMgr struct {
	Action *TestAction
}

func (mgr *TestActionMgr) GetEvtCondTriggers(evt *header.Event) []string {
	return Test_action_mgr.GetEvtCondTriggers(evt)
}
func (mgr *TestActionMgr) Do(node *ActionNode, req *header.RunRequest) ([]string, error) {
	nextActions := node.GetAction().GetNexts()
	for _, n := range nextActions {
		if n.GetAction().GetId() != "" {
			return []string{n.GetAction().GetId()}, nil
		}
	}
	return []string{}, nil
}
func (mgr *TestActionMgr) Wait(node *ActionNode) string {
	if mgr.Action != nil {
		return mgr.Action.Wait(node)
	}
	return Test_action_mgr.Wait(node)
}
func (mgr *TestActionMgr) Triggers(node *ActionNode) []string {
	return Test_action_mgr.Triggers(node)
}

type TestExecBotMgr struct{}

func (mgr *TestExecBotMgr) ReadExecBot(botId, id string) (*ExecBot, error) {
	return nil, nil
}
func (mgr *TestExecBotMgr) UpdateExecBot(execBot *ExecBot) (*ExecBot, error) {
	return nil, nil
}
func (mgr *TestExecBotMgr) CreateExecBotIndex(execBotIndex *ExecBotIndex) (*ExecBotIndex, error) {
	return nil, nil
}
func (mgr *TestExecBotMgr) LoadExecBots(accountId string, pks []*ExecBotPK) chan []*ExecBot {
	return nil
}
