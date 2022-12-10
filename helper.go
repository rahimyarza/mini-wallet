package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
)

func tokenGenerator() string {
	b := make([]byte, 21)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func widGenerator() string {
	a := make([]byte, 5)
	b := make([]byte, 2)
	c := make([]byte, 2)
	d := make([]byte, 2)
	e := make([]byte, 6)

	rand.Read(a)
	rand.Read(b)
	rand.Read(c)
	rand.Read(d)
	rand.Read(e)

	return fmt.Sprintf("%x-%x-%x-%x-%x", a, b, c, d, e)
}

func checkToken(db *sql.DB, token string) bool {
	var xid string
	row := db.QueryRow("SELECT xid FROM wallet WHERE token = $1", token)
	if row.Scan(&xid) != nil {
		return false
	}

	return true
}

func checkIsActiveWallet(db *sql.DB, token string) bool {
	var isEnabled bool
	row := db.QueryRow("SELECT is_enabled FROM wallet WHERE token = $1", token)
	if row.Scan(&isEnabled) != nil || !isEnabled {
		return false
	}
	return true
}
