package main

import (
	"encoding/json"
	"fmt"
	"strings"

	cr "git.subiz.net/bizbot/runner"
	"github.com/dgraph-io/ristretto"
	"github.com/gocql/gocql"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/cassandra"
	"github.com/subiz/errors"
	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
)

// TODO1 integration test
type DB struct {
	session *gocql.Session
	cql     *cassandra.Query
	cache   *ristretto.Cache
}

func NewDB(seeds []string, keyspace string) *DB {
	db := &DB{}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e3, // number of keys to track frequency of (1k).
		MaxCost:     1e8, // maximum cost of cache (100MB).
		BufferItems: 64,  // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
	db.cache = cache

	db.cql = &cassandra.Query{}
	if err := db.cql.Connect(seeds, keyspace); err != nil {
		panic(err)
	}
	db.session = db.cql.Session

	return db
}

func (db *DB) LoadAllBots(c chan []*header.Bot, size int) error {
	defer close(c)

	iter := db.cql.Session.Query(`SELECT bot FROM ` + Bizbot_tbl_bot).Iter()
	var botb []byte
	arr := make([]*header.Bot, size)
	index := 0
	for iter.Scan(&botb) {
		bot := &header.Bot{}
		proto.Unmarshal(botb, bot)
		arr[index] = bot
		index++
		if index == size {
			c <- arr
			arr = make([]*header.Bot, size)
			index = 0
		}
	}

	if index > 0 {
		c <- arr[:index]
	}

	err := iter.Close()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "load all bots")
	}

	return nil
}

func (db *DB) ReadAccountIds() ([]string, error) {
	arr := make([]string, 0)
	iter := db.cql.Session.Query(`
		SELECT
			account_id
		FROM ` + Bizbot_tbl_bot + `
	`).Iter()
	var accountId string
	for iter.Scan(&accountId) {

		arr = append(arr, accountId)
	}
	err := iter.Close()
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error)
	}

	return arr, nil
}

func (db *DB) DeleteBot(accountId, id string) error {
	if err := db.cql.Delete(Bizbot_tbl_bot, header.Bot{AccountId: accountId, Id: id}); err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "delete bot", accountId, id)
	}
	return nil
}

func (db *DB) UpsertBot(b *header.Bot) error {
	botb, _ := proto.Marshal(b)
	err := db.session.Query(
		`INSERT INTO `+Bizbot_tbl_bot+`(account_id,id,fullname,state,bot) VALUES(?,?,?,?,?)`,
		b.AccountId, b.Id, b.Fullname, b.State, botb,
	).Exec()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "upsert bot", b.AccountId, b.Id)
	}

	if _, found := db.cache.Get(b.GetAccountId() + ":" + b.GetId()); found {
		db.cache.Set(b.GetAccountId()+":"+b.GetId(), b, 1024)
	}

	return nil
}

func (db *DB) UpsertBotReport(report *cr.ReportRun) error {
	reportb, _ := json.Marshal(report)
	err := db.session.Query(
		`INSERT INTO `+Bizbot_tbl_bot+`(account_id,id,report) VALUES(?,?,?)`,
		report.AccountId, report.BotId, reportb,
	).Exec()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "upsert bot report", report.AccountId, report.BotId)
	}
	return nil
}

func (db *DB) UpsertBotTag(b *header.Bot, tag string) error {
	botb, _ := proto.Marshal(b)
	err := db.session.Query(`INSERT INTO `+Bizbot_tbl_bot_tag+`(account_id,id,tag,fullname,state,bot) VALUES(?,?,?,?,?,?)`,
		b.AccountId,
		b.Id,
		tag,
		b.Fullname,
		b.State,
		botb,
	).Exec()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "upsert bot tag", b.AccountId, b.Id, b.ActionHash)
	}

	return nil
}

