package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	postgresConnInfo = os.Getenv("DB_URL")
	db               *sql.DB
)

const (
	port            = ":8080"
	exchangeRateAPI = "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?json"
)

type ExchangeRate struct {
	Rate float64 `json:"rate"`
	CC   string  `json:"cc"`
}

func main() {

	if err := dbInitialize(); err != nil {
		log.Fatalf("Unable to run migrations: %v\n", err)
	}

	cronJob := cron.New()
	cronJob.AddFunc("@every 1m", sendDailyExchangeRateEmails)
	cronJob.Start()
	http.HandleFunc("/rate", getRateHandler)
	http.HandleFunc("/subscribe", subscribeHandler)

	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func dbInitialize() error {
	var err error

	//os.Setenv("ETHEREAL_EMAIL", "chelsie.boehm6@ethereal.email")
	//os.Setenv("ETHEREAL_PASSWORD", "GaZbAHrhFF7JB6paH2")
	//os.Setenv("DB_URL", "postgres://postgres:Dmytry090302@localhost:5432/Genesis?sslmode=disable")

	db, err = sql.Open("pgx", os.Getenv("DB_URL"))
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migration",
		"Genesis", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
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

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if err := addSubscriber(email); err != nil {
		http.Error(w, "Email already exists or there is a problem appears", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string]string{"message": "Subscribed successfully"})
	if err != nil {
		return
	}
}

func emailSender(rate float64, to string) error {

	from := os.Getenv("ETHEREAL_EMAIL")
	password := os.Getenv("ETHEREAL_PASSWORD")

	if from == "" || password == "" {
		return errors.New("ETHEREAL_EMAIL or ETHEREAL_PASSWORD is not set")
	}

	subject := "Daily Exchange Rate"
	body := fmt.Sprintf("Current USD to UAH exchange rate: %.2f", rate)

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, "smtp.ethereal.email")
	err := smtp.SendMail("smtp.ethereal.email:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Fatal(err)
		return err
	}
	return err
}

func addSubscriber(email string) error {
	query := `INSERT INTO subscriptions (email) VALUES ($1)`
	_, err := db.Exec(query, email)
	return err
}

func getSubscribers() ([]string, error) {
	rows, err := db.Query("SELECT email FROM subscriptions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func sendDailyExchangeRateEmails() {
	rate, err := getExchangeRate()
	if err != nil {
		log.Println("Error fetching exchange rate:", err)
		return
	}

	subscribers, err := getSubscribers()
	if err != nil {
		log.Println("Error fetching subscribers:", err)
		return
	}

	for _, subscriber := range subscribers {
		err := emailSender(rate, subscriber)
		if err != nil {
			log.Printf("Failed to send email to %s: %v\n", subscriber, err)
		}
	}
}
