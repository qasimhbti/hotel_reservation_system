package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize("root", "root", "restaurant_api")

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(layoutTableCreationQuery); err != nil {
		log.Fatal(err)
	}

	if _, err := a.DB.Exec(seatingTableCreationQuery); err != nil {
		log.Fatal(err)
	}

	if _, err := a.DB.Exec(totTableCreationQuery); err != nil {
		log.Fatal(err)
	}

	if _, err := a.DB.Exec(customersTableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM layout")
	a.DB.Exec("DELETE FROM seating")
	a.DB.Exec("DELETE FROM tot")
	a.DB.Exec("DELETE FROM customers")
	a.DB.Exec("ALTER TABLE customerss AUTO_INCREMENT = 1")
}

const layoutTableCreationQuery = `
CREATE TABLE IF NOT EXISTS layout
(
	table_id INT PRIMARY KEY,
	max_seating INT NOT NULL,
	availability VARCHAR(10) NOT NULL 
)`

const seatingTableCreationQuery = `
CREATE TABLE IF NOT EXISTS seating
(
	seating_capacity INT PRIMARY KEY,
	min_occupency INT NULL
)`

const totTableCreationQuery = `
CREATE TABLE IF NOT EXISTS tot
(
	min_party_size INT PRIMARY KEY,
	max_party_size INT NULL DEFAULT 999,
	avg_tot VARCHAR(10) NOT NULL 
)`

const customersTableCreationQuery = `
CREATE TABLE IF NOT EXISTS customers
(
	booking_id INT AUTO_INCREMENT PRIMARY KEY,
	table_id INT NOT NULL,
	name VARCHAR(45) NULL,
	booking_date VARCHAR(45) NULL,
	booking_time INT NULL,
	party_size INT NULL,
	phone VARCHAR(45) NULL
)`

func TestEmptyTable(t *testing.T) {
	clearTable()

	tableID := 0
	maxSeating := 0
	statement := fmt.Sprintf("SELECT * FROM layout")
	_ = a.DB.QueryRow(statement).Scan(&tableID, &maxSeating)

	if tableID != 0 {
		t.Errorf("Expected an empty value. Got %d", tableID)
	}
}

func TestGetReservation(t *testing.T) {
	clearTable()

	payload := []byte(`{"name":"test","phone":"9560291908", "party_size":10}`)

	cT := time.Now()
	reqAddress := fmt.Sprintf("/booking/%d-%d-%d/%d:%d", cT.Year(), int(cT.Month()), cT.Day(), cT.Hour(), cT.Minute())
	req, _ := http.NewRequest("POST", reqAddress, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestProcessCheckOut(t *testing.T) {
	clearTable()
	var availability string
	statement := fmt.Sprintf("INSERT INTO layout(table_id, max_seating, availability) VALUES('%d', '%d', '%s')", 0, 0, "N")
	_, _ = a.DB.Exec(statement)

	req, _ := http.NewRequest("POST", "/checkout/0", bytes.NewBuffer([]byte{}))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	statement = fmt.Sprintf("SELECT availability FROM layout WHERE table_id = %d", 0)
	_ = a.DB.QueryRow(statement).Scan(&availability)

	if availability != "Y" {
		t.Errorf("Expected availability as 'Y', Got %s", availability)
	}
	clearTable()
}
