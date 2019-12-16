package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Seating is use to represent seating configuration
type Seating struct {
	SeatingCapacity int `json:"seating_capacity"`
	MinOccupency    int `json:"min_occupency"`
}

type seatingConfigManager interface {
	get(DB *sql.DB) ([]Seating, error)
	insert(DB *sql.DB, seating []Seating) error
}

type seatingConfigManagerImpl struct {
}

func (s seatingConfigManagerImpl) get(DB *sql.DB) ([]Seating, error) {

	seatingFile, err := os.Open("seating.json")
	if err != nil {
		fmt.Println("error opening seating file :", err)
		return nil, err
	}
	defer seatingFile.Close()

	byteValue, err := ioutil.ReadAll(seatingFile)
	if err != nil {
		fmt.Println("error reading layout file :", err)
		return nil, err
	}

	var seating []Seating
	err = json.Unmarshal(byteValue, &seating)
	if err != nil {
		fmt.Println("err :", err)
		return nil, err
	}

	return seating, nil
}

func (s seatingConfigManagerImpl) insert(DB *sql.DB, seating []Seating) error {
	for _, v := range seating {
		statement := fmt.Sprintf("INSERT INTO seating(seating_capacity, min_occupency) VALUES('%d', '%d')", v.SeatingCapacity, v.MinOccupency)
		_, err := DB.Exec(statement)
		if err != nil {
			fmt.Println("error inserting in seating :", err)
			return err
		}
	}
	return nil
}
