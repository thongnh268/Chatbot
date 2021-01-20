package main

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	ca "git.subiz.net/bizbot/action"
	cr "git.subiz.net/bizbot/runner"
	"github.com/golang/protobuf/proto"
	"github.com/kelseyhightower/envconfig"
	"github.com/subiz/errors"
	"github.com/subiz/goutils/log"
	"github.com/subiz/goutils/loop"
	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"github.com/subiz/kafka"
	"github.com/subiz/perm"
	"github.com/subiz/sgrpc"
	g "github.com/subiz/sgrpc"
	ggrpc "github.com/subiz/sgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

type Config struct {
	CassandraSeeds           []string
	KeyspacePrefix           string
	ReplicaFactor            int
	ProxyAddrs               []string
	Kafkabrokers             []string
	KafkaCsg                 string `default:"foobar"`
	GrpcPort                 string `default:"foobar"`
	Shards                   int
	Service                  string
	ConvoService             string
	RealtimeService          string
	UserService              string
	SrcaccKeyspacePrefix     string
	SrcpaymentKeyspacePrefix string
}

type Bizbot struct {
	*sync.Mutex
	Cf                Config
	dbMutex           *sync.Mutex
	db                *DB
	schedMutex        *sync.Mutex
	sched             *Schedule
	ShardNumber       int // readonly
	Shards            []*BizbotShard
	kafkaMutex        *sync.Mutex
	pub               *kafka.Publisher
	server            *Server
	srcAccDB          *AccDB
	srcAccDBMutex     *sync.Mutex
	srcPaymentDB      *PaymentDB
	srcPaymentDBMutex *sync.Mutex
	// lock dial to another bizbot service
	dialLock *sync.Mutex
	bizbot   header.BizbotClient
	realtime header.PubsubClient
	msg      header.ConversationEventReaderClient
	user     header.UserMgrClient
	convo    header.ConversationMgrClient
	event    header.EventMgrClient
}

func NewBizbot() *Bizbot {
	var cf Config
	envconfig.MustProcess("bizbot", &cf)
	if len(cf.CassandraSeeds) == 0 {
		cf.CassandraSeeds = []string{"127.0.0.1:9042"}
	}
	app := &Bizbot{
		Mutex:             &sync.Mutex{},
		dialLock:          &sync.Mutex{},
		dbMutex:           &sync.Mutex{},
		srcAccDBMutex:     &sync.Mutex{},
		srcPaymentDBMutex: &sync.Mutex{},
		schedMutex:        &sync.Mutex{},
		kafkaMutex:        &sync.Mutex{},
		Cf:                cf,
	}

	app.ShardNumber = 2048
	app.Shards = make([]*BizbotShard, app.ShardNumber)
	for i := 0; i < app.ShardNumber; i++ {
		app.Shards[i] = &BizbotShard{
			Mutex:         &sync.Mutex{},
			LogicAccounts: make(map[string]*LogicAccount),
		}
	}

	app.sched = NewSchedule()

	return app
}

func (app *Bizbot) GetLogicAccounts() []*LogicAccount {
	accArr := make([]*LogicAccount, 0)
	for _, shard := range app.Shards {
		accArr = append(accArr, shard.GetLogicAccounts()...)
	}
	return accArr
}

func (app *Bizbot) FilterLogicAccounts(idArr []string) []*LogicAccount {
	idMap := make(map[string]struct{})
	for _, id := range idArr {
		idMap[id] = struct{}{}
	}

	accArr := make([]*LogicAccount, 0)
	for id := range idMap {
		acc := app.getLogicAccount(id)
		if acc != nil {
			accArr = append(accArr, acc)
		}
	}
	return accArr
}

func (app *Bizbot) GetDB() *DB {
	app.dbMutex.Lock()
	defer app.dbMutex.Unlock()

	if app.db == nil {
		app.db = NewDB(app.Cf.CassandraSeeds, app.Cf.KeyspacePrefix+Bizbot_keyspace)
	}
	return app.db
}

