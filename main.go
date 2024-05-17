package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	port             = ":8080"
	exchangeRateAPI  = "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?json"
	sendGridAPIKey   = "your-sendgrid-api-key"
	postgresConnInfo = "postgres://username:password@localhost/dbname?sslmode=disable"
)

var db *sql.DB

type ExchangeRate struct {
	Rate float64 `json:"rate"`
	CC   string  `json:"cc"`
}

func main() {
	http.HandleFunc("/rate", getRateHandler)
	//http.HandleFunc("/subscribe", subscribeHandler)

	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func getRateHandler(w http.ResponseWriter, r *http.Request) {
	rate, err := getExchangeRate()
	if err != nil {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	response := map[string]float64{"rate": rate}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}
}

func getExchangeRate() (float64, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, _ := client.Get(exchangeRateAPI)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	var rates []ExchangeRate

	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return 0, err
	}

	for i := 0; i < len(rates); i++ {
		if rates[i].CC == "USD" {
			return rates[i].Rate, nil
		}
	}

	return 0, nil
}
