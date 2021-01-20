package action

import (
	"reflect"
	"testing"
	"time"

	cr "git.subiz.net/bizbot/runner"
	"github.com/subiz/header"
	"github.com/subiz/header/common"
	pb "github.com/subiz/header/common"
)

// basic case, exception case, fixed case
// input, expected output, actual output

func TestAction_ask_questionDo(t *testing.T) {
	mgr := NewActionMgr(nil, &TestUserMgrClient{}, nil, nil)
	node := cr.NewActionNode(&header.Action{Id: "acttest"})
	node.Reboot(&cr.ExecBot{Bot: &header.Bot{AccountId: "testacc", Id: "bot1", Action: &header.Action{
		Id: "acttest",
		AskQuestion: &header.ActionAskQuestion{
			SkipIfAttributeAlreadyExisted: true,
			SaveToAttribute:               "phone",
		},
		Nexts: []*header.NextAction{
			{Action: &header.Action{Id: "next1"}},
		},
	}}})
	req := &header.RunRequest{}
	ids, _ := Action_ask_question.Do(mgr, node, req)
	exo := []string{"next1"}
	aco := ids
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}

func TestGetCondTriggers(t *testing.T) {
	var eo, ao []string
	var in *header.Condition

	eo = []string{"data.content.url"}
	in = &header.Condition{
		Key:   "data.content.url",
		Type:  "string",
		Group: header.Condition_single.String(),
	}
	ao = getCondTriggers(in)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{"data.content.url", "created"}
	in = &header.Condition{
		Group: header.Condition_all.String(),
		Conditions: []*header.Condition{
			{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
			{Key: "created", Type: "number", Group: header.Condition_single.String()},
		},
	}
	ao = getCondTriggers(in)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{"data.content.url", "created"}
	in = &header.Condition{
		Group: header.Condition_any.String(),
		Conditions: []*header.Condition{
			{Key: "data.content.url", Type: "string", Group: header.Condition_single.String()},
			{Key: "created", Type: "number", Group: header.Condition_single.String()},
		},
	}
	ao = getCondTriggers(in)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{"data.content.url"}
	in = &header.Condition{
		Key:  "data.content.url",
		Type: "string",
	}
	ao = getCondTriggers(in)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
}

func TestAction_conditionDo(t *testing.T) {
	var eo, ao []string
	var inNode *cr.ActionNode

	// const
	created := time.Now().UnixNano() / 1e6
	act1 := &header.Action{Id: "act1"}
	act2 := &header.Action{Id: "act2"}
	act3 := &header.Action{Id: "act3"}

	inNode = cr.NewActionNode(&header.Action{
		Id:   "act",
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Action: act1,
				Condition: &header.Condition{
					Group:  header.Condition_single.String(),
					Key:    "created",
					Type:   "number",
					Number: &common.NumberParams{Eq: float32(created)},
				},
			},
			{
				Action: act2,
				Condition: &header.Condition{
					Group:  header.Condition_single.String(),
					Key:    "created",
					Type:   "number",
					Number: &common.NumberParams{Eq: float32(created)},
				},
			},
			{
				Action: act3,
				Condition: &header.Condition{
					Group:  header.Condition_single.String(),
					Key:    "created",
					Type:   "number",
					Number: &common.NumberParams{Neq: float32(created)},
				},
			},
		},
	})
	inNode.State = Node_state_waiting
	inNode.RunRequests = []*header.RunRequest{
		{Event: &header.Event{Id: "evt1", Created: created}},
		{Event: &header.Event{Id: "evt2", Created: created}},
	}
	inNode.RunTimes = []*cr.RunTime{{ReqIndex: 0}, {ReqIndex: 1}}

	eo = []string{act1.GetId(), act2.GetId()}
	ao, _ = Action_condition.Do(nil, inNode, nil)
	if !reflect.DeepEqual(ao, eo) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	actn1 := &header.Action{
		Id: "acc1.2.1",
		Nexts: []*header.NextAction{
			{Action: &header.Action{Id: "acc1.2.1.1"}},
			{Action: &header.Action{Id: "acc1.2.1.2"}},
			{Action: &header.Action{Id: "acc1.2.1.3"}},
		},
	}
	actn2 := &header.Action{
		Id: "acc1.2.2",
		Nexts: []*header.NextAction{
			{Action: &header.Action{Id: "acc1.2.2.1"}},
			{Action: &header.Action{Id: "acc1.2.2.2"}},
			{Action: &header.Action{Id: "acc1.2.2.3"}},
		},
	}
	inNode = cr.NewActionNode(&header.Action{
		Id:   "act",
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Action: actn1,
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
				Action: actn2,
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
	})
	inNode.State = Node_state_waiting
	inNode.RunRequests = []*header.RunRequest{
		{Event: &header.Event{Id: "evt1", Created: created}},
		{Event: &header.Event{Id: "evt2", Created: created}},
	}
	inNode.RunTimes = []*cr.RunTime{{ReqIndex: 0}, {ReqIndex: 1}}

	eo = []string{actn2.GetId()}
	ao, _ = Action_condition.Do(nil, inNode, nil)
	if !reflect.DeepEqual(ao, eo) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
}