func (app *Bizbot) getLogicAccount(accountId string) *LogicAccount {
	if accountId == "" {
		return nil
	}

	accountIdNum := int(crc32.ChecksumIEEE([]byte(accountId)))
	shard := app.Shards[accountIdNum%app.ShardNumber]
	shard.Lock()
	defer shard.Unlock()

	if account, has := shard.LogicAccounts[accountId]; has {
		return account
	} else {
		log.Info("NEWACCOUNTTTTTTTT", accountId)
		db := app.GetDB()
		realtime, _ := app.GetRealtimeClient()
		msg, _ := app.GetMesClient()
		user, _ := app.GetUserClient()
		convo, _ := app.GetConversationMgrClient()
		event, _ := app.GetEventClient()
		execBotMgr := NewExecBotMgr(accountId, db, realtime)
		caMgr := &ActionMgrConv{caMgr: ca.NewActionMgr(msg, user, convo, event)}
		sched := app.GetSched()
		runner := cr.NewBotRunner(accountId, execBotMgr, sched, caMgr)
		runner.Realtime = realtime
		runner.ObjectMgr, _ = app.GetConversationMgrClient()
		botMgr := NewBotMgr(accountId, db, execBotMgr)

		account := &LogicAccount{
			Id:         accountId,
			Mutex:      &sync.Mutex{},
			botRunner:  runner,
			actionMgr:  caMgr,
			botMgr:     botMgr,
			execBotMgr: execBotMgr,
		}
		shard.LogicAccounts[account.Id] = account

		return account
	}
}

func (app *Bizbot) GetPub() *kafka.Publisher {
	app.kafkaMutex.Lock()
	defer app.kafkaMutex.Unlock()

	if app.pub == nil {
		app.pub = kafka.NewPublisher(app.Cf.Kafkabrokers)
	}
	return app.pub
}

func (app *Bizbot) GetSched() *Schedule {
	app.schedMutex.Lock()
	defer app.schedMutex.Unlock()

	if app.sched == nil {
		app.sched = NewSchedule()
	}
	return app.sched
}

func (app *Bizbot) GetRunner(accid string) *cr.BotRunner {
	account := app.getLogicAccount(accid)
	return account.botRunner
}

func (app *Bizbot) GetActionMgr(accid string) *ActionMgrConv {
	account := app.getLogicAccount(accid)
	return account.actionMgr
}

func (app *Bizbot) GetBotMgr(accid string) *BotMgr {
	account := app.getLogicAccount(accid)
	return account.botMgr
}

func (app *Bizbot) GetExecBotMgr(accid string) *ExecBotMgr {
	account := app.getLogicAccount(accid)
	return account.execBotMgr
}

func (app *Bizbot) ServeHTTP1() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) { fmt.Fprintf(w, "pong from bizbot") })
	http.HandleFunc("/botrunner/statuses", func(w http.ResponseWriter, req *http.Request) {
		status := cr.NewBotRunnerStatus(app.Cf.Service)
		for _, shard := range app.Shards {
			for _, account := range shard.GetLogicAccounts() {
				status.Add(account.botRunner.Status())
			}
		}
		if status.SourceStatus != nil {
			sort.Slice(status.SourceStatus, func(i, j int) bool {
				return status.SourceStatus[i].ExecBotCount > status.SourceStatus[j].ExecBotCount
			})
		}
		statusb, err := json.Marshal(status)
		if err != nil {
			fmt.Fprintf(w, "err=%s", err)
		}
		fmt.Fprintf(w, string(statusb))
	})
	http.HandleFunc("/botrunner/status/", func(w http.ResponseWriter, req *http.Request) {
		accountId := req.URL.Path[len("/botrunner/status/"):]
		if accountId == "" {
			fmt.Fprintf(w, "account id is null")
			return
		}
		status := app.GetRunner(accountId).Status()
		statusb, err := json.Marshal(status)
		if err != nil {
			fmt.Fprintf(w, "err=%s", err)
		}
		fmt.Fprintf(w, string(statusb))
	})
	http.ListenAndServe(":9000", nil)
}

