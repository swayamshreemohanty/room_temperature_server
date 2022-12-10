package model

type TempModel struct {
	Id          string `json:"id" bson:"_id" form:"id"`
	Temperature string   `json:"temperature" bson:"temperature" form:"temperature" binding:"required"`
	LastUpdateAt int64	`json:"last_update_at" bson:"last_update_at" form:"last_update_at" binding:"required"`
}