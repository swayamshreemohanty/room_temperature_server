package db

import (
	"context"
	"errors"
	. "room_status/temp/helper"
	. "room_status/temp/model"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type TempDBContext struct {
	ctx         context.Context
	mongoclient *mongo.Client
}

type TempMongoService interface {
	UpdateTempToDB(newTemp *string) (*string, error)
	FetchTempDetails() (*TempModel, error)
}

func TempMongoServiceInit(ctx context.Context,mongoclient *mongo.Client)  TempMongoService{
	return &TempDBContext{
		ctx: ctx,
		mongoclient: mongoclient,
	}
}

func (tempMongoContext *TempDBContext) UpdateTempToDB(newTemp *string) (*string, error) {
	dbref := tempMongoContext.mongoclient.Database(string(DatabasePath.DATABASE)).Collection(string(DatabasePath.STATUS))

	myOption := options.FindOne()
	myOption.SetSort(bson.M{"$natural": -1})

	var tempModel TempModel
	
	dbref.FindOne(tempMongoContext.ctx, bson.M{"_id":"temperature"}, myOption).Decode(&tempModel)
	
	tempModel.Temperature = *newTemp;

	//Add the update time
	tempModel.LastUpdateAt= time.Now().Unix()

	//Set filter
	filter:=bson.M{"_id":tempModel.Id}

	//set update model
	update:=bson.M{"$set":tempModel}

	result, err := dbref. UpdateOne(tempMongoContext.ctx,filter,update)
	if err!=nil {
		return nil,err
	}else if result.MatchedCount !=1 {
		//Insert data if update model is not available
		tempModel.Id = "temperature"
		_, err := dbref. InsertOne(tempMongoContext.ctx,tempModel)
		if err!=nil{
		return nil,errors.New("Unable to add temp status")
		}
		message:="Temp added successfully"
		return &message,nil
		//
	 }	else{
		message:="Temp updated successfully"
		return &message,nil
	}
}
	func (tempMongoContext *TempDBContext)FetchTempDetails() (*TempModel,error)  {
		dbref:= tempMongoContext.mongoclient.Database(string(DatabasePath.DATABASE)).Collection(string(DatabasePath.STATUS))
	
		var tempModel TempModel
	
		filter:=bson.M{"_id":"temperature"}
	
		err:=dbref.FindOne(tempMongoContext.ctx,filter).Decode(&tempModel)
		if err!=nil {
			return nil,errors.New("No temperature details found")
		}else{
			return &tempModel,nil
		}
	}
	