func (app *Bizbot) ServeGrpc() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", app.Cf.GrpcPort))
	if err != nil {
		log.Warn("failed to listen: %v", err)
		return err
	}

	hostname, _ := os.Hostname()
	sp := strings.Split(hostname, "-")
	if len(sp) < 2 {
		panic("invalid hostname" + hostname)
	}
	ordinal := sp[len(sp)-1]
	pari64, _ := strconv.ParseInt(ordinal, 10, 0)
	ordinal_num := int(pari64)
	hosts := make([]string, 0)
	for i := 0; i < app.Cf.Shards; i++ {
		// convo-${i}.convo:{port}
		hosts = append(hosts, sp[0]+"-"+strconv.Itoa(i)+"."+sp[0]+":"+app.Cf.GrpcPort)
	}

	grpcServer := grpc.NewServer(grpc.KeepaliveParams(
		keepalive.ServerParameters{MaxConnectionAge: time.Duration(20) * time.Second},
	), grpc.UnaryInterceptor(g.NewServerShardInterceptor(hosts, ordinal_num)))

	server := &Server{}
	server.accMgr = app

	app.server = server
	header.RegisterBizbotServer(grpcServer, app.server)

	grpcServer.Serve(lis)
	return nil
}

func (app *Bizbot) ServeAsync() {
	log.Info("START serve ASYNC...")
	h := kafka.NewHandler(bizbot.Cf.Kafkabrokers, bizbot.Cf.KafkaCsg, header.E_BotSynced.String(), true)
	h.Serve(map[string]func(*pb.Context, []byte){
		"BotEventCreated": func(ctx *pb.Context, val []byte) {
			req := &header.BotEventCreated{}
			if err := proto.Unmarshal(val, req); err != nil {
				log.Error(err)
				return
			}

			start := time.Now()
			kafkapar := ctx.GetKafkaPartition()
			kafkaoff := ctx.GetKafkaOffset()
			accid := req.GetAccountId()

			if kafkapar == 69 {
				log.Info(kafkapar, kafkaoff, req.GetConversationId(), "bizbot-client.on-event=beginaa", fmt.Sprintf("account-id=%s", accid), fmt.Sprintf("evt.type=%s evt.by-id=%s evt.id=%s", req.GetEvent().GetType(), req.GetEvent().GetBy().GetId(), req.GetEvent().GetId()))
				return
			}

			defer func() {
				if kafkapar == 69 {
					log.Info("TIME EXIT", time.Since(start))
				}

				log.Info(kafkapar, kafkaoff, req.GetConversationId(), "bizbot-client.on-event=begin000end", fmt.Sprintf("account-id=%s", accid), fmt.Sprintf("evt.type=%s evt.by-id=%s evt.id=%s", req.GetEvent().GetType(), req.GetEvent().GetBy().GetId(), req.GetEvent().GetId()))
			}()
			log.Info(kafkapar, kafkaoff, req.GetConversationId(), "bizbot-client.on-event=begin000", fmt.Sprintf("account-id=%s", accid), fmt.Sprintf("evt.type=%s evt.by-id=%s evt.id=%s", req.GetEvent().GetType(), req.GetEvent().GetBy().GetId(), req.GetEvent().GetId()))
			if req.GetEvent().GetType() == header.RealtimeType_message_pong.String() {
				log.Info(req.GetConversationId(), "ignore-event", fmt.Sprintf("account-id=%s", req.GetAccountId()), fmt.Sprintf("evt.type=%s evt.by-id=%s evt.id=%s", req.GetEvent().GetType(), req.GetEvent().GetBy().GetId(), req.GetEvent().GetId()))
				return
			}

			loop.LoopErr(func() error {

				client, err := app.GetBizbotClient(accid)
				if err != nil {
					log.Error(err)
					return nil // TODO
				}

				if kafkapar == 69 {
					log.Info("TIME GET", time.Since(start))
				}

				allperm := perm.MakeBase()

				// call like api
				log.Info(req.GetConversationId(), "bizbot-client.on-event=begin", fmt.Sprintf("account-id=%s", accid), fmt.Sprintf("evt.type=%s evt.by-id=%s evt.id=%s", req.GetEvent().GetType(), req.GetEvent().GetBy().GetId(), req.GetEvent().GetId()))
				_, err = client.OnEvent(ggrpc.ToGrpcCtx(&pb.Context{
					Credential: &pb.Credential{
						AccountId: accid,
						Issuer:    "subiz",
						Type:      pb.Type_subiz,
						Perm:      &allperm,
					},
				}), &header.RunRequest{
					ObjectId:   req.GetConversationId(),
					ObjectType: header.BotCategory_conversations.String(),
					Event:      req.Event,
					Created:    req.GetCreated(),
				})

				if kafkapar == 69 {
					log.Info("TIME ONEVENT", time.Since(start))
				}
				log.Info(req.GetConversationId(), "bizbot-client.on-event=end", fmt.Sprintf("account-id=%s", accid), fmt.Sprintf("err=%s", err))
				return nil
			})
		},
		header.E_StartBot.String(): func(ctx *pb.Context, val []byte) {
			req := &header.BotRequest{}
			if err := proto.Unmarshal(val, req); err != nil {
				log.Error(err)
				return
			}

			loop.LoopErr(func() error {
				accid := req.GetAccountId()
				client, err := app.GetBizbotClient(accid)
				if err != nil {
					log.Error(err)
					return nil // TODO
				}

				allperm := perm.MakeBase()
				// call like api
				log.Info(req.GetConversationId(), "bizbot-client.start-bot=begin", fmt.Sprintf("account-id=%s bot-id=%s", accid, req.GetBotId()))
				_, err = client.StartBot(ggrpc.ToGrpcCtx(&pb.Context{
					Credential: &pb.Credential{
						AccountId: accid,
						Issuer:    "subiz",
						Type:      pb.Type_subiz,
						Perm:      &allperm,
					},
				}), &header.RunRequest{
					AccountId:  req.GetAccountId(),
					BotId:      req.GetBotId(), // or BotTriggerType: "conversation_start" & evt.data.conversation.id
					ObjectId:   req.GetConversationId(),
					ObjectType: header.BotCategory_conversations.String(),
					Created:    req.GetCreated(),
					ObjectContexts: []*pb.KV{
						{
							Key:   "user_id",
							Value: req.UserId,
						},
					},
				})
				if err != nil {
					created := time.Now().UnixNano() / 1e6
					app.GetDB().UpsertExecBot(&cr.ExecBot{
						AccountId:  req.GetAccountId(),
						BotId:      req.GetBotId(),
						ObjType:    header.BotCategory_conversations.String(),
						ObjId:      req.GetConversationId(),
						LeadId:     req.GetUserId(),
						Id:         cr.NewExecBotID(req.GetAccountId(), req.GetBotId(), req.GetConversationId()),
						Created:    created,
						Updated:    created,
						Status:     cr.Exec_bot_state_terminated,
						StatusNote: header.BotTerminated_start.String(),
						ActionTree: &cr.ActionNode{RunTimes: []*cr.RunTime{{NextState: cr.Node_state_error, NextStateNote: err.Error()}}},
					})
				}
				log.Info(req.GetConversationId(), "bizbot-client.start-bot=end", fmt.Sprintf("account-id=%s bot-id=%s", accid, req.GetBotId()), fmt.Sprintf("err=%s", err))
				return nil
			})
		},
	}, func([]int32) {})
	log.Info("END serve ASYNC")
}

