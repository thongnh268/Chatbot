package main

import (
	"fmt"

	"github.com/subiz/cassandra"
	"github.com/urfave/cli"
)

func resetDB(ctx *cli.Context) {
	// drop keyspace by gocql
	// dropDBView(ctx)
	dropDB(ctx)
	createDB(ctx)
	// createDBView(ctx)
}

func createDB(ctx *cli.Context) {
	keyspace := bizbot.Cf.KeyspacePrefix + Bizbot_keyspace
	err := cassandra.CreateKeyspace(bizbot.Cf.CassandraSeeds, keyspace, bizbot.Cf.ReplicaFactor)
	if err != nil {
		panic(err)
	}
	cql := &cassandra.Query{}
	err = cql.Connect(bizbot.Cf.CassandraSeeds, keyspace)
	if err != nil {
		panic(err)
	}

	err = cql.Session.Query(`CREATE TABLE IF NOT EXISTS ` + Bizbot_tbl_bot + `(
		account_id ASCII,
		id ASCII,
		bot BLOB,
		fullname TEXT,
		state ASCII,
		report BLOB,
		PRIMARY KEY ((account_id), id)
	)`).Exec()
	if err != nil {
		panic(err)
	}

	err = cql.Session.Query(`CREATE TABLE IF NOT EXISTS ` + Bizbot_tbl_bot_tag + `(
		account_id ASCII,
		id ASCII,
		tag ASCII,
		bot BLOB,
		fullname TEXT,
		state ASCII,
		PRIMARY KEY ((account_id, id), tag)
	)`).Exec()
	if err != nil {
		panic(err)
	}

	err = cql.Session.Query(`CREATE TABLE IF NOT EXISTS ` + Bizbot_tbl_exec_bot + `(
		account_id ASCII,
		bot_id ASCII,
		obj_type ASCII,
		obj_id ASCII,
		lead_id ASCII,
		lead_payload MAP<ASCII,ASCII>,
		id ASCII,
		bot BLOB,
		status ASCII,
		status_note ASCII,
		run_times LIST<BIGINT>,
		created BIGINT,
		updated BIGINT,
		last_seen BIGINT,
		last_evt BLOB,
		action_tree BLOB,
		sched_acts MAP<ASCII,BIGINT>,
		PRIMARY KEY (account_id, bot_id, id)
	)`).Exec()
	if err != nil {
		panic(err)
	}

	err = cql.Session.Query(`CREATE TABLE IF NOT EXISTS ` + Bizbot_tbl_exec_bot_index + `(
		account_id ASCII,
		bot_id ASCII,
		obj_type ASCII,
		obj_id ASCII,
		lead_id ASCII,
		lead_payload MAP<ASCII,ASCII>,
		id ASCII,
		index_incre BIGINT,
		begin_time_day BIGINT,
		begin_time BIGINT,
		end_time BIGINT,
		created BIGINT,
		updated BIGINT,
		exec_bot BLOB,
		PRIMARY KEY ((account_id, bot_id, id), index_incre)
	)`).Exec()
	if err != nil {
		panic(err)
	}
}

func dropDB(clictx *cli.Context) {
	var err error
	keyspace := bizbot.Cf.KeyspacePrefix + Bizbot_keyspace
	conn := &cassandra.Query{}
	err = conn.Connect(bizbot.Cf.CassandraSeeds, keyspace)
	if err != nil {
		panic(err)
	}
	session := conn.Session

	err = session.Query(fmt.Sprintf(`
		DROP TABLE IF EXISTS %s;
	`, Bizbot_tbl_bot)).Exec()
	if err != nil {
		panic(err)
	}

	err = session.Query(fmt.Sprintf(`
		DROP TABLE IF EXISTS %s;
	`, Bizbot_tbl_exec_bot)).Exec()
	if err != nil {
		panic(err)
	}

	err = session.Query(fmt.Sprintf(`
		DROP TABLE IF EXISTS %s;
	`, Bizbot_tbl_exec_bot_index)).Exec()
	if err != nil {
		panic(err)
	}
}

func createDBView(clictx *cli.Context) {
	keyspace := bizbot.Cf.KeyspacePrefix + Bizbot_keyspace
	cql := &cassandra.Query{}
	err := cql.Connect(bizbot.Cf.CassandraSeeds, keyspace)
	if err != nil {
		panic(err)
	}

	err = cql.Session.Query(fmt.Sprintf(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS %s
		AS SELECT *
		FROM %s
		WHERE account_id IS NOT NULL AND bot_id IS NOT NULL AND begin_time_day IS NOT NULL AND id IS NOT NULL AND index_incre IS NOT NULL
		PRIMARY KEY ((account_id, bot_id, begin_time_day), id, index_incre)
	`, Bitbot_tbl_exec_bot_index_view, Bizbot_tbl_exec_bot_index)).Exec()
	if err != nil {
		panic(err)
	}

	err = cql.Session.Query(fmt.Sprintf(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS %s
		AS SELECT *
		FROM %s
		WHERE account_id IS NOT NULL AND id IS NOT NULL AND tag IS NOT NULL
		PRIMARY KEY ((account_id, tag), id)
	`, Bizbot_tbl_bot_tag_view, Bizbot_tbl_bot_tag)).Exec()
}

func dropDBView(clictx *cli.Context) {
	keyspace := bizbot.Cf.KeyspacePrefix + Bizbot_keyspace
	conn := &cassandra.Query{}
	err := conn.Connect(bizbot.Cf.CassandraSeeds, keyspace)
	if err != nil {
		panic(err)
	}
	session := conn.Session

	err = session.Query(fmt.Sprintf(`
		DROP MATERIALIZED VIEW %s;
	`, Bitbot_tbl_exec_bot_index_view)).Exec()
	if err != nil {
		panic(err)
	}
}
