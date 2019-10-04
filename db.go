package main

import (
	"fmt"
	"os"

	"github.com/FilipAnteKovacic/grpc-gorm-template/protoexpl"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB connection to SQL like db
var DB *gorm.DB

func dbSession() {

	var err error

	DB, err = gorm.Open("mysql", os.Getenv("DBCONN"))
	if err != nil {
		fmt.Println("Error while parsing WS message", err)
	}

	DB.Table("data_items").CreateTable(&dataItem{})

	//defer DB.Close()

}

func create(data dataItem) error {

	// Insert data
	DB.NewRecord(data)

	DB.Create(&data)

	return nil

}

func readByID(ID uint64) (dataItem, error) {

	// Find by ID
	data := dataItem{}

	DB.First(&data, ID)

	return data, nil
}

func updateByID(ID uint64, data dataItem) error {

	//db.First(&data)

	DB.Save(&data)

	return nil

}

func deleteByID(ID uint64) error {

	data, _ := readByID(ID)

	DB.Delete(&data)

	return nil

}

func list() ([]dataItem, error) {

	var items []dataItem

	DB.Find(&items)

	return items, nil

}

func convertListToProto(uD []dataItem) []*protoexpl.Data {

	var listRes []*protoexpl.Data

	for _, d := range uD {

		listRes = append(listRes, structDataToRes(d))

	}

	return listRes

}