func (app *Bizbot) LoadJob() {
	log.Info("LOAD JOB...")
	sched := app.GetSched()

	sched.BatchAsync()

	sched.BatchFunc = func(idArr []string, schedNow int64) {
		logicAccounts := app.FilterLogicAccounts(idArr)
		if len(logicAccounts) == 0 {
			return
		}
		for _, account := range logicAccounts {
			go account.botRunner.OnTime(schedNow)
		}
	}

	sched.StartAsync()

	accountJobId := sched.Index()
	sched.runner.Every(10).Seconds().Do(func() {
		job := &SimpleJob{
			Id: accountJobId,
			Run: func() {
				// replace by shard and online accounts
				/*
					db := bizbot.GetDB()
					accountIds, err := db.ReadAccountIds()
					if err != nil {
						log.Error(err)
						return
					}
					for _, accountId := range accountIds {
						bizbot.getLogicAccount(accountId)
					}
				*/
			},
		}
		job.Execute()
	})

	cleanExecBotJobId := sched.Index()
	sched.runner.Every(15).Seconds().Do(func() {
		job := &AccountJob{
			Id:           cleanExecBotJobId,
			LoadAccounts: app.GetLogicAccounts,
			Run: func(account *LogicAccount) {
				account.botRunner.CleanExecBots()
			},
		}
		job.Execute()
	})

	saveRunBotJobId := sched.Index()
	sched.runner.Every(15).Seconds().Do(func() {
		job := &AccountJob{
			Id:           saveRunBotJobId,
			LoadAccounts: app.GetLogicAccounts,
			Run: func(account *LogicAccount) {
				runBots := account.botRunner.GetBotMap()
				if len(runBots) == 0 {
					return
				}
				// TODO only primary key, last executed
				bots, err := account.botMgr.ListBots(&pb.Id{})
				if err != nil {
					log.Error(err)
					return
				}
				var runBot *header.Bot
				var has bool
				for _, bot := range bots {
					runBot, has = runBots[bot.GetId()]
					if !has || runBot.GetLastExecuted() <= bot.GetLastExecuted() {
						continue
					}
					bot.LastExecuted = runBot.GetLastExecuted()
					_, err := account.botMgr.UpdateBot(bot)
					if err != nil {
						log.Error(err)
						return
					}
				}
			},
		}
		job.Execute()
	})

	saveRunReportJobId := sched.Index()
	sched.runner.Every(5).Seconds().Do(func() {
		job := &AccountJob{
			Id:           saveRunReportJobId,
			LoadAccounts: app.GetLogicAccounts,
			Run: func(account *LogicAccount) {
				botReports := account.botRunner.GetChangedBotReports()
				if len(botReports) == 0 {
					return
				}
				db := bizbot.GetDB()
				for _, botReport := range botReports {
					if err := db.UpsertBotReport(botReport); err != nil {
						log.Error(err)
						return
					}
				}
			},
		}
		job.Execute()
	})

	resetRunReportJobId := sched.Index()
	resetRunReportFunc := func(account *LogicAccount) {
		// TODO only primary key, last executed
		bots, err := account.botMgr.ListBots(&pb.Id{})
		if err != nil {
			log.Error(err)
			return
		}
		runner := account.botRunner
		limitDay := 7
		dayTo := time.Now().Unix() / (24 * 3600)
		dayFrom := dayTo - int64(limitDay)
		for _, bot := range bots {
			metrics := account.execBotMgr.AggregateExecBotIndex(bot.GetId(), dayFrom, dayTo)
			runner.ResetBotReport(bot.GetId(), metrics)
		}
	}
	sched.runner.Every(2).Hours().Do(func() {
		job := &AccountJob{
			Id:           resetRunReportJobId,
			LoadAccounts: app.GetLogicAccounts,
			Run:          resetRunReportFunc,
		}
		job.Execute()
	})
	// TODO
	sched.runner.Every(1).Day().At("10:30").Do(func() {
		/*
			job := &AccountJob{
				Id:           resetRunReportJobId,
				LoadAccounts: app.GetLogicAccounts,
				Run:          resetRunReportFunc,
			}
			job.Execute()
		*/
	})

	loadBotJobId := sched.Index()
	sched.runner.Every(5).Seconds().Do(func() {
		job := &AccountJob{
			Id:           loadBotJobId,
			LoadAccounts: app.GetLogicAccounts,
			Run: func(account *LogicAccount) {
				runner := account.botRunner
				db := bizbot.GetDB()
				botReports, err := db.ReadAllBotReports(account.Id)
				if err != nil {
					log.Error(err)
					return
				}
				dbbots, err := db.ReadAllBots(account.Id)
				if err != nil {
					log.Error(err)
					return
				}
				activeBots := make([]*header.Bot, 0)
				for _, dbbot := range dbbots {
					if dbbot.GetBotState() != "active" {
						continue
					}
					//  default obj type
					if dbbot.GetCategory() == "" {
						dbbot.Category = header.BotCategory_conversations.String()
					}
					activeBots = append(activeBots, dbbot)
				}
				runner.LoadBots(activeBots, botReports)
			},
		}
		job.Execute()
	})

	loadExecBotJobId := sched.Index()
	accMap := make(map[string]struct{}) // one time
	accMapLock := &sync.Mutex{}
	sched.runner.Every(5).Seconds().Do(func() {
		job := &AccountJob{
			Id:           loadExecBotJobId,
			LoadAccounts: app.GetLogicAccounts,
			Run: func(account *LogicAccount) {
				var isLoaded bool
				accMapLock.Lock()
				if _, has := accMap[account.Id]; has {
					isLoaded = true
				}
				accMapLock.Unlock()
				if isLoaded {
					return
				}
				runner := account.botRunner
				db := bizbot.GetDB()
				botReports, err := db.ReadAllBotReports(account.Id)
				if err != nil {
					log.Error(err)
					return
				}
				// ref from load bot job
				dbbots, err := db.ReadAllBots(account.Id)
				if err != nil {
					log.Error(err)
					return
				}
				activeBots := make([]*header.Bot, 0)
				for _, dbbot := range dbbots {
					if dbbot.GetBotState() != "active" {
						continue
					}
					//  default obj type
					if dbbot.GetCategory() == "" {
						dbbot.Category = header.BotCategory_conversations.String()
					}
					activeBots = append(activeBots, dbbot)
				}
				runner.LoadBots(activeBots, botReports)

				runExecBots := runner.GetExecBotMap()
				c := make(chan []*cr.ExecBot)
				go func() {
					err := db.LoadAllExecBots(account.Id, c, 100, runExecBots)
					if err != nil {
						log.Error(err)
					}
				}()
				for srcExecBots := range c {
					runner.LoadExecBots(srcExecBots)
				}
				accMapLock.Lock()
				accMap[account.Id] = struct{}{}
				accMapLock.Unlock()
			},
		}
		job.Execute()
	})

	onLockRunnerJobId := sched.Index()
	sched.runner.Every(30).Seconds().Do(func() {
		job := &AccountJob{
			Id:           onLockRunnerJobId,
			LoadAccounts: app.GetLogicAccounts,
			Run: func(account *LogicAccount) {
				runner := account.botRunner
				runner.OnLock()
			},
		}
		job.Execute()
	})

	cleanIndexingJobId := sched.Index()
	sched.runner.Every(60).Seconds().Do(func() {
		job := &AccountJob{
			Id:           cleanIndexingJobId,
			LoadAccounts: app.GetLogicAccounts,
			Run: func(account *LogicAccount) {
				runner := account.botRunner
				runner.CleanIndexings()
			},
		}
		job.Execute()
	})

	onForgotJobId := sched.Index()
	sched.runner.Every(Job_ticker_number).Seconds().Do(func() {
		job := &AccountJob{
			Id:           onForgotJobId,
			LoadAccounts: app.GetLogicAccounts,
			Run: func(account *LogicAccount) {
				runner := account.botRunner
				runner.OnForgot()
			},
		}
		job.Execute()
	})

	cleanTblExecBotJobId := sched.Index()
	sched.runner.Every(30).Seconds().Do(func() {
		job := &SimpleJob{
			Id: cleanTblExecBotJobId,
			// TODO1 batch delete
			Run: func() {
				db := bizbot.GetDB()
				now := time.Now().UnixNano() / 1e6
				upperBound := now - 7*24*60*60*1000 // cr.Runner_lock_time_expire

				c := make(chan []*cr.ExecBot)
				go func() {
					err := db.LoadTerminatedExecBots(c, 100)
					if err != nil {
						log.Error(err)
					}
				}()
				for execBots := range c {
					for _, execBot := range execBots {
						if execBot.Status != cr.Exec_bot_state_terminated {
							continue
						}
						if execBot.Updated > upperBound {
							continue
						}
						db.DeleteExecBot(execBot.AccountId, execBot.BotId, execBot.Id)
					}
				}
			},
		}
		job.Execute()
	})
}

