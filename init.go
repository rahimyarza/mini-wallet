package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "postgres"
	DB_NAME     = "miniwallet"
)

func setupDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}

	return db
}

type JsonResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type Wallet struct {
	Xid        string     `json:"xid"`
	Token      string     `json:"token"`
	Wid        string     `json:"wid"`
	Balance    int        `json:"balance"`
	EnabledAt  *time.Time `json:"enabled_at"`
	DisabledAt *time.Time `json:"disabled_at"`
	Status     bool       `json:"is_enabled"`
}

type Depo struct {
	Wid         string    `json:"id"`
	Xid         string    `json:"deposited_by"`
	DepositedAt time.Time `json:"deposited_at"`
	Status      string    `json:"status"`
	Balance     int       `json:"amount"`
	RefID       string    `json:"reference_id"`
}

type Withdrawal struct {
	Wid        string    `json:"id"`
	Xid        string    `json:"withdrawn_by"`
	WithdrawAt time.Time `json:"withdrawn_at"`
	Status     string    `json:"status"`
	Balance    int       `json:"amount"`
	RefID      string    `json:"reference_id"`
}

func InitWallet(w http.ResponseWriter, r *http.Request) {
	var response = JsonResponse{}
	var wallet = Wallet{}
	db := setupDB()
	defer db.Close()
	wallet.Xid = r.FormValue("customer_xid")
	if wallet.Xid == "" {
		response.Status = "fail"
		response.Message = "Missing data for required field."
	} else {
		row := db.QueryRow("SELECT token FROM wallet WHERE xid = $1", wallet.Xid)

		switch err := row.Scan(&wallet.Token); err {
		case sql.ErrNoRows:
			wallet.Token = tokenGenerator()
			_, err := db.Exec("INSERT INTO wallet (token, xid) VALUES ($1, $2)", wallet.Token, wallet.Xid)
			if err != nil {
				response.Status = "error"
				response.Message = "Error Execute DB"
				break
			}
			response.Status = "success"
			response.Data = wallet.Token
		case nil:
			response.Status = "success"
			response.Data = wallet.Token
		default:
			response.Status = "error"
			response.Message = "Error Query DB"
		}
	}

	json.NewEncoder(w).Encode(response)
}

func EnableWallet(w http.ResponseWriter, r *http.Request) {
	var response = JsonResponse{}
	var wallet = Wallet{}
	auth := r.Header.Get("Authorization")
	wallet.Token = strings.Trim(auth, "Token ")
	db := setupDB()
	defer db.Close()
	if !checkToken(db, wallet.Token) {
		response.Status = "fail"
		response.Message = "Authentication error"
	} else {
		row := db.QueryRow("SELECT wid, xid, is_enabled, balance FROM wallet WHERE token = $1", wallet.Token)
		switch err := row.Scan(&wallet.Wid, &wallet.Xid, &wallet.Status, &wallet.Balance); err {
		case nil:
			if !wallet.Status {
				if wallet.Wid == "" {
					wallet.Wid = widGenerator()
				}
				now := time.Now()
				wallet.EnabledAt = &now
				wallet.Status = true
				_, err := db.Exec("UPDATE wallet SET is_enabled = $1, wid = $2 WHERE token = $3", wallet.Status, wallet.Wid, wallet.Token)
				if err != nil {
					response.Status = "error"
					response.Message = "Error Execute DB"
					break
				}
				response.Status = "success"
				response.Data = wallet
			} else {
				response.Status = "fail"
				response.Message = "Already enabled"
			}
		default:
			response.Status = "error"
			response.Message = "Internal error"
		}
	}
	json.NewEncoder(w).Encode(response)

}

func GetBalance(w http.ResponseWriter, r *http.Request) {
	var response = JsonResponse{}
	var wallet = Wallet{}
	auth := r.Header.Get("Authorization")
	wallet.Token = strings.Trim(auth, "Token ")
	db := setupDB()
	defer db.Close()
	if !checkToken(db, wallet.Token) {
		response.Status = "fail"
		response.Message = "Authentication error"
	} else if !checkIsActiveWallet(db, wallet.Token) {
		response.Status = "fail"
		response.Message = "Disabled"
	} else {
		row := db.QueryRow("SELECT wid, xid, is_enabled, balance FROM wallet WHERE token = $1", wallet.Token)
		switch err := row.Scan(&wallet.Wid, &wallet.Xid, &wallet.Status, &wallet.Balance); err {
		case nil:
			response.Status = "success"
			response.Data = wallet
		default:
			response.Status = "error"
			response.Message = "Internal error"
		}
	}
	json.NewEncoder(w).Encode(response)
}

