package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/subiz/goutils/log"
)

// 100ms
type Schedule struct {
	*sync.Mutex
	runner *gocron.Scheduler
	index  int64

	head      int
	headAt    int64
	idBatch   [][]string
	running   bool
	stopChan  chan struct{} // signal to stop scheduling
	BatchFunc func(idArr []string, schedNow int64)
}

func NewSchedule() *Schedule {
	sched := &Schedule{Mutex: &sync.Mutex{}}
	sched.index = 1
	sched.runner = gocron.NewScheduler(time.UTC)
	sched.head = 0
	sched.headAt = time.Now().UnixNano() / 1e6 / 1e2
	sched.idBatch = make([][]string, Job_ticker_number)
	for i := 0; i < Job_ticker_number; i++ {
		sched.idBatch[i] = make([]string, 0)
	}
	sched.stopChan = make(chan struct{})
	sched.BatchFunc = func(idMap []string, schedNow int64) {}

	return sched
}

func (sched *Schedule) Push(schedAtMillisecond int64, accountId string) bool {
	sched.Lock()
	defer sched.Unlock()
	schedAt := schedAtMillisecond / 1e2

	if schedAt < sched.headAt-Job_ticker_expire {
		// expired
		return false
	}

	if schedAt <= sched.headAt {
		go sched.BatchFunc([]string{accountId}, schedAt)
		return true
	}

	index := int(schedAt - sched.headAt)
	if index >= Job_ticker_number {
		// too far
		return false
	}

	rindex := (sched.head + index) % Job_ticker_number
	sched.idBatch[rindex] = append(sched.idBatch[rindex], accountId)
	return true
}

func (sched *Schedule) Prepare(schedAtMillisecond int64, accountIds []string) {
	sched.Lock()
	defer sched.Unlock()
	schedAt := schedAtMillisecond / 1e2

	index := int(schedAt - sched.headAt)
	if index < Job_ticker_going {
		// too soon
		return
	}
	if index >= Job_ticker_number {
		// too far
		return
	}

	rindex := (sched.head + index) % Job_ticker_number
	sched.idBatch[rindex] = accountIds
}

func (sched *Schedule) BatchAsync() chan struct{} {
	sched.Lock()
	defer sched.Unlock()
	if sched.running {
		return sched.stopChan
	}
	sched.running = true
	ticker := time.NewTicker(1e2 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				sched.RunBatch()
			case <-sched.stopChan:
				ticker.Stop()
				sched.running = false
				return
			}
		}
	}()

	return sched.stopChan
}

func (sched *Schedule) RunBatch() {
	sched.Lock()
	defer sched.Unlock()
	now := time.Now().UnixNano() / 1e6 / 1e2
	if now <= sched.headAt {
		return
	}
	if now-sched.headAt >= Job_ticker_number {
		sched.head = 0
		sched.headAt = now
		sched.idBatch = make([][]string, Job_ticker_number)
		return
	}

	step := int(now - sched.headAt)
	idArr := make([]string, 0)
	for i := 0; i < int(step); i++ {
		rindex := (sched.head + i + 1) % Job_ticker_number
		if i+Job_ticker_expire >= step {
			idArr = append(idArr, sched.idBatch[rindex]...)
		}
		sched.idBatch[rindex] = make([]string, 0)
	}

	go sched.BatchFunc(idArr, now)

	sched.headAt = now
	sched.head += step
}

func (sched *Schedule) NextBatch(index int) []string {
	sched.Lock()
	defer sched.Unlock()
	if index < 1 || index > (Job_ticker_number-1) {
		return []string{}
	}
	return sched.idBatch[(sched.head+index)%Job_ticker_number]
}

func (sched *Schedule) Index() int {
	return int(atomic.AddInt64(&sched.index, 1))
}

func (sched *Schedule) StartAsync() {
	sched.runner.StartAsync()
}

const (
	Job_max           = 1000
	Job_ticker_number = 3000
	Job_ticker_going  = 150 // 1ds= 0.1s
	Job_ticker_expire = 50  // 1ds = 0.1s
)

var jobStartCounts = make([]uint64, Job_max)
var jobEndCounts = make([]uint64, Job_max)

type SimpleJob struct {
	Id  int
	Run func()
}

func (job *SimpleJob) Execute() {
	if job.Run == nil {
		log.Info("CRONJOB", "invalid")
		return
	}

	jobStartIndex := jobStartCounts[job.Id%Job_max]
	jobEndIndex := jobEndCounts[job.Id%Job_max]

	// guarantee only run if previous done
	if jobStartIndex > jobEndIndex {
		return
	}
	jobIndex := jobStartIndex
	atomic.AddUint64(&jobStartCounts[job.Id%Job_max], 1)
	startT := time.Now()
	log.Info("CRONJOB", "[id:", job.Id, ",index:", jobIndex, ",start:", startT.UnixNano(), "]")

	job.Run()

	atomic.AddUint64(&jobEndCounts[job.Id%Job_max], 1)
	endT := time.Now()
	log.Info("CRONJOB", "[id:", job.Id, ",index:", jobIndex, ",end:", endT.UnixNano(), ",status:", "done", "]")
}
func (job *SimpleJob) Description() string {
	return ""
}
func (job *SimpleJob) Key() int {
	return job.Id
}

type AccountJob struct {
	Id           int
	Run          func(*LogicAccount)
	LoadAccounts func() []*LogicAccount
	GoNum        int
}

func (job *AccountJob) Execute() {
	if job.Run == nil || job.LoadAccounts == nil {
		log.Info("CRONJOB", "invalid")
		return
	}

	jobStartIndex := jobStartCounts[job.Id%Job_max]
	jobEndIndex := jobEndCounts[job.Id%Job_max]

	// guarantee only run if previous done
	if jobStartIndex > jobEndIndex {
		return
	}
	jobIndex := jobStartIndex

	accounts := job.LoadAccounts()
	if accounts == nil || len(accounts) == 0 {
		return
	}

	atomic.AddUint64(&jobStartCounts[job.Id%Job_max], 1)
	startT := time.Now()
	log.Info("CRONJOB", "[id:", job.Id, ",index:", jobIndex, ",start:", startT.UnixNano(), "]")

	accountChan := make(chan *LogicAccount, len(accounts))
	for _, account := range accounts {
		accountChan <- account
	}
	close(accountChan)

	if job.GoNum < 1 {
		job.GoNum = 20
	}
	wg := &sync.WaitGroup{}
	wg.Add(job.GoNum)
	for i := 0; i < job.GoNum; i++ {
		go func(index int) {
			defer func() {
				wg.Done()
			}()
			for {
				account, has := <-accountChan
				if !has {
					break
				}
				job.Run(account)
			}
		}(i)
	}
	wg.Wait()

	atomic.AddUint64(&jobEndCounts[job.Id%Job_max], 1)
	endT := time.Now()
	log.Info("CRONJOB", "[id:", job.Id, ",index:", jobIndex, ",end:", endT.UnixNano(), ",status:", "done", "]")
}

func (job *AccountJob) Description() string {
	return ""
}

func (job *AccountJob) Key() int {
	return job.Id
}
