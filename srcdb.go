package main

import (
	"github.com/gocql/gocql"
	"github.com/subiz/cassandra"
	"github.com/subiz/errors"
	pb "github.com/subiz/header/account"
	ppb "github.com/subiz/header/payment"
)

type AccDB struct {
	session *gocql.Session
	cql     cassandra.Query
}

func NewAccDB(seeds []string, keyspaceprefix string) *AccDB {
	db := &AccDB{}
	db.cql = cassandra.Query{}
	err := db.cql.Connect(seeds, keyspaceprefix)
	if err != nil {
		panic(err)
	}
	db.session = db.cql.Session
	return db
}

func (db *AccDB) ReadAccount(id string) (*pb.Account, error) {
	acc := &pb.Account{}
	// if cache not exist re get from db
	err := db.cql.Read(Srcacc_tbl_accounts, acc, pb.Account{Id: &id})
	if err != nil && err.Error() == gocql.ErrNotFound.Error() {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, id)
	}

	// acc.Limit, _ = me.GetLimit(id)
	return acc, nil
}

type PaymentDB struct {
	session *gocql.Session
	cql     cassandra.Query
}

func NewPaymentDB(seeds []string, keyspaceprefix string) *PaymentDB {
	db := &PaymentDB{}
	db.cql = cassandra.Query{}
	err := db.cql.Connect(seeds, keyspaceprefix)
	if err != nil {
		panic(err)
	}
	db.session = db.cql.Session
	return db
}

func (db *PaymentDB) GetSubscription(accid string) (*ppb.Subscription, error) {
	p := &ppb.Subscription{}
	err := db.cql.Read(Srcpayment_tbl_subscription, p, ppb.Subscription{AccountId: &accid})

	if err != nil && err.Error() == gocql.ErrNotFound.Error() {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, accid)
	}
	return p, nil
}

const (
	Srcacc_keyspace             = "accounts"
	Srcacc_tbl_accounts         = "accounts"
	Srcpayment_keyspace         = "payment"
	Srcpayment_tbl_subscription = "subs"
)
