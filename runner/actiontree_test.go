package runner

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"google.golang.org/protobuf/proto"
)

// basic case, exception case, fixed case
// input, expected output, actual output

func TestHeadActNode(t *testing.T) {
	var exo, aco string
	var inRoot *ActionNode

	inRoot = &ActionNode{
		ActionId: "root",
		action:   &header.Action{Id: "root"},
		Nexts: []*ActionNode{
			{
				ActionId: "act1",
				action:   &header.Action{Id: "act1"},
				Nexts: []*ActionNode{
					{ActionId: "act1.1", action: &header.Action{Id: "act1.1"}},
					{ActionId: "act1.2", action: &header.Action{Id: "act1.2"}},
				},
			},
			{ActionId: "act2", action: &header.Action{Id: "act2"}},
		},
	}
	inRoot.HeadActId = "act1.1"
	exo = "act1.1"
	aco = inRoot.HeadActNode().GetAction().GetId()
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inRoot = &ActionNode{
		ActionId: "root",
		action:   &header.Action{Id: "root"},
		Nexts: []*ActionNode{
			{
				ActionId: "act1",
				action:   &header.Action{Id: "act1"},
				Nexts: []*ActionNode{
					{ActionId: "act1.1", action: &header.Action{Id: "act1.1"}},
					{ActionId: "act1.2", action: &header.Action{Id: "act1.2"}},
				},
			},
			{ActionId: "act2", action: &header.Action{Id: "act2"}},
		},
	}
	inRoot.HeadActId = ""
	exo = ""
	aco = inRoot.HeadActNode().GetAction().GetId()
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}

