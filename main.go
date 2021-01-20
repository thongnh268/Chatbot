package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	cr "git.subiz.net/bizbot/runner"
	"github.com/gorilla/mux"
	"github.com/subiz/errors"
	"github.com/subiz/goutils/log"
	"github.com/subiz/header"
	pb "github.com/subiz/header/account"
	ppb "github.com/subiz/header/payment"
	"github.com/urfave/cli"
)

var bizbot *Bizbot

func main() {
	bizbot = NewBizbot()

	app := cli.NewApp()
	app.Commands = []cli.Command{
		{Name: "daemon", Usage: "run server", Action: daemon},
		{Name: "dbcreate", Usage: "create db", Action: createDB},
		/*
			{Name: "dbdrop", Usage: "drop db", Action: dropDB},

			{Name: "dbseed", Usage: "", Action: seedDB},
			{Name: "dbcreateview", Usage: "create materialized view", Action: createDBView},
			{Name: "dbdropview", Usage: "drop materialized view", Action: dropDBView},
		*/
		{Name: "dbreset", Usage: "reset db", Action: resetDB},
		{Name: "dbcreate", Usage: "create db", Action: createDB},
		{Name: "dbseed", Usage: "seed bot", Action: seedBot},
		{Name: "botcount", Usage: "count bot state", Action: botcount},
		{Name: "botruni", Usage: "bot run information", Action: botruni},
		{Name: "boti", Usage: "bot run information", Action: boti},
		{Name: "botexport", Usage: "export all bots", Action: botexport},
		{Name: "execbotexport", Usage: "export terminated execbots in for last days", Action: exportTerminatedExecBot},
		{Name: "botrunreport", Usage: "reload report all bots", Action: botrunreport},
	}
	app.RunAndExitOnError()
}

func daemon(ctx *cli.Context) {
	// go func() { http.ListenAndServe("localhost:6060", nil) }()
	go bizbot.LoadJob()
	// go bizbot.ServeAsync()
	bizbot.ServeHTTP()
	// bizbot.ServeGrpc()
}

func (app *Bizbot) ServeHTTP() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/bots", GetBotHttp).Methods("GET")
	router.HandleFunc("/bot", CreateBotHttp).Methods("GET")
	log.Fatal(http.ListenAndServe(":9000", router))
}

func CreateBotHttp(w http.ResponseWriter, r *http.Request) {
	//TODO perm
	db := bizbot.GetDB()
	bot := fakeBot()
	err := db.UpsertBot(bot)
	if err != nil {
		return
	}
	b, err := db.ReadBot(bot.AccountId, bot.Id)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}

func GetBotHttp(w http.ResponseWriter, r *http.Request) {
	// TODO2 perm
	db := bizbot.GetDB()
	accountId := "acqwotroyghdejucjoxk"
	if accountId == "" {
		return
	}

	out, err := db.ReadAllBots(accountId)
	if err != nil {
		return
	}

	json.NewEncoder(w).Encode(out)
}

func botrunreport(ctx *cli.Context) {
	db := bizbot.GetDB()
	c := make(chan []*header.Bot)
	var dberr error
	go func() {
		err := db.LoadAllBots(c, 2)
		if err != nil {
			dberr = err
			return
		}
	}()

	limitDay := 30
	dayTo := time.Now().Unix() / (24 * 3600)
	dayFrom := dayTo - int64(limitDay)
	wg := &sync.WaitGroup{}
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func(index int) {
			defer func() {
				wg.Done()
			}()
			for {
				bots, has := <-c
				if !has {
					break
				}
				for _, bot := range bots {
					botReport, err := db.ReadBotReport(bot.GetAccountId(), bot.GetId())
					if err != nil {
						log.Error(err)
						return
					}
					if botReport == nil {
						botReport = cr.NewReportRun(bot.GetAccountId(), bot.GetId())
					}
					metrics := aggregateExecBotIndex(bot.GetAccountId(), bot.GetId(), dayFrom, dayTo)
					botReport.Reset(metrics)
					if err := db.UpsertBotReport(botReport); err != nil {
						log.Error(err)
						return
					}
					var totalLead, totalObject int64
					for _, metric := range botReport.List(dayFrom, dayTo) {
						totalLead += metric.LeadCount
						totalObject += metric.ObjectCount
					}
					fmt.Println(fmt.Sprintf("bot-id=%s total-lead=%d total-object=%d", bot.GetId(), totalLead, totalObject))
				}
			}
		}(i)
	}
	wg.Wait()
	if dberr != nil {
		fmt.Println(dberr)
		return
	}
}

