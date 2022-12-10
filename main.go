package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	. "room_status/temp/controller"
	. "room_status/temp/db"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var 
(
	ctx context.Context
	mongoClient *mongo.Client
	tempController TempController
	err error
)

func sub(client mqtt.Client) {
    topic := "temp_topic"
    token := client.Subscribe(topic, 1, nil)
    token.Wait()
  fmt.Printf("Subscribed to topic: %s", topic)
}


var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    // fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	if msg.Topic()=="temp_topic" {
		tempController.SubscribedTempStatus(string(msg.Payload()))
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    fmt.Printf("Connect lost: %v", err)
}


func connectMQTT()(mqtt.Client){
//
var broker = "192.168.0.60"
var port = 1883
opts := mqtt.NewClientOptions()
opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
opts.SetClientID("go_mqtt_client")
opts.SetUsername("swayamshree")
opts.SetPassword("Asphalt@7")
opts.SetDefaultPublishHandler(messagePubHandler)
opts.OnConnect = connectHandler
opts.OnConnectionLost = connectLostHandler
client := mqtt.NewClient(opts)
if token := client.Connect(); token.Wait() && token.Error() != nil {
	panic(token.Error())
}
sub(client)
return client
// publish(client)
// client.Disconnect(250)
//
}


func init() {
	ctx=context.TODO()
	mongoDbConnection:=options.Client().ApplyURI("mongodb://0.0.0.0:27017")
	mongoClient,err=mongo.Connect(ctx,mongoDbConnection)
	if err!=nil {
		log.Fatal("error while connecting with mongo",err)
	}
	err=mongoClient.Ping(ctx,readpref.Primary())
	if err!=nil {
		log.Fatal("error while tring to ping mongo",err)
	}
	
}

func main() {
	client:=connectMQTT()
	tempMongoService:=TempMongoServiceInit(ctx,mongoClient)
	tempController=	TempNewController(client,tempMongoService)

	router := gin.Default()
	versionOne:=router.Group("/swayamroom/v1/")
	
	tempController.RegisterTempManagerRoutes(versionOne)
    router.GET("/ws", func(c *gin.Context) {
        wshandler(c.Writer, c.Request)
    })

    router.Run("0.0.0.0:8000")
}

var wsupgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func wshandler(w http.ResponseWriter, r *http.Request) {
    conn, err := wsupgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: %+v", err)
        return
    }

    for {
        t, msg, err := conn.ReadMessage()
		fmt.Println("**")
        if err != nil {
            break
        }
        conn.WriteMessage(t, msg)
    }
}