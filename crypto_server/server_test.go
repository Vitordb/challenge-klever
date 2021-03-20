package main

import (
	"context"
	"fmt"
	"log"
	"testing"

	postpb "github.com/Vitordb/crypto/cryptopb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitDb() {

	mongoCtx = context.Background()
	db, err := mongo.Connect(mongoCtx, options.Client().ApplyURI("mongodb+srv://dbvitor:92072101@cryptocluster.fdk7o.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping(mongoCtx, nil)
	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %v\n", err)
	} else {
		fmt.Println("Connected to Mongodb")
	}

	postDB = db.Database("myFirstDatabase").Collection("posts")

}

func TestCreatePost(t *testing.T) {

	InitDb()

	srv := PostServiceServer{}

	req := &postpb.CreatePostRequest{
		Post: &postpb.Post{Title: "Hitcoin", Votes: 0},
	}

	res, err := srv.CreatePost(context.Background(), req)
	require.Nil(t, err)
	assert.Equal(t, "Hitcoin", res.GetPost().GetTitle())

}
