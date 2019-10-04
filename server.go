package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/FilipAnteKovacic/grpc-gorm-template/protoexpl"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sacOO7/gowebsocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	status "google.golang.org/grpc/status"
)

// WSMessage receive message
type WSMessage struct {
	Source string   `bson:"source,omitempty" json:"source"`
	Action string   `bson:"action,omitempty" json:"action"`
	Data   dataItem `bson:"data,omitempty" json:"data"`
}

// Init socket conn for grpc
var wsConn gowebsocket.Socket

func wsConection() {

	wsConn = gowebsocket.New("ws://localhost:" + os.Getenv("UI") + "/ws")

	wsConn.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Println("Received connect error - ", err)
	}
	wsConn.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}

	wsConn.Connect()

}

func sendWSMessage(wsMsg WSMessage) {

	if wsConn.IsConnected {

		msg, err := json.Marshal(wsMsg)
		if err != nil {
			fmt.Println("Error while parsing WS message", err)
			return
		}

		wsConn.SendText(string(msg))

	}

	return
}

type server struct {
}

func (*server) Create(ctx context.Context, req *protoexpl.CreateRequest) (*protoexpl.CreateResponse, error) {

	// Request data to item data
	data := reqDataToStuct(req.GetData())

	// Create data in DB
	err := create(data)
	if err != nil {

		// Return response error
		return &protoexpl.CreateResponse{
			Status: "error",
		}, err

	}

	// Push msg to websocket clients
	sendWSMessage(WSMessage{
		Source: "server",
		Action: "create",
	})

	// Return response success
	return &protoexpl.CreateResponse{
		Status: "success",
	}, nil

}

func (*server) Read(ctx context.Context, req *protoexpl.ReadRequest) (*protoexpl.ReadResponse, error) {

	// Read by ID
	data, err := readByID(uint64(req.GetId()))
	if err != nil {

		// Return response error
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find  with specified ID: %v", err),
		)
	}

	// Return response data
	return &protoexpl.ReadResponse{
		Data: structDataToRes(data),
	}, nil
}

func (*server) Update(ctx context.Context, req *protoexpl.UpdateRequest) (*protoexpl.UpdateResponse, error) {

	// Get request data
	reqData := req.GetData()

	// Read request ID
	ID := reqData.GetId()

	// Request data to item data
	data := reqDataToStuct(reqData)

	// Update in DB
	err := updateByID(ID, data)
	if err != nil {

		// Return response error
		return &protoexpl.UpdateResponse{
			Status: "error",
		}, err

	}

	// Push msg to websocket clients
	sendWSMessage(WSMessage{
		Source: "server",
		Action: "update",
		Data:   data,
	})

	// Return response success
	return &protoexpl.UpdateResponse{
		Status: "success",
	}, nil

}

func (*server) Delete(ctx context.Context, req *protoexpl.DeleteRequest) (*protoexpl.DeleteResponse, error) {

	// Read request ID
	ID := req.GetId()

	// Delete
	err := deleteByID(uint64(ID))
	if err != nil {

		// Return response error
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete object in MongoDB: %v", err),
		)

	}

	// Push msg to websocket clients
	sendWSMessage(WSMessage{
		Source: "server",
		Action: "update",
	})
	// Return response success
	return &protoexpl.DeleteResponse{
		Status: "success",
	}, nil
}

func (*server) List(ctx context.Context, req *protoexpl.ListRequest) (*protoexpl.ListResponse, error) {

	listRes, err := list()
	if err != nil {

		// Return response error
		return nil, status.Errorf(
			codes.DataLoss,
			err.Error(),
		)
	}

	// Return response success
	return &protoexpl.ListResponse{
		Data: convertListToProto(listRes),
	}, nil
}

func startGRPCServer(address, certFile, keyFile string) error {
	// create a listener on TCP port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("could not load TLS keys: %s", err)
	}

	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{grpc.Creds(creds)}

	// create a gRPC server object
	grpcServer := grpc.NewServer(opts...)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	protoexpl.RegisterDataServiceServer(grpcServer, &server{})

	// start the server
	log.Printf("starting HTTP/2 gRPC server on %s", address)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %s", err)
	}

	return nil
}

func startRESTServer(address, grpcAddress, certFile string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		return fmt.Errorf("could not load TLS certificate: %s", err)
	}

	// Setup the client gRPC options
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	// Register server name
	err = creds.OverrideServerName("localhost")
	if err != nil {
		return fmt.Errorf("could not load TLS certificate: %s", err)
	}

	// Register ping
	err = protoexpl.RegisterDataServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return fmt.Errorf("could not register service Ping: %s", err)
	}

	log.Printf("starting HTTP/1.1 REST server on %s", address)
	http.ListenAndServe(address, mux)

	return nil
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// Hub WSclients hub
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				//client.conn.Close()
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {

				if err := client.conn.WriteMessage(1, message); err != nil {
					return
				}
			}
		}
	}
}