func (db *DB) ReadBotTag(accountid, id, tag string) (*header.Bot, error) {
	botb := make([]byte, 0)
	err := db.session.Query(`SELECT bot FROM `+Bizbot_tbl_bot_tag+`WHERE account_id=? AND id=? AND tag=?`, accountid, id, tag).Scan(&botb)
	if err != nil {
		if err.Error() == gocql.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, 500, errors.E_database_error, "read tag bot", accountid, id, tag)
		}
	}

	var bot header.Bot
	proto.Unmarshal(botb, &bot)
	return &bot, nil
}

func (db *DB) ReadAllTags(accountid, id string) ([]*header.Bot, error) {
	arr := make([]*header.Bot, 0)
	iter := db.cql.Session.Query(`SELECT bot FROM`+Bizbot_tbl_bot_tag+`WHERE account_id=? AND id=?`, accountid, id).Iter()
	var botb []byte
	for iter.Scan(&botb) {
		bot := &header.Bot{}
		proto.Unmarshal(botb, bot)
		arr = append(arr, bot)
	}
	err := iter.Close()
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, "read all tags", accountid, id)
	}

	return arr, nil
}

func (db *DB) ReadBot(accountId, id string) (*header.Bot, error) {
	if value, found := db.cache.Get(accountId + ":" + id); found {
		return value.(*header.Bot), nil
	}

	botb := make([]byte, 0)
	err := db.session.Query(`SELECT bot FROM `+Bizbot_tbl_bot+` WHERE account_id=? AND id=? LIMIT 1`, accountId, id).Scan(&botb)
	if err != nil {
		if err.Error() == gocql.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, 500, errors.E_database_error, "read bot", accountId, id)
		}
	}

	var bot header.Bot
	proto.Unmarshal(botb, &bot)
	db.cache.Set(accountId+":"+id, &bot, 1024)
	return &bot, nil
}

func (db *DB) ReadBotReport(accountId, id string) (*cr.ReportRun, error) {
	reportb := make([]byte, 0)
	err := db.session.Query(`SELECT report FROM `+Bizbot_tbl_bot+` WHERE account_id=? AND id=? LIMIT 1`, accountId, id).Scan(&reportb)
	if err != nil && err.Error() == gocql.ErrNotFound.Error() {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, "read bot report", accountId, id)
	}
	if len(reportb) == 0 {
		return nil, nil
	}
	var report cr.ReportRun
	json.Unmarshal(reportb, &report)
	return &report, nil
}

func (db *DB) ReadAllBots(accountId string) ([]*header.Bot, error) {
	arr := make([]*header.Bot, 0)
	iter := db.cql.Session.Query(`SELECT bot FROM `+Bizbot_tbl_bot+` WHERE account_id=?`, accountId).Iter()
	var botb []byte
	for iter.Scan(&botb) {
		bot := &header.Bot{}
		proto.Unmarshal(botb, bot)
		arr = append(arr, bot)
	}
	err := iter.Close()
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, "read all bots", accountId)
	}

	return arr, nil
}

func (db *DB) ReadAllBotReports(accountId string) ([]*cr.ReportRun, error) {
	arr := make([]*cr.ReportRun, 0)
	iter := db.cql.Session.Query(`SELECT report FROM `+Bizbot_tbl_bot+` WHERE account_id=?`, accountId).Iter()
	var reportb []byte
	for iter.Scan(&reportb) {
		if len(reportb) == 0 {
			continue
		}
		report := &cr.ReportRun{}
		json.Unmarshal(reportb, report)
		arr = append(arr, report)
	}
	err := iter.Close()
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, "read all bot reports", accountId)
	}

	return arr, nil
}

func (db *DB) DeleteExecBot(accountId, botId, id string) error {
	if err := db.cql.Delete(Bizbot_tbl_exec_bot, cr.ExecBot{AccountId: accountId, BotId: botId, Id: id}); err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "delete exec bot", accountId, botId, id)
	}
	return nil
}

