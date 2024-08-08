package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	USD_BRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", handleCotacao)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleCotacao(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var rate ExchangeRate
	if err := json.NewDecoder(resp.Body).Decode(&rate); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go saveRate(rate.USD_BRL.Bid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"bid": rate.USD_BRL.Bid})
}

func saveRate(bid string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Println("Failed to connect to the database:", err)
		return
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS exchange_rates (id INTEGER PRIMARY KEY, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)"); err != nil {
		log.Println("Failed to create table:", err)
		return
	}

	if _, err := db.ExecContext(ctx, "INSERT INTO exchange_rates (bid) VALUES (?)", bid); err != nil {
		log.Println("Failed to insert rate into database:", err)
	}
}
