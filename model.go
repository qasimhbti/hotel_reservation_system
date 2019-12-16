package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

type reservationDetail struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	PartySize int    `json:"party_size"`
}

type reservationResult struct {
	TableID     int       `json:"table_id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	BookingDate time.Time `json:"booking_date"`
	BookingTime int64     `json:"booking_time"`
	Duration    int       `json:"duration"`
	NoOfSeats   int       `json:"no_of_seats"`
	SchCoutDur  int64     `json:"sch_checkout_duration"`
}

func (r *reservationResult) parseTime(vars map[string]string) error {
	t := strings.Split(vars["time"], ":")
	timeHour, err := strconv.Atoi(t[0])
	if err != nil {
		return errors.New("Invalid Time")
	}

	if timeHour < 10 || timeHour > 21 {
		return errors.New("Booking can be done between 10AM to 10PM")
	}

	return nil
}

func (r *reservationResult) parseDate(vars map[string]string) error {
	currentTime := time.Now()
	d, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		return errors.New("Invalid Date")
	}

	t := strings.Split(vars["time"], ":")
	timeHour, err := strconv.Atoi(t[0])
	if err != nil {
		return errors.New("Invalid Time")
	}
	timeMin, err := strconv.Atoi(t[1])
	if err != nil {
		return errors.New("Invalid Time")
	}

	bookingTime := time.Date(d.Year(), d.Month(), d.Day(), timeHour, timeMin, 0, 0, time.Local)

	if (bookingTime.Unix() < (currentTime.Add(time.Hour * 2)).Unix()) || (bookingTime.Unix() > (currentTime.AddDate(0, 0, 2)).Unix()) {
		return errors.New("Reservation can be made only 2 to 48 hrs prior")
	}

	r.BookingDate = bookingTime
	r.BookingTime = bookingTime.Unix()
	r.SchCoutDur = r.BookingTime - currentTime.Unix()
	return nil
}

type tableMaxSeating struct {
	TableID    int
	MaxSeating int
}

func (r *reservationResult) getAvgTOT(db *sql.DB) error {
	var avgtots string
	statement := fmt.Sprintf("SELECT avg_tot FROM tot WHERE min_party_size <= %d AND max_party_size >= %d", r.NoOfSeats, r.NoOfSeats)
	err := db.QueryRow(statement).Scan(&avgtots)
	if err != nil {
		fmt.Println("No such table is present in restaurant")
		return errors.New("No such table is present in restaurant")
	}

	fmt.Println("avgtots :", avgtots)
	avgtot := strings.Split(avgtots, "m")
	tot, err := strconv.Atoi(avgtot[0])
	if err != nil {
		return errors.New("avgtot is not valid")
	}
	r.Duration = tot * 60
	r.SchCoutDur += int64(r.Duration)
	return nil
}

func (r *reservationResult) checkTableAvailability(db *sql.DB) error {
	tableWithMaxSeating := []*tableMaxSeating{}
	statement := fmt.Sprintf("SELECT table_id, max_seating FROM layout WHERE max_seating >= '%d'", r.NoOfSeats)
	rows, err := db.Query(statement)
	if err != nil {
		return errors.New("Table is not Available")
	}
	defer rows.Close()

	for rows.Next() {
		tableSeating := new(tableMaxSeating)
		if err := rows.Scan(&tableSeating.TableID, &tableSeating.MaxSeating); err != nil {
			fmt.Println("Error while accessing tableid and maxseating")
		}
		tableWithMaxSeating = append(tableWithMaxSeating, tableSeating)
	}

	availableTables := []*tableMaxSeating{}
	for i := range tableWithMaxSeating {
		if tableWithMaxSeating[i].MaxSeating == 1 {
			availableTables = append(availableTables, tableWithMaxSeating[i])
			continue
		}

		minOccupency := 0
		statement = fmt.Sprintf("SELECT min_occupency FROM seating WHERE seating_capacity = '%d'", tableWithMaxSeating[i].MaxSeating)
		err = db.QueryRow(statement).Scan(&minOccupency)
		if err != nil {
			return errors.New("No such table is present in restaurant")
		}
		//fmt.Println("minOccupency :", minOccupency)

		if r.NoOfSeats >= minOccupency {
			availableTables = append(availableTables, tableWithMaxSeating[i])
		}
	}

	for i := range availableTables {
		fmt.Println("availableTables :", availableTables[i])
	}

	bookingDetail := &bookingDetails{
		BDate: r.BookingDate,
		BTime: r.BookingTime,
		CTime: r.BookingTime + int64(r.Duration),
	}

	tableID := 0
	for i := range availableTables {
		tableID = bookingDetail.checkTableAvailiblityTimeSlot(availableTables[i].TableID)
		if tableID == availableTables[i].TableID {
			break
		}
	}

	if tableID == -1 {
		return errors.New("Table is not available")
	}

	bookingDetail.updateReserveTableStatus(tableID)
	r.TableID = tableID
	return nil
}

func (r *reservationResult) confirmBooking(db *sql.DB) error {
	statement := fmt.Sprintf("INSERT INTO customers(name, table_id, booking_date, booking_time, party_size, phone) VALUES('%s', '%d', '%s', '%d', '%d', '%s')", r.Name, r.TableID, r.BookingDate, r.BookingTime, r.NoOfSeats, r.Phone)
	_, err := db.Exec(statement)
	if err != nil {
		return errors.New("Error while booking -CustomerDB")
	}
	return nil
}

func (r *reservationResult) processCheckOut(db *sql.DB) error {
	statement := fmt.Sprintf("UPDATE layout SET availability = '%s' WHERE table_id = %d", "Y", r.TableID)
	_, err := db.Exec(statement)
	if err != nil {
		return errors.New("Error while CheckOut")
	}
	return nil
}

func (r *reservationResult) scheduleCheckOut(db *sql.DB) error {
	ticker := time.NewTicker(time.Duration(r.SchCoutDur) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				processCheckOut(db, r)
				ticker.Stop()
			}
		}
	}()
	return nil
}
