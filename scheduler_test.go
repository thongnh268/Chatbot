package main

import (
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

// 75μs
func BenchmarkExecute(b *testing.B) {
	accnum := 500
	accounts := make([]*LogicAccount, accnum)
	for i := 0; i < accnum; i++ {
		accounts[i] = &LogicAccount{Id: strconv.Itoa(i)}
	}
	job := &AccountJob{
		Id: 1,
		Run: func(acc *LogicAccount) {
			// fmt.Println(acc.Id)
		},
		LoadAccounts: func() []*LogicAccount {
			return accounts
		},
	}
	for n := 0; n < b.N; n++ {
		job.Execute()
	}
}

// basic case, exception case, fixed case
// input, expected output, actual output

func TestPush(t *testing.T) {
	var sched *Schedule
	var exoIdBatch, acoIdBatch []string

	sched = NewSchedule()
	sched.Push(time.Now().UnixNano()/1e6+1e3, "testacc")
	exoIdBatch = []string{"testacc"}
	acoIdBatch = sched.NextBatch(10)
	if !reflect.DeepEqual(acoIdBatch, exoIdBatch) {
		t.Errorf("want %#v, actual %#v", exoIdBatch, acoIdBatch)
	}

	sched = NewSchedule()
	var acoIdArr []string
	sched.BatchFunc = func(idMap []string, schedNow int64) {
		acoIdArr = idMap
	}
	sched.Push(time.Now().UnixNano()/1e6, "testacc1")
	time.Sleep(100 * time.Millisecond)
	exoIdArr := []string{"testacc1"}
	if !reflect.DeepEqual(acoIdArr, exoIdArr) {
		t.Errorf("want %#v, actual %#v", exoIdArr, acoIdArr)
	}

	sched = NewSchedule()
	sched.Push(time.Now().UnixNano()/1e6+300*1e3, "testacc")
	exoIdBatch = []string{}
	acoIdBatch = sched.NextBatch(3000)
	if !reflect.DeepEqual(acoIdBatch, exoIdBatch) {
		t.Errorf("want %#v, actual %#v", exoIdBatch, acoIdBatch)
	}
}

func TestPrepare(t *testing.T) {
	var sched *Schedule
	var exoIdBatch, acoIdBatch []string
	var accountIds []string

	sched = NewSchedule()
	accountIds = []string{"testacc1", "testacc2"}
	sched.Prepare(time.Now().UnixNano()/1e6+1e3, accountIds)
	exoIdBatch = []string{}
	acoIdBatch = sched.NextBatch(10)
	if !reflect.DeepEqual(acoIdBatch, exoIdBatch) {
		t.Errorf("want %#v, actual %#v", exoIdBatch, acoIdBatch)
	}

	sched = NewSchedule()
	accountIds = []string{"testacc1", "testacc2"}
	sched.Prepare(time.Now().UnixNano()/1e6+14*1e3, accountIds)
	exoIdBatch = []string{}
	acoIdBatch = sched.NextBatch(140)
	if !reflect.DeepEqual(acoIdBatch, exoIdBatch) {
		t.Errorf("want %#v, actual %#v", exoIdBatch, acoIdBatch)
	}

	sched = NewSchedule()
	accountIds = []string{"testacc1", "testacc2"}
	sched.Prepare(time.Now().UnixNano()/1e6+15*1e3, accountIds)
	exoIdBatch = []string{"testacc1", "testacc2"}
	acoIdBatch = sched.NextBatch(150)
	if !reflect.DeepEqual(acoIdBatch, exoIdBatch) {
		t.Errorf("want %#v, actual %#v", exoIdBatch, acoIdBatch)
	}

	sched = NewSchedule()
	accountIds = []string{"testacc1", "testacc2"}
	sched.Prepare(time.Now().UnixNano()/1e6+299*1e3, accountIds)
	exoIdBatch = []string{"testacc1", "testacc2"}
	acoIdBatch = sched.NextBatch(2990)
	if !reflect.DeepEqual(acoIdBatch, exoIdBatch) {
		t.Errorf("want %#v, actual %#v", exoIdBatch, acoIdBatch)
	}

	sched = NewSchedule()
	accountIds = []string{"testacc1", "testacc2"}
	sched.Prepare(time.Now().UnixNano()/1e6+300*1e3, accountIds)
	exoIdBatch = []string{}
	acoIdBatch = sched.NextBatch(3000)
	if !reflect.DeepEqual(acoIdBatch, exoIdBatch) {
		t.Errorf("want %#v, actual %#v", exoIdBatch, acoIdBatch)
	}
}

func TestBatchAsync(t *testing.T) {
	var sched *Schedule
	sched = NewSchedule()
	head := sched.head
	idArrs := make([][]string, 5)
	for i := 0; i < 5; i++ {
		idArrs[i] = randomAccountIds(1000)
		sched.idBatch[(head+i+1)%3000] = idArrs[i]
	}
	idMapLock := &sync.Mutex{}
	outIdArrs := make([][]string, 0)
	sched.BatchFunc = func(idArr []string, schedNow int64) {
		idMapLock.Lock()
		if len(idArr) > 0 {
			outIdArrs = append(outIdArrs, idArr)
		}
		idMapLock.Unlock()
	}
	sched.BatchAsync()
	time.Sleep(600 * time.Millisecond)
	if len(idArrs) != len(outIdArrs) {
		t.Errorf("diff len(aco)!=len(exo) %v!=%v", len(outIdArrs), len(idArrs))
	}
	for i := 0; i < len(idArrs) && i < len(outIdArrs); i++ {
		if !reflect.DeepEqual(outIdArrs[i], idArrs[i]) {
			t.Errorf("diff at %v", i)
		}
	}
}

func randomAccountIds(size int) []string {
	n := int(rand.Int63n(int64(size)))
	arr := make([]string, n)
	for i := 0; i < n; i++ {
		arr[i] = strconv.Itoa(i)
	}
	return arr
}