func (db *DB) UpsertExecBot(execBot *cr.ExecBot) error {
	var actionTreeb []byte
	actionTreeb = execBot.ActionTreeb
	if len(actionTreeb) == 0 && execBot.ActionTree != nil {
		actionTreeb, _ = json.Marshal(execBot.ActionTree)
	}
	err := db.session.Query(`
			INSERT INTO `+Bizbot_tbl_exec_bot+`(
				account_id,
				bot_id,
				obj_type,
				obj_id,
				lead_id,
				lead_payload,
				id,
				status,
				status_note,
				run_times,
				created,
				updated,
				last_seen,
				action_tree,
				sched_acts
			)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		`,
		execBot.AccountId,
		execBot.BotId,
		execBot.ObjType,
		execBot.ObjId,
		execBot.LeadId,
		execBot.LeadPayload,
		execBot.Id,
		execBot.Status,
		execBot.StatusNote,
		execBot.RunTimes,
		execBot.Created,
		execBot.Updated,
		execBot.LastSeen,
		actionTreeb,
		execBot.SchedActs,
	).Exec()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "upsert exec bot", execBot.AccountId, execBot.BotId, execBot.Id)
	}

	return nil
}

func (db *DB) ReadExecBot(accountId, botId, id string) (*cr.ExecBot, error) {
	execBot := &cr.ExecBot{AccountId: accountId, BotId: botId, Id: id}
	var actionTreeb []byte
	err := db.session.Query(`
			SELECT
				obj_type,
				obj_id,
				lead_id,
				lead_payload,
				status,
				status_note,
				run_times,
				created,
				updated,
				last_seen,
				action_tree,
				sched_acts
			FROM `+Bizbot_tbl_exec_bot+`
			WHERE account_id=? AND bot_id=? AND id=?
			LIMIT 1
		`,
		accountId, botId, id,
	).Scan(
		&execBot.ObjType,
		&execBot.ObjId,
		&execBot.LeadId,
		&execBot.LeadPayload,
		&execBot.Status,
		&execBot.StatusNote,
		&execBot.RunTimes,
		&execBot.Created,
		&execBot.Updated,
		&execBot.LastSeen,
		&actionTreeb,
		&execBot.SchedActs,
	)
	actionTree := &cr.ActionNode{}
	json.Unmarshal(actionTreeb, actionTree)
	execBot.ActionTree = actionTree
	if err != nil {
		if err.Error() == gocql.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, 500, errors.E_database_error, "read exec bot", accountId, botId, id)
		}
	}

	return execBot, nil
}

func (db *DB) ReadExecBots(accountId, botId string) ([]*cr.ExecBot, error) {
	arr := make([]*cr.ExecBot, 0)
	iter := db.session.Query(`
			SELECT
				obj_type,
				obj_id,
				lead_id,
				lead_payload,
				id,
				status,
				status_note,
				run_times,
				created,
				updated,
				last_seen,
				action_tree,
				sched_acts
			FROM `+Bizbot_tbl_exec_bot+`
			WHERE account_id=? AND bot_id=?
		`,
		accountId, botId,
	).Iter()
	var objType, objId, leadId, id, status, statusNote string
	var leadPayload map[string]string
	var runTimes []int64
	var created, updated, lastSeen int64
	var actionTreeb []byte
	var schedActs map[string]int64
	for iter.Scan(
		&objType,
		&objId,
		&leadId,
		&leadPayload,
		&id,
		&status,
		&statusNote,
		&runTimes,
		&created,
		&updated,
		&lastSeen,
		&actionTreeb,
		&schedActs,
	) {
		execBot := &cr.ExecBot{
			AccountId:   accountId,
			BotId:       botId,
			ObjType:     objType,
			ObjId:       objId,
			LeadId:      leadId,
			LeadPayload: leadPayload,
			Id:          id,
			Status:      status,
			StatusNote:  statusNote,
			RunTimes:    runTimes,
			Created:     created,
			Updated:     updated,
			LastSeen:    lastSeen,
		}
		actionTree := &cr.ActionNode{}
		json.Unmarshal(actionTreeb, actionTree)
		execBot.ActionTree = actionTree

		var schedActsCopy map[string]int64
		if schedActs != nil && len(schedActs) > 0 {
			schedActsCopy = make(map[string]int64, len(schedActs))
			for key, value := range schedActs {
				schedActsCopy[key] = value
			}
		}
		execBot.SchedActs = schedActsCopy

		arr = append(arr, execBot)
	}

	err := iter.Close()
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, "read all bots", accountId)
	}

	return arr, nil
}

