package runner

import (
	"github.com/subiz/header"
)

// Problem. Key map
// Event --- Bot Action (compare cond values) (II. Conditional)
// Event --- Bot (list of account id + obj type + cond types) (I. Trigger2ck)
// Event --- ExecBot (list of account id + obj id + cond types) (I. Trigger2ck)
// Event --- ExecBot Action (compare cond values) (II. Conditional)

// Problem. Bot-ExecBot-Actions are waiting what events, times?
// Solution. Do what type of events

// Trigger is relationship between event and bot|execbot
// Trigger type: time(before, after, cron), event type(click, open)
// Trigger on client (leakage) E.g. Exit intent
// Trigger obj: user, system...
// Trigger conditions or not conditional
// Trigger conditions: value, time, with sdk (include cached)

// I. Trigger2ck
// Trigger Id = Account Id + Object Id + Condition Id
// Object is target of bot E.g. users, systems, conversations
// Condition unit is func(field, operator, value). Condition Id = field + operator
//
//                 Account
//                    |
// Event --(all)-- Condition --(k,v)-- Actions - ExecBot - Bot(category)
//       |            |                   |
//     (all) ----- Object ----------- (category)
//
// Expected cost for 1 event is O(n) = k c c 1

// Trigger of exec bot
// I. Trigger2ck
// 1. Object Id define by execBot.Bot.Category.
// e.g. category is user, object id is user id
// 2. Condition Ids define by HeadActs

// bot.GetCategory() is obj type

func TriggerOfHeadNode(execBot *ExecBot, actionMgr ActionMgr) []string {
	node := execBot.ActionTree.HeadActNode()
	if node == nil {
		return Str_empty_arr
	}
	// condition trigger
	keyTriggers := actionMgr.Triggers(node)
	// default trigger, ease
	if len(keyTriggers) == 0 {
		keyTriggers = append(keyTriggers, Trigger_obj)
	}
	return concreteTriggers(execBot.AccountId, execBot.Bot.GetCategory(), execBot.ObjId, keyTriggers)
}

// TODO replace actionMgr by triggerFunc
func TriggerOfExecBot(execBot *ExecBot, actionMgr ActionMgr) []string {
	keyTriggers := make([]string, 0)
	isDefault := false
	// condition trigger
	headNodes := execBot.ActionTree.HeadActNodes()
	for _, node := range headNodes {
		condTriggers := actionMgr.Triggers(node)
		if len(condTriggers) == 0 {
			isDefault = true
			continue
		}
		keyTriggers = append(keyTriggers, condTriggers...)
	}
	// default trigger, ease
	if isDefault {
		keyTriggers = append(keyTriggers, Trigger_obj)
	}
	return concreteTriggers(execBot.AccountId, execBot.Bot.GetCategory(), execBot.ObjId, keyTriggers)
}

func TriggerOfActionNode(node *ActionNode, actionMgr ActionMgr) []string {
	execBot := node.GetExecBot()
	// condition trigger
	keyTriggers := actionMgr.Triggers(node)

	// TODO consistency cat
	return concreteTriggers(execBot.AccountId, execBot.Bot.GetCategory(), execBot.ObjId, keyTriggers)
}

func TriggerOfBot(bot *header.Bot, actionMgr ActionMgr) []string {
	keyTriggers := make([]string, 0)
	// bot trigger type
	for _, trigger := range bot.GetTriggers() {
		keyTriggers = append(keyTriggers, trigger.GetType())
	}
	// condition action trigger
	keyTriggers = append(keyTriggers, actionMgr.Triggers(NewActionNode(bot.Action))...)
	// default trigger, ease
	if len(keyTriggers) == 0 {
		keyTriggers = append(keyTriggers, Trigger_obj)
	}
	return abstractTriggers(bot.GetAccountId(), bot.GetCategory(), keyTriggers)
}

// run request params
func RunCondition(bot *header.Bot, req *header.RunRequest) bool {
	conds := bot.GetConditions()
	if len(conds) == 0 {
		return true
	}

	for _, cond := range conds {
		if checkCondition(cond) {
			return true
		}
	}

	return false
}

// TODO2 bot condition
func checkCondition(*header.BotCondition) bool {
	return true
}

// Default is foreground, exec bot of user,
// sync with trigger of exec bot
// If is background job, dont include user id
// I. Trigger2ck
// 1. Object Ids define by all category constants
// 2. Condition Ids define by all exist fields x type operators
// exist field if value != nil
// e.g. int operators eq, neq, gt, gte, lt, lte

func TriggerOnEvent(objType, objId string, evt *header.Event, actionMgr ActionMgr) []string {
	if evt == nil {
		return Str_empty_arr
	}
	if evt.GetAccountId() == "" {
		return Str_empty_arr
	}

	keyTriggers := make([]string, 0)
	// event trigger type - bot trigger type
	if evt.GetType() != "" {
		keyTriggers = append(keyTriggers, evt.GetType())
	}

	// condition trigger
	keyTriggers = append(keyTriggers, actionMgr.GetEvtCondTriggers(evt)...)
	// default trigger, ease
	keyTriggers = append(keyTriggers, Trigger_obj)

	accountId := evt.GetAccountId()
	return concreteTriggers(accountId, objType, objId, keyTriggers)
}

// TODO rename
func TriggerOnBotEvent(req *header.RunRequest) []string {
	if req.GetAccountId() == "" {
		return Str_empty_arr
	}
	if req.GetObjectType() == "" {
		return Str_empty_arr
	}

	evt := req.GetEvent()
	keyTriggers := make([]string, 0)
	// event trigger type - bot trigger type
	if evt != nil && evt.GetType() != "" {
		keyTriggers = append(keyTriggers, evt.GetType())
	}
	if req.GetBotTriggerType() != "" {
		keyTriggers = append(keyTriggers, req.GetBotTriggerType())
	}
	return abstractTriggers(req.GetAccountId(), req.GetObjectType(), keyTriggers)
}

// require belongs to account, cat
func abstractTriggers(accid, cat string, keys []string) []string {
	triggers := make([]string, len(keys))
	for i, key := range keys {
		triggers[i] = accid + "." + cat + "." + key
	}
	return triggers
}

// require belongs to account, cat, obj
func concreteTriggers(accid, cat, objid string, keys []string) []string {
	triggers := make([]string, len(keys))
	for i, key := range keys {
		triggers[i] = accid + "." + cat + "." + objid + "." + key
	}
	return triggers
}

const (
	Trigger_obj = "_"
)

var Str_empty_arr = []string{}
