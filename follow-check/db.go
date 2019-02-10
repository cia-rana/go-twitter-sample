package main

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type databaseManager struct {
	db    *sql.DB
	dbMtx sync.Mutex
	name  string
}

func (dm *databaseManager) close() {
	dm.db.Close()
	dm.db = nil
}

func (dm *databaseManager) lock() {
	dm.dbMtx.Lock()
}

func (dm *databaseManager) unlock() {
	dm.dbMtx.Unlock()
}

func (a *app) openDB() error {
	db, err := sql.Open("sqlite3", a.dm.name)
	if err != nil {
		return err
	}

	a.dm.db = db

	err = a.createDB()
	if err != nil {
		a.dm.close()
		return err
	}

	return nil
}

func (a *app) closeDB() {
	a.dm.lock()
	defer a.dm.unlock()

	a.dm.close()
}

func (a *app) createDB() error {
	a.dm.lock()
	defer a.dm.unlock()

	if _, err := a.dm.db.Exec(`
	CREATE TABLE IF NOT EXISTS friends (
		id          INTEGER NOT NULL,
		screen_name TEXT,
		name        TEXT,
		protected   INTEGER,
		verified    INTEGER,
		created_at  INTEGER NOT NULL,
		PRIMARY KEY(id, created_at)
	)
	`); err != nil {
		return err
	}

	if _, err := a.dm.db.Exec(`
	CREATE TABLE IF NOT EXISTS followers (
		id          INTEGER NOT NULL,
		screen_name TEXT,
		name        TEXT,
		protected   INTEGER,
		verified    INTEGER,
		created_at  INTEGER NOT NULL,
		PRIMARY KEY(id, created_at)
	)
	`); err != nil {
		return err
	}

	return nil
}

func (a app) saveFriends(friends []user) error {
	a.dm.lock()
	defer a.dm.unlock()

	log.Println("insert friends")

	tx, err := a.dm.db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}

	stmt, err := a.dm.db.Prepare(`
	INSERT INTO friends(
		id,
		screen_name,
		name,
		protected,
		verified,
		created_at
	) values (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	for _, friend := range friends {
		_, err = stmt.Exec(
			friend.id,
			friend.screenName,
			friend.name,
			btoi(friend.protected),
			btoi(friend.verified),
			friend.createdAt.Unix(),
		)
	}
	tx.Commit()

	return nil
}

func (a app) saveFollowers(followers []user) error {
	a.dm.lock()
	defer a.dm.unlock()

	log.Println("insert followers")

	tx, err := a.dm.db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}

	stmt, err := a.dm.db.Prepare(`
	INSERT INTO followers(
		id,
		screen_name,
		name,
		protected,
		verified,
		created_at
	) values(?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	for _, follower := range followers {
		_, err = stmt.Exec(
			follower.id,
			follower.screenName,
			follower.name,
			btoi(follower.protected),
			btoi(follower.verified),
			follower.createdAt.Unix(),
		)
	}
	tx.Commit()

	return nil
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