func (db *DB) ReadAllExecBots(accountId string) ([]*cr.ExecBot, error) {
	arr := make([]*cr.ExecBot, 0)
	iter := db.session.Query(`
		SELECT
			bot_id,
			obj_type,
			obj_id,
			lead_id,
			lead_payload,
			id,
			status,
			status_note,
			run_times,
			created,
			updated,
			last_seen,
			action_tree,
			sched_acts
		FROM `+Bizbot_tbl_exec_bot+`
		WHERE account_id=?
	`, accountId).Iter()
	var botId, objType, objId, leadId, id, status, statusNote string
	var leadPayload map[string]string
	var runTimes []int64
	var created, updated, lastSeen int64
	var actionTreeb []byte
	var schedActs map[string]int64
	for iter.Scan(
		&botId,
		&objType,
		&objId,
		&leadId,
		&leadPayload,
		&id,
		&status,
		&statusNote,
		&runTimes,
		&created,
		&updated,
		&lastSeen,
		&actionTreeb,
		&schedActs,
	) {
		execBot := &cr.ExecBot{
			AccountId:   accountId,
			BotId:       botId,
			ObjType:     objType,
			ObjId:       objId,
			LeadId:      leadId,
			LeadPayload: leadPayload,
			Id:          id,
			Status:      status,
			StatusNote:  statusNote,
			RunTimes:    runTimes,
			Created:     created,
			Updated:     updated,
			LastSeen:    lastSeen,
		}
		actionTree := &cr.ActionNode{}
		json.Unmarshal(actionTreeb, actionTree)
		execBot.ActionTree = actionTree

		var schedActsCopy map[string]int64
		if schedActs != nil && len(schedActs) > 0 {
			schedActsCopy = make(map[string]int64, len(schedActs))
			for key, value := range schedActs {
				schedActsCopy[key] = value
			}
		}
		execBot.SchedActs = schedActsCopy

		arr = append(arr, execBot)
	}

	err := iter.Close()
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, "read all bots", accountId)
	}

	return arr, nil
}

// TODO only read primary key
func (db *DB) LoadTerminatedExecBots(c chan []*cr.ExecBot, size int) error {
	defer close(c)

	iter := db.session.Query(`
		SELECT
			account_id,
			bot_id,
			obj_type,
			obj_id,
			id,
			status,
			status_note,
			created,
			updated,
			last_seen,
			sched_acts
		FROM ` + Bizbot_tbl_exec_bot + `
	`).Iter()
	var accountId, botId, objType, objId, id, status, statusNote string
	var created, updated, lastSeen int64
	var schedActs map[string]int64
	arr := make([]*cr.ExecBot, size)
	index := 0
	for iter.Scan(
		&accountId,
		&botId,
		&objType,
		&objId,
		&id,
		&status,
		&statusNote,
		&created,
		&updated,
		&lastSeen,
		&schedActs,
	) {
		if status != cr.Exec_bot_state_terminated {
			continue
		}

		execBot := &cr.ExecBot{
			AccountId:  accountId,
			BotId:      botId,
			ObjType:    objType,
			ObjId:      objId,
			Id:         id,
			Status:     status,
			StatusNote: statusNote,
			Created:    created,
			Updated:    updated,
			LastSeen:   lastSeen,
		}

		var schedActsCopy map[string]int64
		if schedActs != nil && len(schedActs) > 0 {
			schedActsCopy = make(map[string]int64, len(schedActs))
			for key, value := range schedActs {
				schedActsCopy[key] = value
			}
		}
		execBot.SchedActs = schedActsCopy

		arr[index] = execBot
		index++
		if index == size {
			c <- arr
			arr = make([]*cr.ExecBot, size)
			index = 0
		}
	}

	if index > 0 {
		c <- arr[:index]
	}

	err := iter.Close()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "read all bots", accountId)
	}

	return nil
}