type BizbotShard struct {
	*sync.Mutex
	LogicAccounts map[string]*LogicAccount
}

func (shard *BizbotShard) GetLogicAccounts() []*LogicAccount {
	shard.Lock()
	defer shard.Unlock()
	accArr := make([]*LogicAccount, len(shard.LogicAccounts))
	index := 0
	for _, acc := range shard.LogicAccounts {
		accArr[index] = acc
		index++
	}
	return accArr
}

// TODO2 replace by global mgr
type LogicAccount struct {
	*sync.Mutex
	Id         string
	botRunner  *cr.BotRunner
	actionMgr  *ActionMgrConv
	botMgr     *BotMgr
	execBotMgr *ExecBotMgr
}

type ActionMgrConv struct {
	caMgr *ca.ActionMgr
}

func (mgr *ActionMgrConv) GetEvtCondTriggers(evt *header.Event) []string {
	return mgr.caMgr.GetEvtCondTriggers(evt)
}

func (mgr *ActionMgrConv) Do(node *cr.ActionNode, req *header.RunRequest) ([]string, error) {
	var crNode interface{}
	crNode = node
	var caNode ca.ActionNode
	var ok bool
	if caNode, ok = crNode.(ca.ActionNode); !ok {
		panic("trigger func connect failure")
	}
	return mgr.caMgr.Do(caNode, req)
}
func (mgr *ActionMgrConv) Wait(node *cr.ActionNode) string {
	var crNode interface{}
	crNode = node
	var caNode ca.ActionNode
	var ok bool
	if caNode, ok = crNode.(ca.ActionNode); !ok {
		panic("trigger func connect failure")
	}
	return mgr.caMgr.Wait(caNode)
}
func (mgr *ActionMgrConv) Triggers(node *cr.ActionNode) []string {
	var crNode interface{}
	crNode = node
	var caNode ca.ActionNode
	var ok bool
	if caNode, ok = crNode.(ca.ActionNode); !ok {
		panic("trigger func connect failure")
	}
	return mgr.caMgr.Triggers(caNode)
}