func Deposit(w http.ResponseWriter, r *http.Request) {
	// check db if wallet enabled
	var response = JsonResponse{}
	var wallet = Wallet{}
	auth := r.Header.Get("Authorization")
	deposit := r.FormValue("amount")
	depo, _ := strconv.Atoi(deposit)
	wallet.Token = strings.Trim(auth, "Token ")
	db := setupDB()
	defer db.Close()
	if !checkToken(db, wallet.Token) {
		response.Status = "fail"
		response.Message = "Authentication error"
	} else if !checkIsActiveWallet(db, wallet.Token) {
		response.Status = "fail"
		response.Message = "Disabled"
	} else {
		row := db.QueryRow("SELECT xid, wid, balance FROM wallet WHERE token = $1", wallet.Token)
		switch err := row.Scan(&wallet.Xid, &wallet.Wid, &wallet.Balance); err {
		case nil:
			_, err := db.Exec("UPDATE wallet SET balance = $1 WHERE token = $2", depo+wallet.Balance, wallet.Token)
			if err != nil {
				response.Status = "error"
				response.Message = "Error Execute DB"
			} else {
				response.Status = "success"
				response.Data = Depo{
					Wid:         wallet.Wid,
					Xid:         wallet.Xid,
					DepositedAt: time.Now(),
					Status:      "success",
					Balance:     depo,
					RefID:       widGenerator(),
				}
			}
		default:
			fmt.Println(err)
			response.Status = "error"
			response.Message = "Internal error"
		}
	}
	json.NewEncoder(w).Encode(response)
}

func Withdraw(w http.ResponseWriter, r *http.Request) {
	// check db if wallet enabled
	var response = JsonResponse{}
	var wallet = Wallet{}
	auth := r.Header.Get("Authorization")
	withdraw := r.FormValue("amount")
	wd, _ := strconv.Atoi(withdraw)
	wallet.Token = strings.Trim(auth, "Token ")
	db := setupDB()
	defer db.Close()
	if !checkToken(db, wallet.Token) {
		response.Status = "fail"
		response.Message = "Authentication error"
	} else if !checkIsActiveWallet(db, wallet.Token) {
		response.Status = "fail"
		response.Message = "Disabled"
	} else {
		row := db.QueryRow("SELECT xid, wid, balance FROM wallet WHERE token = $1", wallet.Token)
		switch err := row.Scan(&wallet.Xid, &wallet.Wid, &wallet.Balance); err {
		case nil:
			if wd > wallet.Balance {
				response.Status = "fail"
				response.Message = "Balance is insufficient"
			} else {
				_, err := db.Exec("UPDATE wallet SET balance = $1 WHERE token = $2", wallet.Balance-wd, wallet.Token)
				if err != nil {
					response.Status = "error"
					response.Message = "Error Execute DB"
				} else {
					response.Status = "success"
					response.Data = Withdrawal{
						Wid:        wallet.Wid,
						Xid:        wallet.Xid,
						WithdrawAt: time.Now(),
						Status:     "success",
						Balance:    wd,
						RefID:      widGenerator(),
					}
				}
			}

		default:
			fmt.Println(err)
			response.Status = "error"
			response.Message = "Internal error"
		}
	}
	json.NewEncoder(w).Encode(response)
}

func DisableWallet(w http.ResponseWriter, r *http.Request) {
	var response = JsonResponse{}
	var wallet = Wallet{}
	auth := r.Header.Get("Authorization")
	wallet.Token = strings.Trim(auth, "Token ")
	db := setupDB()
	defer db.Close()
	if !checkToken(db, wallet.Token) {
		response.Status = "fail"
		response.Message = "Authentication error"
	} else {
		row := db.QueryRow("SELECT xid, is_enabled, balance FROM wallet WHERE token = $1", wallet.Token)
		switch err := row.Scan(&wallet.Xid, &wallet.Status, &wallet.Balance); err {
		case nil:
			if wallet.Status {
				now := time.Now()
				wallet.DisabledAt = &now
				wallet.Status = false
				_, err := db.Exec("UPDATE wallet SET is_enabled = $1 WHERE token = $2", wallet.Status, wallet.Token)
				if err != nil {
					response.Status = "error"
					response.Message = "Error Execute DB"
					break
				}
				response.Status = "success"
				response.Data = wallet
			} else {
				response.Status = "fail"
				response.Message = "Already disabled"
			}
		default:
			response.Status = "error"
			response.Message = "Internal error"
		}
	}
	json.NewEncoder(w).Encode(response)
}
