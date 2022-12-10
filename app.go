package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/init", InitWallet).Methods("POST")
	router.HandleFunc("/api/v1/wallet", EnableWallet).Methods("POST")
	router.HandleFunc("/api/v1/wallet", GetBalance).Methods("GET")
	router.HandleFunc("/api/v1/wallet", DisableWallet).Methods("PATCH")
	router.HandleFunc("/api/v1/wallet/deposits", Deposit).Methods("POST")
	router.HandleFunc("/api/v1/wallet/withdrawals", Withdraw).Methods("POST")

	fmt.Println("Listening on 8080...")
	http.ListenAndServe(":8080", router)
}