func (db *DB) LoadExecBots(accountId string, c chan []*cr.ExecBot, size int, botPKMap map[string][]*cr.ExecBotPK) error {
	defer close(c)

	for botId, botPKs := range botPKMap {
		in := ","
		for _, pk := range botPKs {
			in += `'` + pk.Id + `',`
		}
		iter := db.session.Query(`
			SELECT
				bot_id,
				obj_type,
				obj_id,
				lead_id,
				lead_payload,
				id,
				status,
				status_note,
				run_times,
				created,
				updated,
				last_seen,
				action_tree,
				sched_acts
			FROM `+Bizbot_tbl_exec_bot+`
			WHERE account_id=? AND bot_id=? AND id IN(`+strings.Trim(in, ",")+`)
		`, accountId, botId).Iter()
		var botId, objType, objId, leadId, id, status, statusNote string
		var leadPayload map[string]string
		var runTimes []int64
		var created, updated, lastSeen int64
		var actionTreeb []byte
		var schedActs map[string]int64
		arr := make([]*cr.ExecBot, size)
		index := 0
		for iter.Scan(
			&botId,
			&objType,
			&objId,
			&leadId,
			&leadPayload,
			&id,
			&status,
			&statusNote,
			&runTimes,
			&created,
			&updated,
			&lastSeen,
			&actionTreeb,
			&schedActs,
		) {
			execBot := &cr.ExecBot{
				AccountId:   accountId,
				BotId:       botId,
				ObjType:     objType,
				ObjId:       objId,
				LeadId:      leadId,
				LeadPayload: leadPayload,
				Id:          id,
				Status:      status,
				StatusNote:  statusNote,
				RunTimes:    runTimes,
				Created:     created,
				Updated:     updated,
				LastSeen:    lastSeen,
			}

			actionTree := &cr.ActionNode{}
			json.Unmarshal(actionTreeb, actionTree)
			execBot.ActionTree = actionTree

			var schedActsCopy map[string]int64
			if schedActs != nil && len(schedActs) > 0 {
				schedActsCopy = make(map[string]int64, len(schedActs))
				for key, value := range schedActs {
					schedActsCopy[key] = value
				}
			}
			execBot.SchedActs = schedActsCopy

			arr[index] = execBot
			index++
			if index == size {
				c <- arr
				arr = make([]*cr.ExecBot, size)
				index = 0
			}
		}

		if index > 0 {
			c <- arr[:index]
		}

		err := iter.Close()
		if err != nil {
			return errors.Wrap(err, 500, errors.E_database_error, "read all bots", accountId)
		}
	}

	return nil
}

