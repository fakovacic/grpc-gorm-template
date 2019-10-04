package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/FilipAnteKovacic/grpc-gorm-template/protoexpl"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GRPCclient client consuming grpc server
// following Create, List, Update, Read, Delete
func GRPCclient() {

	// Create the client TLS credentials
	creds, err := credentials.NewClientTLSFromFile("cert/ca.crt", "")
	if err != nil {
		log.Fatalf("could not load tls cert: %s", err)
	}

	cc, err := grpc.Dial("localhost:"+os.Getenv("GRPC"), grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	c := protoexpl.NewDataServiceClient(cc)

	// Create
	create := dataItem{
		Name: "London",
	}

	createRes, err := c.Create(
		context.Background(),
		&protoexpl.CreateRequest{
			Data: structDataToRes(create),
		},
	)
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	fmt.Println("Created:", createRes.GetStatus())

	// List
	listResponse, err := c.List(context.Background(), &protoexpl.ListRequest{})
	if err != nil {
		log.Fatalf("error while calling ListBlog RPC: %v", err)
	}

	list := listResponse.GetData()

	// Case data exist
	if len(list) != 0 {

		for _, l := range list {

			d := reqDataToStuct(l)

			// Update
			d.Name = "Mancaster"

			updateRes, updateErr := c.Update(context.Background(), &protoexpl.UpdateRequest{Data: structDataToRes(d)})
			if updateErr != nil {
				fmt.Printf("Error happened while updating: %v \n", updateErr)
			}
			fmt.Println("Updated: ", updateRes.GetStatus())

			// Read
			readReq := &protoexpl.ReadRequest{Id: uint64(d.ID)}
			readRes, readErr := c.Read(context.Background(), readReq)
			if readErr != nil {
				fmt.Println("Error happened while reading:", readErr)
			}

			fmt.Println("Read:", readRes.GetData())

			// Delete
			deleteRes, deleteErr := c.Delete(context.Background(), &protoexpl.DeleteRequest{Id: uint64(d.ID)})

			if deleteErr != nil {
				fmt.Println("Error happened while deleting:", deleteErr)
			}
			fmt.Println("Deleted:", deleteRes.GetStatus())

		}

	}

}

// RESTclient client consuming rest server
// following Create, List, Update, Read, Delete
func RESTclient() {

	// Create
	item := dataItem{
		Name: "Rest create",
	}

	itemBytes, err := json.Marshal(item)

	if err != nil {
		fmt.Println("cannot marshal data", err)
	}

	body := httpRequest("create", itemBytes)

	fmt.Println("Create:", string(body))

	// List
	body = httpRequest("list", itemBytes)

	var list dataList

	fmt.Println(string(body))

	err = json.Unmarshal(body, &list)
	if err != nil {
		fmt.Println("cannot unmarshall list:", err)
	}

	if len(list.Data) != 0 {

		for _, l := range list.Data {

			// Update

			l.Name = "Rest update"

			itemBytes, err = json.Marshal(l)

			if err != nil {
				fmt.Println("cannot marshal data", err)
			}
			body = httpRequest("update", itemBytes)

			fmt.Println("Update:", string(body))

			// Read
			body = httpRequest("read", []byte(strconv.Itoa(int(l.ID))))
			fmt.Println("Read:", string(body))

			// Delete
			body = httpRequest("delete", []byte(strconv.Itoa(int(l.ID))))
			fmt.Println("Delete:", string(body))

		}

	}

}

//GetJSON from target and encode body
func httpRequest(reqType string, body []byte) []byte {

	client := &http.Client{}

	switch reqType {
	case "read":

		fmt.Println(string(body))

		req, err := http.NewRequest(
			"GET",
			"http://localhost:"+os.Getenv("REST")+"/v1/read/"+string(body),
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ = ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			fmt.Println("Read error:", resp.Status)
			fmt.Println("Read body:", string(body))
		}

		return body

	case "create":

		req, err := http.NewRequest(
			"POST",
			"http://localhost:"+os.Getenv("REST")+"/v1/create",
			bytes.NewBuffer(body),
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			fmt.Println("Create error:", resp.Status)
			fmt.Println("Create body:", string(body))
		}

		return body

	case "update":

		req, err := http.NewRequest(
			"PUT",
			"http://localhost:"+os.Getenv("REST")+"/v1/update",
			bytes.NewBuffer(body),
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			fmt.Println("Update error:", resp.Status)
			fmt.Println("Update body:", string(body))
		}

		return body

	case "delete":

		req, err := http.NewRequest(
			"DELETE",
			"http://localhost:"+os.Getenv("REST")+"/v1/delete/"+string(body),
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ = ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			fmt.Println("Delete error:", resp.Status)
			fmt.Println("Delete body:", string(body))
		}

		return body

	case "list":

		req, err := http.NewRequest(
			"GET",
			"http://localhost:"+os.Getenv("REST")+"/v1/list",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ = ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			fmt.Println("List error:", resp.Status)
			fmt.Println("List body:", string(body))
		}

		return body

	}

	return nil
}