type wsPage struct {
	WSURL string
}

func startUIServer(uiAddress string) error {

	// Init ws path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		uiP := wsPage{
			WSURL: "ws://localhost:" + os.Getenv("UI") + "/ws",
		}

		crudTemplate, err := template.ParseFiles(
			"ui/home.html",
		)
		if err != nil {
			fmt.Println("Error occurred while parsing template", err)
			return
		}

		err = crudTemplate.Execute(w, &uiP)
		if err != nil {
			fmt.Println("Error occurred while executing the template  or writing its output", err)
			return
		}

	})

	hub := newHub()
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{}

		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		// Init client
		client := &Client{hub: hub, conn: c, send: make(chan []byte, 256)}

		// Register client in hub channel
		hub.register <- client

		go func() {

			defer c.Close()

			for {

				// Read incoming msg
				_, message, err := c.ReadMessage()
				if err != nil {
					hub.unregister <- client
					break
				}

				var wsMsg WSMessage

				json.Unmarshal(message, &wsMsg)

				switch wsMsg.Source {
				case "server":

					// List
					listRes, err := list()
					if err != nil {
						hub.broadcast <- []byte(err.Error())
					}

					// List to json
					listMarshal, err := json.Marshal(convertListToProto(listRes))
					if err != nil {
						fmt.Println("cannot marshal data", err)
					}

					// Push to clients
					hub.broadcast <- listMarshal
					break

				case "client":

					switch wsMsg.Action {
					case "list":

						// List
						listRes, err := list()
						if err != nil {
							hub.broadcast <- []byte(err.Error())
						}

						// List to json
						listMarshal, err := json.Marshal(convertListToProto(listRes))
						if err != nil {
							fmt.Println("cannot marshal data", err)
						}

						// Push to clients
						hub.broadcast <- listMarshal

						break

					case "create":

						// Create data in DB
						err := create(wsMsg.Data)
						if err != nil {

							fmt.Println("error by create", err)

						}

						// Push msg to websocket clients
						sendWSMessage(WSMessage{
							Source: "server",
							Action: "create",
						})

						break
					case "update":

						ID := wsMsg.Data.ID

						// Update in DB
						err := updateByID(uint64(ID), wsMsg.Data)
						if err != nil {

							fmt.Println("error by delete", err)
						}

						// Push msg to websocket clients
						sendWSMessage(WSMessage{
							Source: "server",
							Action: "update",
							Data:   wsMsg.Data,
						})

						break

					case "delete":

						// Delete data in DB
						err := deleteByID(uint64(wsMsg.Data.ID))
						if err != nil {

							fmt.Println("error by delete", err)

						}

						// Push msg to websocket clients
						sendWSMessage(WSMessage{
							Source: "server",
							Action: "delete",
						})

						break
					}
					break
				}

			}

		}()

	})

	fs := http.FileServer(http.Dir("ui/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("starting HTTP/1.1 UI server on %s", uiAddress)
	err := http.ListenAndServe(uiAddress, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return nil
}

// Serve start grpc server
func Serve() {

	restAddress := ":" + os.Getenv("REST")
	grpcAddress := ":" + os.Getenv("GRPC")
	uiAddress := ":" + os.Getenv("UI")
	certFile := "cert/server.crt"
	keyFile := "cert/server.pem"

	if os.Getenv("GRPC") != "" {

		// fire the gRPC server in a goroutine
		go func() {
			err := startGRPCServer(grpcAddress, certFile, keyFile)
			if err != nil {
				log.Fatalf("failed to start gRPC server: %s", err)
			}
		}()

	}

	if os.Getenv("REST") != "" && os.Getenv("GRPC") != "" {

		// fire the REST server in a goroutine
		go func() {
			err := startRESTServer(restAddress, grpcAddress, certFile)
			if err != nil {
				log.Fatalf("failed to start rest server: %s", err)
			}
		}()

	}
	// connect wsConnections

	if os.Getenv("UI") != "" {

		go func() {
			wsConection()
		}()

		// fire the UI server in a goroutine
		go func() {
			err := startUIServer(uiAddress)
			if err != nil {
				log.Fatalf("failed to start ws server: %s", err)
			}
		}()

	}

	// infinite loop
	log.Printf("Entering infinite loop")
	select {}
}
