package runner

import (
	"reflect"
	"testing"
	"time"

	"github.com/subiz/header"
	"github.com/subiz/header/common"
	pb "github.com/subiz/header/common"
)

// basic case, exception case, fixed case
// input, expected output, actual output

func TestTriggerOfExecBot(t *testing.T) {
	var eo, ao []string
	var in *ExecBot
	var inNode *ActionNode

	bot := &header.Bot{
		Category: header.BotCategory_users.String(),
	}

	eo = []string{"acc123.users.obj123.data.content.url"}
	inNode, _ = MakeTree(&header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{Id: "act1.1"},
				Condition: &header.Condition{
					Key:   "data.content.url",
					Type:  "string",
					Group: header.Condition_single.String(),
				},
			},
		},
	})
	in = &ExecBot{
		AccountId:  "acc123",
		ObjId:      "obj123",
		Bot:        bot,
		ActionTree: inNode,
	}
	ao = TriggerOfExecBot(in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	inNode, _ = MakeTree(&header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{Id: "act2"},
				Condition: &header.Condition{
					Group: header.Condition_all.String(),
					Conditions: []*header.Condition{
						{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
						{Key: "created", Type: "number", Group: header.Condition_single.String()},
					},
				},
			},
		},
	})
	eo = []string{"acc123.obj.obj123.data.content.url", "acc123.obj.obj123.created"}
	in = &ExecBot{
		AccountId:  "acc123",
		ObjId:      "obj123",
		ActionTree: inNode,
		Bot:        &header.Bot{Category: "obj"},
	}
	ao = TriggerOfExecBot(in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	inNode, _ = MakeTree(&header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{Id: "act2"},
				Condition: &header.Condition{
					Group: header.Condition_any.String(),
					Conditions: []*header.Condition{
						{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
						{Key: "created", Type: "number", Group: header.Condition_single.String()},
					},
				},
			},
		},
	})
	eo = []string{"acc123.obj.obj123.data.content.url", "acc123.obj.obj123.created"}
	in = &ExecBot{
		AccountId:  "acc123",
		ObjId:      "obj123",
		ActionTree: inNode,
		Bot:        &header.Bot{Category: "obj"},
	}
	ao = TriggerOfExecBot(in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	inNode, _ = MakeTree(&header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{Id: "act2"},
				Condition: &header.Condition{
					Key:  "data.content.url",
					Type: "string",
				},
			},
		},
	})
	eo = []string{"acc123.obj.obj123.data.content.url"}
	in = &ExecBot{
		AccountId:  "acc123",
		ObjId:      "obj123",
		ActionTree: inNode,
		Bot:        &header.Bot{Category: "obj"},
	}
	ao = TriggerOfExecBot(in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
}

func TestTriggerOfBot(t *testing.T) {
	var eo, ao []string
	var in *header.Bot

	eo = []string{"acc123.users.data.content.url"}
	in = &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
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
	ao = TriggerOfBot(in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{"acc123.users.data.content.url", "acc123.users.created"}
	in = &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
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
	ao = TriggerOfBot(in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{"acc123.users.data.content.url", "acc123.users.created"}
	in = &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{Id: "acc123.1"},
					Condition: &header.Condition{
						Group: header.Condition_any.String(),
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
	ao = TriggerOfBot(in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{"acc123.users.data.content.url"}
	in = &header.Bot{
		AccountId: "acc123",
		Category:  header.BotCategory_users.String(),
		Action: &header.Action{
			Type: "condition",
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{Id: "acc123.1"},
					Condition: &header.Condition{
						Key:  "data.content.url",
						Type: "string",
					},
				},
			},
		},
	}
	ao = TriggerOfBot(in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
}

func TestTriggerOnEvent(t *testing.T) {
	var eo, ao []string
	var in *header.Event

	objType := "users"
	objId := "user123"
	// order by EventFunc
	eo = []string{
		"acc123.users.user123.content_viewed",
		"acc123.users.user123.event_exist",
		"acc123.users.user123.created",
		"acc123.users.user123.data.content.url",
		"acc123.users.user123._",
	}
	in = &header.Event{
		AccountId: "acc123",
		Type:      header.RealtimeType_content_viewed.String(),
		Created:   time.Now().UnixNano() / 1e6,
		By: &common.By{
			Type: pb.Type_user.String(),
			Id:   "user123",
		},
	}
	ao = TriggerOnEvent(objType, objId, in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{}
	in = &header.Event{}
	ao = TriggerOnEvent(objType, objId, in, Test_action_mgr)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
}
