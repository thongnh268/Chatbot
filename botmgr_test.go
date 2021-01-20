package main

import (
	"reflect"
	"sort"
	"testing"

	cr "git.subiz.net/bizbot/runner"
	"github.com/subiz/header"
	"google.golang.org/protobuf/proto"
)

func TestHashAction(t *testing.T) {
	var exo, aco string
	inReq := proto.Clone(Test_fixtures.RunRequest6).(*header.RunRequest)
	exo = "4e6ffb08f135018944bf9f019ad7eb16fa0ee1a4"
	inReq.Bot.Action.Type = "typetest"
	aco = hashAction(inReq.GetBot().GetAction())
	if aco != exo {
		t.Errorf("want %s, actual %s", exo, aco)
	}
}

func TestListObjectIds(t *testing.T) {
	var exo, aco []*header.ListObjectId
	mgr := &ExecBotMgr{db: &DBTest{}}
	exo = []*header.ListObjectId{
		{
			ObjectType: "users",
			ObjectIds:  []string{"users.2", "users.1"},
		},
		{
			ObjectType: "conversations",
			ObjectIds:  []string{"conversations.2", "conversations.1"},
		},
	}
	aco = mgr.ListObjectIds("", 1, 2)
	if len(aco) != len(exo) {
		t.Error("report go wrong")
		return
	}
	for i := 0; i < len(exo); i++ {
		if !proto.Equal(aco[i], exo[i]) {
			t.Error("report go wrong")
		}
	}
}

func TestAggregateExecBotIndex(t *testing.T) {
	var exo, aco []*header.Metric

	mgr := &ExecBotMgr{db: &DBTest{}}
	exo = []*header.Metric{
		{
			Submetrics: []*header.Metric{
				{ObjectType: "users", ObjectCount: 2, Count: 12},
				{ObjectType: "conversations", ObjectCount: 2, Count: 12},
			},
			DateDim:     2,
			Count:       24,
			ObjectCount: 4,
		},
		{
			Submetrics: []*header.Metric{
				{ObjectType: "users", ObjectCount: 2, Count: 12, LeadCount: 1},
				{ObjectType: "conversations", ObjectCount: 2, Count: 12},
			},
			DateDim:     1,
			ObjectCount: 4,
			Count:       24,
			LeadCount:   1,
		},
	}
	aco = mgr.AggregateExecBotIndex("", 1, 2)
	if len(aco) != len(exo) {
		t.Error("report go wrong")
		return
	}
	for i := 0; i < len(exo); i++ {
		if !proto.Equal(aco[i], exo[i]) {
			t.Error("report go wrong")
		}
	}
}

type DBTest struct{}

func (db *DBTest) LoadExecBotIndexs(accountId, botId string, beginTimeDayFrom, beginTimeDayTo int64, orderDesc []string, c chan []*cr.ExecBotIndex, size int) error {
	defer close(c)
	arr := []*cr.ExecBotIndex{
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 1, LeadId: "users.1", LeadPayload: map[string]string{"email": "test@subiz.com"}},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 2},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 3},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 1},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 2},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 3},

		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 1},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 2},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 3},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 1},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 2},
		{BeginTimeDay: 1, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 3},

		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 1},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 2},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 3},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 1},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 2},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 3},

		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 1},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 2},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 3},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 1},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 2},
		{BeginTimeDay: 1, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 3},

		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 1},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 2},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 3},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 1},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 2},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.1", Id: "users.1", IndexIncre: 3},

		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 1},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 2},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 3},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 1},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 2},
		{BeginTimeDay: 2, ObjType: "users", ObjId: "users.2", Id: "users.2", IndexIncre: 3},

		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 1},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 2},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 3},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 1},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 2},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.1", Id: "conversations.1", IndexIncre: 3},

		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 1},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 2},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 3},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 1},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 2},
		{BeginTimeDay: 2, ObjType: "conversations", ObjId: "conversations.2", Id: "conversations.2", IndexIncre: 3},
	}

	if reflect.DeepEqual([]string{"id"}, orderDesc) {
		sort.Slice(arr, func(i, j int) bool { return arr[i].Id > arr[j].Id })
	}

	if reflect.DeepEqual([]string{"begin_time_day", "id"}, orderDesc) {
		sort.Slice(arr, func(i, j int) bool {
			if arr[i].BeginTimeDay == arr[j].BeginTimeDay {
				return arr[i].Id > arr[j].Id
			}
			return arr[i].BeginTimeDay > arr[j].BeginTimeDay
		})
	}

	c <- arr
	return nil
}

func (db *DBTest) UpsertExecBot(*cr.ExecBot) error {
	return nil
}
func (db *DBTest) ReadExecBot(accid, botid, id string) (*cr.ExecBot, error) {
	return nil, nil
}
func (db *DBTest) ReadExecBots(accid, botid string) ([]*cr.ExecBot, error) {
	return nil, nil
}
func (db *DBTest) DeleteExecBot(accid, botid, id string) error {
	return nil
}
func (db *DBTest) UpsertExecBotIndex(execBotIndex *cr.ExecBotIndex) error {
	return nil
}
func (db *DBTest) LastExecBotIndex(accountId, botId, id string) (int64, error) {
	return 0, nil
}
func (db *DBTest) LoadExecBots(accountId string, c chan []*cr.ExecBot, size int, botPKMap map[string][]*cr.ExecBotPK) error {
	return nil
}