// TODO1
func (db *DB) LoadAllExecBots(accountId string, c chan []*cr.ExecBot, size int, runExecBots map[string]*cr.ExecBot) error {
	defer close(c)

	iter := db.session.Query(`
		SELECT
			bot_id,
			obj_type,
			obj_id,
			lead_id,
			lead_payload,
			id,
			status,
			status_note,
			run_times,
			created,
			updated,
			last_seen,
			action_tree,
			sched_acts
		FROM `+Bizbot_tbl_exec_bot+`
		WHERE account_id=?
	`, accountId).Iter()
	var botId, objType, objId, leadId, id, status, statusNote string
	var leadPayload map[string]string
	var runTimes []int64
	var created, updated, lastSeen int64
	var actionTreeb []byte
	var schedActs map[string]int64
	arr := make([]*cr.ExecBot, size)
	index := 0
	for iter.Scan(
		&botId,
		&objType,
		&objId,
		&leadId,
		&leadPayload,
		&id,
		&status,
		&statusNote,
		&runTimes,
		&created,
		&updated,
		&lastSeen,
		&actionTreeb,
		&schedActs,
	) {
		// ignore nil tree, terminated
		if len(actionTreeb) == 0 {
			continue
		}
		if status == cr.Exec_bot_state_terminated {
			continue
		}
		if runExecBot, has := runExecBots[id]; has {
			log.Debug("load-all-exec-bots", fmt.Sprintf("account-id=%s bot-id=%s", accountId, botId), fmt.Sprintf("run-exec-bot.updated=%d updated=%d", runExecBot.Updated, updated))
			if updated <= runExecBot.Updated {
				continue
			}
		}
		log.Debug("load-all-exec-bots", fmt.Sprintf("account-id=%s bot-id=%s", accountId, botId), fmt.Sprintf("status=%s", status))

		execBot := &cr.ExecBot{
			AccountId:   accountId,
			BotId:       botId,
			ObjType:     objType,
			ObjId:       objId,
			LeadId:      leadId,
			LeadPayload: leadPayload,
			Id:          id,
			Status:      status,
			StatusNote:  statusNote,
			RunTimes:    runTimes,
			Created:     created,
			Updated:     updated,
			LastSeen:    lastSeen,
		}

		actionTree := &cr.ActionNode{}
		json.Unmarshal(actionTreeb, actionTree)
		execBot.ActionTree = actionTree

		var schedActsCopy map[string]int64
		if schedActs != nil && len(schedActs) > 0 {
			schedActsCopy = make(map[string]int64, len(schedActs))
			for key, value := range schedActs {
				schedActsCopy[key] = value
			}
		}
		execBot.SchedActs = schedActsCopy

		arr[index] = execBot
		index++
		if index == size {
			c <- arr
			arr = make([]*cr.ExecBot, size)
			index = 0
		}
	}

	if index > 0 {
		c <- arr[:index]
	}

	err := iter.Close()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "read all bots", accountId)
	}

	return nil
}

func (db *DB) UpsertExecBotIndex(execBotIndex *cr.ExecBotIndex) error {
	execBotb, _ := json.Marshal(execBotIndex.ExecBot)
	err := db.session.Query(`
			INSERT INTO `+Bizbot_tbl_exec_bot_index+`(
				account_id,
				bot_id,
				obj_type,
				obj_id,
				lead_id,
				lead_payload,
				id,
				index_incre,
				begin_time_day,
				begin_time,
				end_time,
				created,
				updated,
				exec_bot
			)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		`,
		execBotIndex.AccountId,
		execBotIndex.BotId,
		execBotIndex.ObjType,
		execBotIndex.ObjId,
		execBotIndex.LeadId,
		execBotIndex.LeadPayload,
		execBotIndex.Id,
		execBotIndex.IndexIncre,
		execBotIndex.BeginTimeDay,
		execBotIndex.BeginTime,
		execBotIndex.EndTime,
		execBotIndex.Created,
		execBotIndex.Updated,
		execBotb,
	).Exec()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "upsert exec bot index")
	}
	return nil
}

func (db *DB) LastExecBotIndex(accountId, botId, id string) (int64, error) {
	var indexIncre int64
	err := db.session.Query(`
			SELECT index_incre
			FROM `+Bizbot_tbl_exec_bot_index+`
			WHERE account_id=?
				AND bot_id=?
				AND id=?
			ORDER BY index_incre DESC
			LIMIT 1
		`,
		accountId,
		botId,
		id,
	).Scan(&indexIncre)
	if err != nil && err.Error() == gocql.ErrNotFound.Error() {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, 500, errors.E_database_error, "last exec bot index", accountId, id)
	}
	return indexIncre, nil
}

