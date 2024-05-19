# Currency Exchange Rate App

This is a Go application that fetches the current USD to UAH exchange rate from the NBU API and sends daily email notifications to subscribed users with the latest exchange rate.

## IMPORTANT!!!
In this project I have been using smtp emulated server to test the email sending functionality. https://ethereal.email/ is the emulated server have used in this project.

## Running Locally

1. Install Go:
   [Go install](https://golang.org/doc/install)
2. Set the following environment variables:
    - `ETHEREAL_EMAIL`: Your Ethereal email address for sending emails
    - `ETHEREAL_PASSWORD`: Your Ethereal password for sending emails
    - `DB_URL`: The connection string for your PostgreSQL database

3. Make sure you have a PostgreSQL database running and accessible with the provided connection string.

4. Run the application using:
   ```bash
   go run main.go
    ```

## Docker

1. Make sure you have Docker and Docker Compose installed on your system.

2. Build and run the application using Docker Compose:
    ```bash
    docker-compose build
    docker-compose up
    ```
   
The application will start a server on `http://localhost:8080`. You can subscribe to receive daily exchange rate emails by sending a POST request to `http://localhost:8080/subscribe` with the `email` parameter in the request body.