const (
	Bizbot_keyspace                = "bizbot"
	Bizbot_tbl_bot                 = "bots"
	Bizbot_tbl_exec_bot            = "exec_bots"
	Bizbot_tbl_exec_bot_index      = "exec_bot_indexs"
	Bitbot_tbl_exec_bot_index_view = "exec_bot_indexs_begin_time_day"

	// TODO ((account_id, id), tag)
	Bizbot_tbl_bot_tag      = "bot_tags"
	Bizbot_tbl_bot_tag_view = "bot_tags_tag"
)

func (app *Bizbot) GetBizbotClient(key string) (header.BizbotClient, error) {
	app.dialLock.Lock()
	defer app.dialLock.Unlock()

	if app.bizbot == nil {
		parts := strings.SplitN(app.Cf.Service, ":", 2)
		name, port := parts[0], parts[1]
		// address: [pod name] + "." + [service name] + ":" + [pod port]
		conn, err := dialGrpc(name+"-0."+name+":"+port, sgrpc.WithShardRedirect())
		if err != nil {
			return nil, err
		}
		app.bizbot = header.NewBizbotClient(conn)
	}
	return app.bizbot, nil
}

func dialGrpc(service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append([]grpc.DialOption{}, opts...)
	opts = append(opts, grpc.WithInsecure())
	// Enabling WithBlock tells the client to not give up trying to find a server
	opts = append(opts, grpc.WithBalancerName(roundrobin.Name))
	// opts = append(opts, sgrpc.WithCache())

	return grpc.Dial(service, opts...)
}

