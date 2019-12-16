package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// LayOut is use to represent one layout json entry
type LayOut struct {
	TableID    int `json:"table_id"`
	MaxSeating int `json:"max_seating"`
	//Availibility string `json:"availability"` // Default value is 'Y'
}

type layOutConfigManager interface {
	get(DB *sql.DB) ([]LayOut, error)
	insert(DB *sql.DB, layOut []LayOut) error
}

type layOutConfigManagerImpl struct {
}

func (l layOutConfigManagerImpl) get(DB *sql.DB) ([]LayOut, error) {

	layOutFile, err := os.Open("layout.json")
	if err != nil {
		fmt.Println("error opening layout file :", err)
		return nil, err
	}
	defer layOutFile.Close()

	byteValue, err := ioutil.ReadAll(layOutFile)
	if err != nil {
		fmt.Println("error reading layout file :", err)
		return nil, err
	}

	var layOut []LayOut
	err = json.Unmarshal(byteValue, &layOut)
	if err != nil {
		fmt.Println("err :", err)
		return nil, err
	}

	return layOut, nil
}

func (l layOutConfigManagerImpl) insert(DB *sql.DB, layOut []LayOut) error {
	for _, v := range layOut {
		statement := fmt.Sprintf("INSERT INTO layout(table_id, max_seating) VALUES('%d', '%d')", v.TableID, v.MaxSeating)
		_, err := DB.Exec(statement)
		if err != nil {
			fmt.Println("error inserting in layout :", err)
			return err
		}
	}
	return nil
}
