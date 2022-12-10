package controller

import (
	"fmt"
	"net/http"
	. "room_status/temp/db"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

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

func (tempController *TempController) RegisterTempManagerRoutes(ginRouter *gin.RouterGroup) {
	switchRoute := ginRouter.Group("temp")
	switchRoute.GET("details", tempController.ReadTempStatus)
}