func botcount(ctx *cli.Context) {
	db := bizbot.GetDB()
	c := make(chan []*header.Bot)
	var dberr error
	go func() {
		err := db.LoadAllBots(c, 100)
		if err != nil {
			dberr = err
			return
		}
	}()
	var count, countactive, countinactive int64
	for bots := range c {
		for _, bot := range bots {
			count++
			if bot.GetBotState() != "active" {
				countactive++
			}
			if bot.GetBotState() != "inactive" {
				countinactive++
			}
		}
	}
	if dberr != nil {
		fmt.Println(dberr)
		return
	}
	fmt.Println("count", count)
	fmt.Println("count-active", countactive)
	fmt.Println("count-inactive", countinactive)
}

func boti(clictx *cli.Context) {
	accountId := strings.TrimSpace(clictx.Args().Get(0))
	botId := strings.TrimSpace(clictx.Args().Get(1))
	db := bizbot.GetDB()
	bot, err := db.ReadBot(accountId, botId)
	if err != nil {
		fmt.Println(err)
		return
	}
	if bot == nil {
		fmt.Println(fmt.Sprintf("bot run information of (account-id,bot-id)=(%s,%s) not found", accountId, botId))
		return
	}
	b, err := json.MarshalIndent(bot, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fmt.Sprintf("bot run information of (account-id,conversation-id)=(%s,%s)", accountId, botId))
	fmt.Println(string(b))

	botReport, err := db.ReadBotReport(accountId, botId)
	if err != nil {
		fmt.Println(err)
		return
	}
	dayTo := time.Now().Unix() / (24 * 3600)
	dayFrom := dayTo - 6
	var totalLead, totalObject int64
	for _, metric := range botReport.List(dayFrom, dayTo) {
		totalLead += metric.LeadCount
		totalObject += metric.ObjectCount
	}
	fmt.Println(fmt.Sprintf("day-from=%d day-to=%d total-lead=%d total-object=%d", dayFrom, dayTo, totalLead, totalObject))
	fmt.Println(fmt.Sprintf("%#v", botReport.List(dayFrom, dayTo)))
	botReportb, err := json.MarshalIndent(botReport, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(botReportb))
}

func botexport(clictx *cli.Context) {
	limitDayStr := strings.TrimSpace(clictx.Args().Get(0))
	limitDay, _ := strconv.Atoi(limitDayStr)
	if limitDay == 0 {
		limitDay = 30
	}
	db := bizbot.GetDB()
	accdb := bizbot.GetSrcAccDB()
	paymentdb := bizbot.GetSrcPaymentDB()

	data := make([][]string, 0)
	now := time.Now()
	dayTo := now.Unix() / (24 * 3600)
	dayFrom := dayTo - int64(limitDay)
	data = append(data, []string{"account-id", "account_name", "subscription", "bot-id", "bot-fullname", "bot-state", "count-convos", "count-leads", "percent-convert-leads"})
	c := make(chan []*header.Bot)
	go func() {
		err := db.LoadAllBots(c, 100)
		if err != nil {
			log.Error(err)
			return
		}
	}()
	var index int
	// luu account da read
	var accMap = make(map[string]*pb.Account)
	var subMap = make(map[string]*ppb.Subscription)
	var convertLeadPercent float64
	var dberr error
	for bots := range c {
		for _, bot := range bots {
			if index%10 == 0 {
				fmt.Println(index)
			}
			index++
			metrics := aggregateExecBotIndex(bot.GetAccountId(), bot.GetId(), dayFrom, dayTo)
			var totalLead, totalObject int64
			for _, metric := range metrics {
				totalLead += metric.LeadCount
				totalObject += metric.ObjectCount
			}
			bot.CountLeadInLast_7Days = totalLead
			bot.CountConvInLast_7Days = totalObject
			if bot.GetCountConvInLast_7Days() == 0 {
				convertLeadPercent = 0
			} else {
				convertLeadPercent = (float64)(bot.CountLeadInLast_7Days) * 100 / (float64)(bot.CountConvInLast_7Days)
			}
			var acc *pb.Account
			acc, has := accMap[bot.GetAccountId()]
			if !has {
				acc, dberr = accdb.ReadAccount(bot.GetAccountId())
				if dberr != nil {
					break
				}
				accMap[bot.GetAccountId()] = acc
			}
			sub, has := subMap[bot.GetAccountId()]
			if !has {
				sub, dberr = paymentdb.GetSubscription(bot.GetAccountId())
				if dberr != nil {
					break
				}
				subMap[sub.GetAccountId()] = sub
			}
			data = append(data, []string{bot.GetAccountId(), acc.GetName(), sub.GetPlan(), bot.GetId(), bot.GetFullname(), bot.GetState(), strconv.Itoa(int(bot.GetCountConvInLast_7Days())), strconv.Itoa(int(bot.GetCountLeadInLast_7Days())), fmt.Sprintf("%f", convertLeadPercent)})
		}
	}
	if dberr != nil {
		fmt.Println(dberr)
		return
	}
	filename := now.Format("20060102") + "-" + now.AddDate(0, 0, -limitDay).Format("20060102") + ".csv"
	csvExport(filename, data)
	fmt.Println("End export bots")

}