func TestAction_conditionWait(t *testing.T) {
	var eoState, aoState string
	var inNode *cr.ActionNode

	// const
	created := time.Now().UnixNano() / 1e6

	eoState = Node_state_ready
	inNode = cr.NewActionNode(&header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{Id: "act2"},
				Condition: &header.Condition{
					Group:  header.Condition_single.String(),
					Key:    "created",
					Type:   "number",
					Number: &common.NumberParams{Eq: float32(created)},
				},
			},
		},
	})
	inNode.RunRequests = []*header.RunRequest{
		nil,
		{Event: &header.Event{Id: "evt1", Created: created}},
		{Event: &header.Event{Id: "evt2", Created: created}},
	}
	inNode.RunTimes = []*cr.RunTime{{ReqIndex: 1}, {ReqIndex: 2}}
	aoState = Action_condition.Wait(inNode)
	if aoState != eoState {
		t.Errorf("want %v, actual %v", eoState, aoState)
	}

	eoState = Node_state_ready
	inNode = cr.NewActionNode(&header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{Id: "act2"},
				Condition: &header.Condition{
					Group:  header.Condition_single.String(),
					Key:    "created",
					Type:   "number",
					Number: &common.NumberParams{Neq: float32(created)},
				},
			},
		},
	})
	inNode.RunRequests = []*header.RunRequest{
		nil,
		{Event: &header.Event{Id: "evt1", Created: created}},
		{Event: &header.Event{Id: "evt2", Created: created}},
	}
	inNode.RunTimes = []*cr.RunTime{{ReqIndex: 1}, {ReqIndex: 2}}
	aoState = Action_condition.Wait(inNode)
	if aoState != eoState {
		t.Errorf("want %v, actual %v", eoState, aoState)
	}
}