// TODO order by id
// Order by begin_time_day desc, id desc
func (db *DB) LoadExecBotIndexs(accountId, botId string, beginTimeDayFrom, beginTimeDayTo int64, orderDesc []string, c chan []*cr.ExecBotIndex, size int) error {
	defer close(c)

	for day := beginTimeDayTo; day >= beginTimeDayFrom; day-- {
		iter := db.session.Query(`
				SELECT
					obj_type,
					obj_id,
					lead_id,
					lead_payload,
					id,
					index_incre,
					begin_time_day,
					begin_time,
					end_time,
					created,
					updated
				FROM `+Bitbot_tbl_exec_bot_index_view+`
				WHERE account_id=? AND bot_id=? AND begin_time_day=?
				ORDER BY id DESC
			`,
			accountId, botId, day).Iter()

		var objType, objId, leadId, id string
		var leadPayload map[string]string
		var indexIncre, beginTimeDay, beginTime, endTime, created, updated int64
		arr := make([]*cr.ExecBotIndex, size)
		arrIndex := 0
		for iter.Scan(
			&objType,
			&objId,
			&leadId,
			&leadPayload,
			&id,
			&indexIncre,
			&beginTimeDay,
			&beginTime,
			&endTime,
			&created,
			&updated,
		) {
			execBotIndex := &cr.ExecBotIndex{
				AccountId:    accountId,
				BotId:        botId,
				ObjType:      objType,
				ObjId:        objId,
				LeadId:       leadId,
				LeadPayload:  leadPayload,
				Id:           id,
				IndexIncre:   indexIncre,
				BeginTimeDay: beginTimeDay,
				BeginTime:    beginTime,
				EndTime:      endTime,
				Created:      created,
				Updated:      updated,
			}

			arr[arrIndex] = execBotIndex
			arrIndex++
			if arrIndex == size {
				c <- arr
				arr = make([]*cr.ExecBotIndex, size)
				arrIndex = 0
			}
		}

		if arrIndex > 0 {
			c <- arr[:arrIndex]
		}

		err := iter.Close()
		if err != nil {
			return errors.Wrap(err, 500, errors.E_database_error, "read all bots", accountId)
		}
	}

	return nil
}

func (db *DB) ReadAllExecBotIndexs(accountId, botId, id string) ([]*cr.ExecBotIndex, error) {
	arr := make([]*cr.ExecBotIndex, 0)
	iter := db.session.Query(`
		SELECT
			obj_type,
			obj_id,
			lead_id,
			lead_payload,
			index_incre,
			begin_time_day,
			begin_time,
			end_time,
			created,
			updated,
			exec_bot
		FROM `+Bizbot_tbl_exec_bot_index+`
		WHERE account_id=? AND bot_id=? AND id=?
	`, accountId, botId, id).Iter()
	var objType, objId, leadId string
	var leadPayload map[string]string
	var indexIncre, beginTimeDay, beginTime, endTime, created, updated int64
	var execBotb []byte
	for iter.Scan(
		&objType,
		&objId,
		&leadId,
		&leadPayload,
		&indexIncre,
		&beginTimeDay,
		&beginTime,
		&endTime,
		&created,
		&updated,
		&execBotb,
	) {
		execBotIndex := &cr.ExecBotIndex{
			AccountId:    accountId,
			BotId:        botId,
			Id:           id,
			ObjType:      objType,
			ObjId:        objId,
			LeadId:       leadId,
			LeadPayload:  leadPayload,
			IndexIncre:   indexIncre,
			BeginTimeDay: beginTimeDay,
			BeginTime:    beginTime,
			EndTime:      endTime,
			Created:      created,
			Updated:      updated,
		}

		execBot := &cr.ExecBot{}
		json.Unmarshal(execBotb, execBot)
		execBotIndex.ExecBot = execBot

		arr = append(arr, execBotIndex)
	}

	err := iter.Close()
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, "read all bots", accountId)
	}

	return arr, nil
}
