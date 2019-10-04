package main

import (
	"log"
	"time"

	"github.com/FilipAnteKovacic/grpc-gorm-template/protoexpl"
)

type dataList struct {
	Data []dataItem `json:"data"`
}

type dataItem struct {
	ID        uint `json:",string" gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	Name      string     `gorm:"size:255"` // set field size to 255
}

func reqDataToStuct(reqData *protoexpl.Data) dataItem {

	id := reqData.GetId()

	d := dataItem{
		Name: reqData.GetName(),
	}

	if id != 0 {
		d.ID = uint(id)
	}

	return d

}

func structDataToRes(data dataItem) *protoexpl.Data {

	id := data.ID

	d := &protoexpl.Data{
		Name: data.Name,
	}

	if id != 0 {
		d.Id = uint64(id)
	}

	return d

}

func main() {

	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dbSession()

	Serve()

	//GRPCclient()
	//RESTclient()

}
