package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostServiceServer struct {
}

type Post struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	Title     string             `json:"title" bson:"title"`
	Votes     int64              `json:"votes" bson:"votes"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type Posts []Post