func exportTerminatedExecBot(clictx *cli.Context) {
	limitDayStr := strings.TrimSpace(clictx.Args().Get(0))
	limitDay, _ := strconv.Atoi(limitDayStr)
	if limitDay == 0 {
		limitDay = 30
	}
	data := make([][]string, 0)
	data = append(data, []string{"account-id", "bot-id", "execbot-id", "error"})
	c := make(chan []*cr.ExecBot)
	go func() {
		err := loadRecentTerminateExecBots(c, 100, limitDay)
		if err != nil {
			log.Error(err)
			return
		}
	}()
	var index int
	for execbots := range c {
		for _, execbot := range execbots {
			if index%10 == 0 {
				fmt.Println(index)
			}
			index++
			if execbot.ActionTree != nil && len(execbot.ActionTree.RunTimes) > 0 {
				runtimes := execbot.ActionTree.RunTimes
				if runtimes[len(runtimes)-1].NextState == cr.Node_state_error {
					data = append(data, []string{execbot.AccountId, execbot.BotId, execbot.Id, runtimes[len(runtimes)-1].NextStateNote})
				}
			}
		}
	}
	now := time.Now()
	filename := now.Format("20060102") + "-" + now.AddDate(0, 0, -limitDay).Format("20060102") + ".csv"
	csvExport(filename, data)
	fmt.Println("End export terminated execbots")
}