func (app *Bizbot) GetRealtimeClient() (header.PubsubClient, error) {
	app.dialLock.Lock()
	defer app.dialLock.Unlock()
	if app.realtime == nil {
		conn, err := grpc.Dial(app.Cf.RealtimeService, grpc.WithInsecure(), sgrpc.WithShardRedirect())
		if err != nil {
			return nil, err
		}
		app.realtime = header.NewPubsubClient(conn)
	}
	return app.realtime, nil
}

func (app *Bizbot) GetConversationMgrClient() (header.ConversationMgrClient, error) {
	app.dialLock.Lock()
	defer app.dialLock.Unlock()

	if app.convo == nil {
		parts := strings.SplitN(app.Cf.ConvoService, ":", 2)
		name, port := parts[0], parts[1]
		// address: [pod name] + "." + [service name] + ":" + [pod port]
		conn, err := dialGrpc(name+"-0."+name+":"+port, sgrpc.WithShardRedirect())
		if err != nil {
			return nil, err
		}
		app.convo = header.NewConversationMgrClient(conn)
	}
	return app.convo, nil
}

func (app *Bizbot) GetMesClient() (header.ConversationEventReaderClient, error) {
	app.dialLock.Lock()
	defer app.dialLock.Unlock()

	if app.msg == nil {
		parts := strings.SplitN(app.Cf.ConvoService, ":", 2)
		name, port := parts[0], parts[1]
		// address: [pod name] + "." + [service name] + ":" + [pod port]
		conn, err := dialGrpc(name+"-0."+name+":"+port, sgrpc.WithShardRedirect())
		if err != nil {
			return nil, err
		}
		app.msg = header.NewConversationEventReaderClient(conn)
	}
	return app.msg, nil
}

