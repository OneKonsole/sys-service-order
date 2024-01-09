package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	oko "github.com/OneKonsole/order-model"
	okmq "github.com/OneKonsole/sys-queueing"

	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
)

type App struct {
	Router       *mux.Router
	MQChannel    *amqp.Channel
	MQConnection *amqp.Connection
	AppConf      *AppConf
}

type AppConf struct {
	ServedPort string `json:"served_port"`
	MQUser     string `json:"mq_user"`
	MQPassword string `json:"mq_password"`
	MQUrl      string `json:"mq_url"`
	MQVhost    string `json:"mq_vhost"`
}

// ===========================================================================================================
// Runs the HTTP server
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	addr (string): Full URL to use for the server
//
// Examples:
//
//	a.Run("localhost:8010")
//
// ===========================================================================================================
func (a *App) Run() {
	log.Fatal(http.ListenAndServe(":"+a.AppConf.ServedPort, a.Router))
}

// ===========================================================================================================
// Initialize database and http server for the order service
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// Parameters:
//
//	user (string) : Database user
//	password (string) : Database password
//	dbName (string) : Database name
//
// Examples:
//
//	a.Initialize("testuser","testpassword","mydb")
//
// ===========================================================================================================
func (a *App) Initialize() {

	rabbitConnectionString := fmt.Sprintf("amqp://%s:%s@%s:5672/%s",
		a.AppConf.MQUser,
		a.AppConf.MQPassword,
		a.AppConf.MQUrl,
		a.AppConf.MQVhost)
	a.MQConnection = okmq.NewMQConnection(rabbitConnectionString)
	a.MQChannel = okmq.NewMQChannel(a.MQConnection)
	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

func (appConf *AppConf) Initialize() {
	confFile, _ := os.ReadFile("app-configuration.json")

	err := json.Unmarshal([]byte(confFile), &appConf)
	if err != nil {
		log.Fatalf("Unable to parse application config file due to : %s", err)
	}
}

func (a *App) produceOrder(w http.ResponseWriter, r *http.Request) {
	var o oko.Order

	fmt.Print("Received request \n\n")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading request body")
		return
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
	if err := decoder.Decode(&o); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	fmt.Printf("Body: %v\n", o)

	produce(
		a.MQChannel,
		"provisioning",
		"order-service-exchange",
		"application/json",
		5,
		bodyBytes,
	)

	respondWithJSON(w, http.StatusCreated, o)

}

// ===========================================================================================================
// Initialize every HTTP route of our application
//
// Used on:
//
//	a (*App) : App struct containing the service necessary items
//
// ===========================================================================================================
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/produce/order", a.produceOrder).Methods("POST") // Create an order and call sys order service
}
