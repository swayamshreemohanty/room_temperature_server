package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	. "room_status/temp/db"
	. "room_status/temp/model"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"
)
var clients = make(map[*websocket.Conn]bool)
type TempController struct {
	tempMongoService TempMongoService
	mqttClient mqtt.Client
}

func TempNewController(mqttClient mqtt.Client, tempMongoService TempMongoService) TempController {
	return TempController{mqttClient: mqttClient,tempMongoService:tempMongoService}
}


func (tempController *TempController)SubscribedTempStatus(tempData string )  {
	
	message,err:=tempController.tempMongoService.UpdateTempToDB(&tempData)
	if err!=nil {
		text:= fmt.Sprint(err.Error())
		tempController.mqttClient.Publish("rpi_sender",0,false,text);
		return
	}

	tempController.mqttClient.Publish("rpi_sender",0,false,*message);
	err= tempController.broadCastTempDataInWebSocket()
	if err!=nil {
		log.Println(err)
		return
	}
}

func (tempController *TempController)ReadTempStatus(c *gin.Context)  {
	tempModel,err:=tempController.tempMongoService.FetchTempDetails()
		if err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{
				"status": false,
				"message":err.Error(),
				}) 
			return
		}else{
			c.JSON(http.StatusOK,gin.H{
				"status": true,
				"data":tempModel,
				}) 
			return
		}
}



func convertTempModelToByte(tempModel *TempModel)([]byte,error){
	body:=bson.M{"data":tempModel}
	finalBody,err:=json.Marshal(body)
	if err!=nil {
		return nil,err
	}else{
		return finalBody,nil
	}
}

//WebSockets
func (tempController *TempController)sendTempDetailsInWebSocket(conn *websocket.Conn)error  {
	var tempModel *TempModel
	tempModel,err:=tempController.tempMongoService.FetchTempDetails()
	if err!=nil{
		return err
	}
	body,err:=convertTempModelToByte(tempModel)
	if err==nil{
		if err := conn.WriteMessage(websocket.TextMessage, []byte(body)); err != nil {
			log.Println(err)
			return err
		}
	}
	return err
}

func(tempController *TempController) broadCastTempDataInWebSocket()(error) {
		for client :=range clients{
			err:=tempController.sendTempDetailsInWebSocket(client)
				if err!=nil {
					log.Println(err)
					//Close the client if unable to send data
					client.Close()
					delete(clients, client)
				return err
			}	
		}
		return nil
}

func (tempController *TempController)WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	http.ListenAndServe(":80",nil)
    upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	//upgrade this connection to a WebSocket
	//connection
	ws,err:=upgrader.Upgrade(w,r,nil)
	if err != nil {
        log.Println(err)
    }

	log.Println("Client Connected")

	//Add the current websocket
	clients[ws]=true


    // err = ws.WriteMessage(1, []byte("Hi Client!"))
    // if err != nil {
    //     log.Println(err)
    // }

	err= tempController.broadCastTempDataInWebSocket()
	if err!=nil {
		log.Println(err)
		return
	}
}

func (tempController *TempController) RegisterTempManagerRoutes(ginRouter *gin.RouterGroup) {
	switchRoute := ginRouter.Group("temp")
	switchRoute.GET("details", tempController.ReadTempStatus)
	switchRoute.GET("/ws", func(c *gin.Context) {
		tempController.WebSocketHandler(c.Writer, c.Request)
    })
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}