func TestAction_conditionGetTriggers(t *testing.T) {
	var eo, ao []string
	var inNode *cr.ActionNode

	eo = []string{"data.content.url"}
	inNode, _ = cr.MakeTree(&header.Action{
		Id:   "act1",
		Type: "condition",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{Id: "act2"},
				Condition: &header.Condition{
					Key:   "data.content.url",
					Type:  "string",
					Group: header.Condition_single.String(),
				},
			},
		},
	})
	ao = Action_condition.Triggers(inNode)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{"data.content.url", "created"}
	inNode, _ = cr.MakeTree(&header.Action{
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
	ao = Action_condition.Triggers(inNode)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{"data.content.url", "created"}
	inNode, _ = cr.MakeTree(&header.Action{
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
	ao = Action_condition.Triggers(inNode)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	eo = []string{"data.content.url"}
	inNode, _ = cr.MakeTree(&header.Action{
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
	ao = Action_condition.Triggers(inNode)
	if !reflect.DeepEqual(eo, ao) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
}

func TestCompareNode(t *testing.T) {
	var exo, aco bool
	var inCond *header.Condition
	var inNode *cr.ActionNode

	inCond = &header.Condition{
		Group: header.Condition_single.String(),
		Type:  "string",
		Key:   "last_response",
		String_: &common.StringParams{
			In: []string{"no"},
		},
	}
	inNode = &cr.ActionNode{
		RunRequests: []*header.RunRequest{
			nil,
			{
				Event: &header.Event{
					By:      &common.By{Type: "user"},
					Created: 1,
					Data: &header.Event_Data{
						Message: &header.Message{Text: "no"},
					},
				},
			},
			{
				Event: &header.Event{
					By:      &common.By{Type: "user"},
					Created: 2,
					Data: &header.Event_Data{
						Message: &header.Message{Text: "test2"},
					},
				},
			},
		},
		RunTimes: []*cr.RunTime{{ReqIndex: 1}, {ReqIndex: 2}},
	}
	exo = false
	aco = compareNode(inCond, inNode, nil)
	if exo != aco {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inCond = &header.Condition{
		Group: header.Condition_single.String(),
		Type:  "string",
		Key:   "last_response",
		String_: &common.StringParams{
			In: []string{"no"},
		},
	}
	inNode = &cr.ActionNode{
		RunRequests: []*header.RunRequest{
			nil,
			{
				Event: &header.Event{
					Type:    header.RealtimeType_message_sent.String(),
					By:      &common.By{Type: "user"},
					Created: 1,
					Data: &header.Event_Data{
						Message: &header.Message{Text: "test1"},
					},
				},
			},
			{
				Event: &header.Event{
					Type:    header.RealtimeType_message_sent.String(),
					By:      &common.By{Type: "user"},
					Created: 2,
					Data: &header.Event_Data{
						Message: &header.Message{Text: "no"},
					},
				},
			},
		},
		RunTimes: []*cr.RunTime{{ReqIndex: 1}, {ReqIndex: 2}},
	}
	exo = true
	aco = compareNode(inCond, inNode, nil)
	if exo != aco {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inCond = &header.Condition{
		Group: header.Condition_single.String(),
		Type:  "string",
		Key:   "last_response",
		String_: &common.StringParams{
			In: []string{"no"},
		},
	}
	inNode = &cr.ActionNode{
		RunRequests: []*header.RunRequest{
			nil,
			{
				Event: &header.Event{
					By:      &common.By{Type: "user"},
					Created: 1,
					Data: &header.Event_Data{
						Message: &header.Message{Text: "test1"},
					},
				},
			},
			{
				Event: &header.Event{
					By:      &common.By{Type: "user"},
					Created: 2,
					Data: &header.Event_Data{
						Message: &header.Message{Text: "no"},
					},
				},
			},
		},
		RunTimes: []*cr.RunTime{{ReqIndex: 1}, {ReqIndex: 2}},
	}
	exo = false
	aco = compareNode(inCond, inNode, []*InCondition{{Type: "string", Key: "last_response", Value: "test"}})
	if exo != aco {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inCond = &header.Condition{
		Group:   header.Condition_single.String(),
		Type:    "boolean",
		Key:     "last_response",
		Boolean: &common.BooleanParams{Is: true},
	}
	inNode = &cr.ActionNode{}
	exo = true
	aco = compareNode(inCond, inNode, []*InCondition{{Type: "boolean", Key: "last_response", Value: true}})
	if exo != aco {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inCond = &header.Condition{
		Group:   header.Condition_single.String(),
		Type:    "boolean",
		Key:     "last_response",
		Boolean: &common.BooleanParams{Is: true},
	}
	inNode = &cr.ActionNode{}
	exo = false
	aco = compareNode(inCond, inNode, []*InCondition{{Type: "boolean", Key: "last_response", Value: false}})
	if exo != aco {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inCond = &header.Condition{
		Group:   header.Condition_single.String(),
		Type:    "boolean",
		Key:     "last_response",
		Boolean: &common.BooleanParams{Is: true},
	}
	inNode = &cr.ActionNode{}
	exo = true
	aco = compareNode(inCond, nil, []*InCondition{{Type: "boolean", Key: "last_response", Value: true}})
	if exo != aco {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inCond = &header.Condition{
		Group:   header.Condition_single.String(),
		Type:    "boolean",
		Key:     "last_response",
		Boolean: &common.BooleanParams{Is: false},
	}
	inNode = &cr.ActionNode{}
	exo = false
	aco = compareNode(inCond, nil, []*InCondition{})
	if exo != aco {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}