func TestRoot(t *testing.T) {
	bot := &header.Bot{
		AccountId: "acqcsmrppbftadjzxnvo",
		Id:        "bbquyjvpcltfndekbd",
		Fullname:  "Hello bot",
		AvatarUrl: "/src/assets/img/bot//bot-1.5747dc7.svg",
		Category:  "conversations",
		State:     "active",
		Created:   1603711578758,
		CreatedBy: "agqcsmrppbfqstjwen",
		Updated:   1603766392087087045,
		UpdatedBy: "agqcsmrppbfqstjwen",
		Triggers:  []*header.Trigger{{Type: "conversation_start"}},
		Channels:  []string{"subiz", "facebook", "zalo"},
		Action: &header.Action{
			Id:   "actionhello1",
			Name: "Gretting",
			Type: "send_message",
			AskQuestion: &header.ActionAskQuestion{
				Messages: []*header.Message{
					{
						Text:       "Hello",
						Format:     "html",
						QuillDelta: "{ \"ops\": [{\"insert\": \"Hello\"}] }",
					},
					{
						Text:       "I am Subot. How can I help you",
						Format:     "html",
						QuillDelta: "{ \"ops\": [{\"insert\": \"I am Subot. How can I help you\"}] }",
					},
				},
			},
			Nexts: []*header.NextAction{
				{
					Action: &header.Action{
						Id:   "baltqboiwtegmvwdbc",
						Name: "ask email",
						Type: "ask_question",
						AskQuestion: &header.ActionAskQuestion{
							AllowOpenResponse:             true,
							SaveToAttribute:               "emails",
							Validation:                    "email",
							SkipIfAttributeAlreadyExisted: true,
							Messages: []*header.Message{
								{
									Text:       "\u003cp\u003eHey What is your email?\u003c/p\u003e",
									Format:     "markdown",
									QuillDelta: "{\"ops\":[{\"insert\":\"Hey What is your email?\\n\"}]}",
								},
							},
						},
						Nexts: []*header.NextAction{
							{
								Condition: &header.Condition{},
								Action: &header.Action{
									Id:   "bajdtnkarvvkbcqgbe",
									Name: "ask phone",
									Type: "ask_question",
									AskQuestion: &header.ActionAskQuestion{
										AllowOpenResponse:             true,
										SaveToAttribute:               "phones",
										Validation:                    "phone",
										SkipIfAttributeAlreadyExisted: true,
										Messages: []*header.Message{
											{
												Text:       "What is your phone number?",
												Format:     "markdown",
												QuillDelta: "{ \"ops\": [{\"insert\": \"What is your phone number?\"}] }",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	ao, _ := MakeTree(bot.Action)

	n := ao.Find("bajdtnkarvvkbcqgbe")
	root := n.Root()
	if root == nil {
		t.Error("root loop")
	}
}

func TestMakeTree(t *testing.T) {
	var eo, ao *ActionNode
	var in *header.Action

	in = &header.Action{
		Id: "1",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{
					Id: "1_1",
					Nexts: []*header.NextAction{
						{Action: &header.Action{Id: "1_1_1"}},
						{Action: &header.Action{Id: "1_1_2"}},
						{Action: &header.Action{Id: "1_1_3"}},
					},
				},
			},
			{
				Action: &header.Action{Id: "1_2"},
			},
		},
	}
	node1 := &ActionNode{State: Node_state_new, action: &header.Action{Id: "1"}}
	node1_1 := &ActionNode{State: Node_state_new, action: &header.Action{Id: "1_1"}}
	node1_1_1 := &ActionNode{State: Node_state_new, action: &header.Action{Id: "1_1_1"}}
	node1_1_2 := &ActionNode{State: Node_state_new, action: &header.Action{Id: "1_1_2"}}
	node1_1_3 := &ActionNode{State: Node_state_new, action: &header.Action{Id: "1_1_3"}}
	node1_2 := &ActionNode{State: Node_state_new, action: &header.Action{Id: "1_2"}}

	node1.Nexts = []*ActionNode{node1_1, node1_2}
	node1_1.prev = node1
	node1_2.prev = node1
	node1_1.Nexts = []*ActionNode{node1_1_1, node1_1_2, node1_1_3}
	node1_1_1.prev = node1_1
	node1_1_2.prev = node1_1
	node1_1_3.prev = node1_1

	eo = node1
	ao, _ = MakeTree(in)
	if !compareTree(ao, eo) {
		t.Errorf("want %v, actual %v", eo, ao)
	}

	now := time.Now().UnixNano() / 1e6
	bot := &header.Bot{
		AccountId: "testacc",
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
	eo = &ActionNode{
		State: Node_state_new,
		action: &header.Action{
			Id:   "acc1",
			Type: header.ActionType_send_message.String(),
			AskQuestion: &header.ActionAskQuestion{
				Messages: []*header.Message{
					{Text: "hello, i am x"},
				},
			},
		},
		Nexts: []*ActionNode{
			{
				State: Node_state_new,
				action: &header.Action{
					Id:    "acc1.1",
					Type:  header.ActionType_sleep.String(),
					Sleep: &header.ActionSleep{Duration: 5},
				},
				Nexts: []*ActionNode{
					{State: Node_state_new, action: &header.Action{Id: "acc1.1.1"}},
					{State: Node_state_new, action: &header.Action{Id: "acc1.1.2"}},
					{State: Node_state_new, action: &header.Action{Id: "acc1.1.3"}},
				},
			},
			{
				State: Node_state_new,
				action: &header.Action{
					Id:   "acc1.3",
					Type: header.ActionType_jump.String(),
					Jump: &header.ActionJump{
						ActionId: "acc1.1",
					},
				},
			},
			{
				State: Node_state_new,
				action: &header.Action{
					Id:   "acc1.2",
					Type: header.ActionType_condition.String(),
				},
				Nexts: []*ActionNode{
					{
						State:  Node_state_new,
						action: &header.Action{Id: "acc1.2.1"},
						Nexts: []*ActionNode{
							{State: Node_state_new, action: &header.Action{Id: "acc1.2.1.1"}},
							{State: Node_state_new, action: &header.Action{Id: "acc1.2.1.2"}},
							{State: Node_state_new, action: &header.Action{Id: "acc1.2.1.3"}},
						},
					},
					{
						State:  Node_state_new,
						action: &header.Action{Id: "acc1.2.2"},
						Nexts: []*ActionNode{
							{State: Node_state_new, action: &header.Action{Id: "acc1.2.2.1"}},
							{State: Node_state_new, action: &header.Action{Id: "acc1.2.2.2"}},
							{State: Node_state_new, action: &header.Action{Id: "acc1.2.2.3"}},
						},
					},
				},
			},
		},
	}
	eoAct := proto.Clone(bot.Action)
	ao, _ = MakeTree(bot.Action)
	if !compareTree(ao, eo) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
	if !proto.Equal(bot.Action, eoAct) {
		t.Errorf("actions are muta")
	}
}

func TestTrunkNodes(t *testing.T) {
	var eo, ao []*ActionNode
	var in *ActionNode

	root := &ActionNode{
		action: &header.Action{Id: "root"},
		State:  Node_state_done,
	}
	node1 := &ActionNode{
		action: &header.Action{Id: "node2"},
		State:  Node_state_done,
	}
	node1_1 := &ActionNode{
		State:  Node_state_done,
		action: &header.Action{Id: "node1_1"},
	}
	node1_2 := &ActionNode{
		action: &header.Action{Id: "node1_2"},
	}
	node1_3 := &ActionNode{
		action: &header.Action{Id: "node1_3"},
	}
	node1_4 := &ActionNode{
		action: &header.Action{Id: "node1_4"},
	}
	node1.Nexts = []*ActionNode{node1_1, node1_2, node1_3, node1_4}
	node2 := &ActionNode{
		action: &header.Action{Id: "node2"},
	}
	node3 := &ActionNode{
		action: &header.Action{Id: "node3"},
	}
	node4 := &ActionNode{
		action: &header.Action{Id: "node4"},
	}
	node5 := &ActionNode{
		action: &header.Action{Id: "node5"},
	}
	root.Nexts = []*ActionNode{node1, node2, node3, node4, node5}

	in = root
	eo = []*ActionNode{
		root,
		node1,
		node1_1,
	}
	ao = in.trunkNodes()
	if !compareNodes(ao, eo) {
		t.Errorf("want %v, actual %v", eo, ao)
	}
}

func TestFind(t *testing.T) {
	var eo, ao *ActionNode
	var in *ActionNode

	nodeId := "findme"
	act := &header.Action{
		Id: "1",
		Nexts: []*header.NextAction{
			{
				Action: &header.Action{
					Id: "1_1",
					Nexts: []*header.NextAction{
						{Action: &header.Action{Id: "1_1_1"}},
						{Action: &header.Action{Id: "1_1_2"}},
						{Action: &header.Action{Id: nodeId}},
					},
				},
			},
			{
				Action: &header.Action{Id: "1_2"},
			},
		},
	}
	in, _ = MakeTree(act)
	eo = &ActionNode{
		action: &header.Action{Id: nodeId},
	}
	ao = in.Find(nodeId)
	if ao.action.Id != eo.action.Id {
		t.Errorf("want %v, actual %v", eo.action.Id, ao.action.Id)
	}
}

func compareNodes(srcNext, dstNext []*ActionNode) bool {
	srcNextIsNil := false
	dstNextIsNil := false
	if srcNext == nil || len(srcNext) == 0 {
		srcNextIsNil = true
	}
	if dstNext == nil || len(dstNext) == 0 {
		dstNextIsNil = true
	}
	if srcNextIsNil != dstNextIsNil {
		return false
	}
	if !srcNextIsNil && !dstNextIsNil {
		if len(srcNext) != len(dstNext) {
			return false
		}
		sort.Slice(srcNext, func(i, j int) bool { return srcNext[i].action.GetId() < srcNext[j].action.GetId() })
		sort.Slice(dstNext, func(i, j int) bool { return dstNext[i].action.GetId() < dstNext[j].action.GetId() })
		for i := 0; i < len(srcNext) && i < len(dstNext); i++ {
			if !compareNode(srcNext[i], dstNext[i]) {
				return false
			}
		}
	}

	return true
}

func compareTree(src, dst *ActionNode) bool {
	if !compareNode(src, dst) {
		return false
	}
	srcNext := src.Nexts
	dstNext := dst.Nexts
	srcNextIsNil := false
	dstNextIsNil := false
	if srcNext == nil || len(srcNext) == 0 {
		srcNextIsNil = true
	}
	if dstNext == nil || len(dstNext) == 0 {
		dstNextIsNil = true
	}
	if srcNextIsNil != dstNextIsNil {
		return false
	}
	if !srcNextIsNil && !dstNextIsNil {
		if len(srcNext) != len(dstNext) {
			return false
		}
		sort.Slice(srcNext, func(i, j int) bool { return strings.Compare(srcNext[i].action.GetId(), srcNext[j].action.GetId()) < 0 })
		sort.Slice(dstNext, func(i, j int) bool { return strings.Compare(dstNext[i].action.GetId(), dstNext[j].action.GetId()) < 0 })
		for i := 0; i < len(srcNext) && i < len(dstNext); i++ {
			if !compareTree(srcNext[i], dstNext[i]) {
				return false
			}
		}
	}

	return true
}

func compareNode(src, dst *ActionNode) bool {
	if src == nil && dst == nil {
		return true
	}
	if src != nil && dst == nil {
		return false
	}
	if src == nil && dst != nil {
		return false
	}
	if !compareAction(src.action, dst.action) {
		return false
	}
	if src.State != dst.State {
		return false
	}
	return true
}

// TODO reflect.DeepEqual
func compareAction(src, dst *header.Action) bool {
	if src.Id != dst.Id {
		return false
	}
	return true
}