func loadRecentTerminateExecBots(c chan []*cr.ExecBot, size, limitDay int) error {
	defer close(c)
	db := bizbot.GetDB()
	now := time.Now().UnixNano() / 1e6
	limit := now - int64(limitDay)*24*3600*1e3
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
			action_tree
		FROM `+Bizbot_tbl_exec_bot+`
		WHERE updated >=?
	`, limit).Iter()
	var accountId, botId, objType, objId, id, status, statusNote string
	var created, updated, lastSeen int64
	var actionTreeb []byte
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
		&actionTreeb,
	) {
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
		actionTree := &cr.ActionNode{}
		json.Unmarshal(actionTreeb, actionTree)
		execBot.ActionTree = actionTree
		arr[index] = execBot
		index++
		if index == size {
			c <- arr
			arr = make([]*cr.ExecBot, size)
			index = 0
		}
	}

	err := iter.Close()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error, "load all updated execbots in last %s days", limitDay)
	}
	return nil
}

// TODO rm
// Ref from exec bot mgr
func aggregateExecBotIndex(accountId, botId string, dayFrom, dayTo int64) []*header.Metric {
	c := make(chan []*cr.ExecBotIndex)
	db := bizbot.GetDB()

	go func() {
		err := db.LoadExecBotIndexs(
			accountId,
			botId,
			dayFrom,
			dayTo,
			[]string{"begin_time_day", "id"},
			c,
			100,
		)
		if err != nil {
			log.Error(err)
		}
	}()

	var objType string
	var dateDime int64
	var objId string
	var objLead bool

	metrics := make([]*header.Metric, 0)
	index := 0
	subindex := 0

	var metric *header.Metric
	var submetric *header.Metric

	metric = &header.Metric{}
	submetric = &header.Metric{}
	for execBotIndexs := range c {
		for _, execBotIndex := range execBotIndexs {
			if objId != execBotIndex.ObjId {
				submetric.ObjectCount += 1
				metric.ObjectCount += 1
				if objLead {
					submetric.LeadCount += 1
					metric.LeadCount += 1
				}
				objId = execBotIndex.ObjId
				objLead = false
			}

			if dateDime != execBotIndex.BeginTimeDay {
				metrics = append(metrics, &header.Metric{
					DateDim:    execBotIndex.BeginTimeDay,
					Submetrics: []*header.Metric{},
				})
				subindex = 0
				dateDime = execBotIndex.BeginTimeDay
				metric = metrics[index]
				index++
			}

			if objType != execBotIndex.ObjType {
				metric.Submetrics = append(metric.Submetrics, &header.Metric{
					ObjectType: execBotIndex.ObjType,
				})
				objType = execBotIndex.ObjType
				submetric = metric.Submetrics[subindex]
				subindex++
			}

			submetric.Count += 1
			metric.Count += 1
			if len(execBotIndex.LeadPayload) > 0 {
				objLead = true
			}
		}
	}

	submetric.ObjectCount += 1
	metric.ObjectCount += 1
	if objLead {
		submetric.LeadCount += 1
		metric.LeadCount += 1
	}

	return metrics
}

func csvExport(filename string, data [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		if err := writer.Write(value); err != nil {
			return err // let's return errors if necessary, rather than having a one-size-fits-all error handler
		}
	}
	return nil
}

func botruni(clictx *cli.Context) {
	accountId := strings.TrimSpace(clictx.Args().Get(0))
	convoId := strings.TrimSpace(clictx.Args().Get(1))
	limitStr := strings.TrimSpace(clictx.Args().Get(2))
	limitIndexStr := strings.TrimSpace(clictx.Args().Get(3))
	var limit, limitIndex int
	limit, _ = strconv.Atoi(limitStr)
	limitIndex, _ = strconv.Atoi(limitIndexStr)
	if limit == 0 {
		limit = 10
	}
	if limitIndex == 0 {
		limitIndex = 10
	}
	db := bizbot.GetDB()
	bots, err := db.ReadAllBots(accountId)
	if err != nil {
		fmt.Println(err)
		return
	}
	execBotIds := make([]string, len(bots))
	execBots := make([]*cr.ExecBot, 0)
	execBotIndexArr := make([]*cr.ExecBotIndex, 0)
	for i, bot := range bots {
		execBotId := cr.NewExecBotID(accountId, bot.GetId(), convoId)
		execBot, err := db.ReadExecBot(accountId, bot.GetId(), execBotId)
		if err != nil {
			fmt.Println(err)
			return
		}
		if execBot != nil {
			execBots = append(execBots, execBot)
		}
		execBotIndexs, err := db.ReadAllExecBotIndexs(accountId, bot.GetId(), execBotId)
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(execBotIndexs) > 0 {
			execBotIndexArr = append(execBotIndexArr, execBotIndexs...)
		}
		execBotIds[i] = execBotId
	}
	if len(execBots) == 0 && len(execBotIndexArr) == 0 {
		fmt.Println(fmt.Sprintf("bot run information of (account-id,conversation-id)=(%s,%s) not found", accountId, convoId))
	}
	b, err := json.MarshalIndent(&Botruni{ExecBots: execBots, ExecBotIndexs: execBotIndexArr}, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fmt.Sprintf("bot run information of (account-id,conversation-id)=(%s,%s)", accountId, convoId))
	fmt.Println(string(b))
}

type Botruni struct {
	ExecBots      []*cr.ExecBot      `json:"exec_bots"`
	ExecBotIndexs []*cr.ExecBotIndex `json:"exec_bot_indexs"`
}