func (app *Bizbot) GetUserClient() (header.UserMgrClient, error) {
	app.dialLock.Lock()
	defer app.dialLock.Unlock()

	if app.user == nil {
		parts := strings.SplitN(app.Cf.UserService, ":", 2)
		name, port := parts[0], parts[1]
		conn, err := dialGrpc(name+"-0."+name+":"+port, sgrpc.WithShardRedirect())
		if err != nil {
			return nil, errors.Wrap(err, 500, errors.E_server_error, conn)
		}
		app.user = header.NewUserMgrClient(conn)
	}

	return app.user, nil
}

func (app *Bizbot) GetEventClient() (header.EventMgrClient, error) {
	app.dialLock.Lock()
	defer app.dialLock.Unlock()

	if app.event == nil {
		parts := strings.SplitN(app.Cf.UserService, ":", 2)
		name, port := parts[0], parts[1]
		conn, err := dialGrpc(name+"-0."+name+":"+port, sgrpc.WithShardRedirect())
		if err != nil {
			return nil, errors.Wrap(err, 500, errors.E_server_error, conn)
		}
		app.event = header.NewEventMgrClient(conn)
	}

	return app.event, nil
}

func (app *Bizbot) GetSrcAccDB() *AccDB {
	app.srcAccDBMutex.Lock()
	defer app.srcAccDBMutex.Unlock()

	if app.srcAccDB == nil {
		app.srcAccDB = NewAccDB(app.Cf.CassandraSeeds, app.Cf.SrcaccKeyspacePrefix+Srcacc_keyspace)
	}

	return app.srcAccDB
}

func (app *Bizbot) GetSrcPaymentDB() *PaymentDB {
	app.srcPaymentDBMutex.Lock()
	defer app.srcPaymentDBMutex.Unlock()

	if app.srcPaymentDB == nil {
		log.Info(app.Cf.SrcpaymentKeyspacePrefix + Srcpayment_keyspace)
		app.srcPaymentDB = NewPaymentDB(app.Cf.CassandraSeeds, app.Cf.SrcpaymentKeyspacePrefix+Srcpayment_keyspace)
	}

	return app.srcPaymentDB
}

// Report, start time or end time, run time matter
// 1. belongs to account
// 2. time dimension: day, hour ? start time
// 3. obj dimension: type, id
// 4. bot dimension: bot, bot version/append-tags
// 5. fact: exec bot id, index ? exec bot has index E.g. bot/update, exec bot/reset, exec bot logs/append-indexs
// 6. fact items: goal id (in exec bot log)

// Join other fact
// 1. user facts
// 2. conversation facts

// use bot.bot_state, bot.state is agent state(=active)
