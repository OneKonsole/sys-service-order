package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	appConf.ServedPort = os.Getenv("served_port")
	appConf.MQUser = os.Getenv("mq_user")
	appConf.MQPassword = os.Getenv("mq_password")
	appConf.MQUrl = os.Getenv("mq_URL")
	appConf.MQVhost = os.Getenv("mq_vhost")

	fmt.Print(appConf)
}

func (a *App) validatePodHealth(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "")
}
func (a *App) produceOrder(w http.ResponseWriter, r *http.Request) {
	var o oko.Order

	bodyBytes, err := io.ReadAll(r.Body)
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
	a.Router.HandleFunc("/readiness", a.validatePodHealth).Methods("GET") // Create an order and call sys order service
}
