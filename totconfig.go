package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Tot is use to represent average total occupancy
type Tot struct {
	MinPartySize int    `json:"min_party_size"`
	MaxPartySize int    `json:"max_party_size"`
	AvgTOT       string `json:"avg_tot"`
}

type totConfigManager interface {
	get(DB *sql.DB) ([]Tot, error)
	insert(DB *sql.DB, tot []Tot) error
}

type totConfigManagerImpl struct {
}

func (t totConfigManagerImpl) get(DB *sql.DB) ([]Tot, error) {

	totFile, err := os.Open("tot.json")
	if err != nil {
		fmt.Println("error opening tot file :", err)
		return nil, err
	}
	defer totFile.Close()

	byteValue, err := ioutil.ReadAll(totFile)
	if err != nil {
		fmt.Println("error reading tot file :", err)
		return nil, err
	}

	var tot []Tot
	err = json.Unmarshal(byteValue, &tot)
	if err != nil {
		fmt.Println("err :", err)
		return nil, err
	}

	return tot, nil
}

func (t totConfigManagerImpl) insert(DB *sql.DB, tot []Tot) error {
	for _, v := range tot {
		statement := fmt.Sprintf("INSERT INTO tot(min_party_size, max_party_size, avg_tot) VALUES('%d', '%d', '%s')", v.MinPartySize, v.MaxPartySize, v.AvgTOT)
		_, err := DB.Exec(statement)
		if err != nil {
			fmt.Println("error inserting in tot :", err)
			return err
		}
	}
	return nil
}
