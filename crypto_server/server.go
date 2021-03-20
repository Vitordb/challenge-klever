package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	postpb "github.com/Vitordb/crypto/cryptopb"
	"github.com/Vitordb/crypto/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostServiceServer struct {
	postpb.UnimplementedPostServiceServer
}

var db *mongo.Client
var postDB *mongo.Collection
var mongoCtx context.Context

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Port :50051...")

	listener, err := net.Listen("tcp", "localhost:50051")

	if err != nil {
		log.Fatalf("Unable to listen on port :50051: %v", err)
	}

	opts := []grpc.ServerOption{}

	s := grpc.NewServer(opts...)

	srv := &PostServiceServer{}

	postpb.RegisterPostServiceServer(s, srv)

	fmt.Println("Connecting...")
	mongoCtx = context.Background()
	db, err = mongo.Connect(mongoCtx, options.Client().ApplyURI("mongodb+srv://dbvitor:92072101@cryptocluster.fdk7o.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
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

	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	fmt.Println("Server succesfully started on port :50051")

	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt)

	<-c

	fmt.Println("\nStopping the server...")
	s.Stop()
	listener.Close()
	fmt.Println("Closing")
	db.Disconnect(mongoCtx)
	fmt.Println("Done.")
}

func (s *PostServiceServer) CreatePost(ctx context.Context, request *postpb.CreatePostRequest) (*postpb.CreatePostResponse, error) {

	post := request.GetPost()

	data := model.Post{
		Id:    primitive.NewObjectID(),
		Title: post.GetTitle(),
		Votes: post.GetVotes(),
	}

	result, err := postDB.InsertOne(mongoCtx, data)

	if err != nil {

		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	if len(post.GetTitle()) == 0 {
		fmt.Println("String is empty!", err)
	}

	oid := result.InsertedID.(primitive.ObjectID)
	post.Id = oid.Hex()

	return &postpb.CreatePostResponse{Post: post}, nil
}

func (s *PostServiceServer) GetPost(ctx context.Context, request *postpb.GetPostRequest) (*postpb.GetPostResponse, error) {
	oid, err := primitive.ObjectIDFromHex(request.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}
	result := postDB.FindOne(ctx, bson.M{"_id": oid})

	data := model.Post{}

	if err := result.Decode(&data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find post with Object Id %s: %v", request.GetId(), err))
	}

	response := &postpb.GetPostResponse{
		Post: &postpb.Post{
			Id:    oid.Hex(),
			Title: data.Title,
			Votes: data.Votes,
		},
	}
	return response, nil
}

func (s *PostServiceServer) ListPosts(ctx context.Context, request *postpb.ListPostsRequest) (*postpb.ListPostsResponse, error) {

	filter := bson.M{}
	data := []*postpb.Post{}

	cursor, _ := postDB.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var p model.Post
		err := cursor.Decode(&p)
		if err != nil {
			continue
		}
		data = append(data, &postpb.Post{
			Id:    p.Id.Hex(),
			Title: p.Title,
			Votes: p.Votes,
		})
	}
	return &postpb.ListPostsResponse{
		Post: data,
	}, nil
}

func (s *PostServiceServer) DeletePost(ctx context.Context, request *postpb.DeletePostRequest) (*postpb.DeletePostResponse, error) {
	oid, err := primitive.ObjectIDFromHex(request.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	filter := bson.M{"_id": oid}

	result := postDB.FindOneAndDelete(ctx, filter)

	decoded := model.Post{}
	err = result.Decode(&decoded)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find/delete post with id %s: %v", request.GetId(), err))
	}
	return &postpb.DeletePostResponse{
		Success: true,
	}, nil
}

func (s *PostServiceServer) UpdatePost(ctx context.Context, request *postpb.UpdatePostRequest) (*postpb.UpdatePostResponse, error) {

	post := request.GetPost()

	oid, err := primitive.ObjectIDFromHex(post.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not convert the supplied post id to a MongoDB ObjectId: %v", err),
		)
	}

	update := bson.M{
		"title": post.GetTitle(),
	}

	filter := bson.M{"_id": oid}

	result := postDB.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(1))

	decoded := model.Post{}
	err = result.Decode(&decoded)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Could not find post with supplied ID: %v", err),
		)
	}
	return &postpb.UpdatePostResponse{
		Post: &postpb.Post{
			Id:    decoded.Id.Hex(),
			Title: decoded.Title,
			Votes: decoded.Votes,
		},
	}, nil
}

func (s *PostServiceServer) UpVote(ctx context.Context, request *postpb.UpVoteRequest) (*postpb.UpVoteResponse, error) {
	post := request.GetPost()
	oid, err := primitive.ObjectIDFromHex(post.GetId())

	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not convert the supplied post id to a MongoDB ObjectId: %v", err),
		)
	}

	filter := bson.M{"_id": oid}

	result := postDB.FindOneAndUpdate(ctx, filter, bson.M{"$inc": bson.M{"votes": 1}}, options.FindOneAndUpdate().SetReturnDocument(1))

	decoded := model.Post{}
	err = result.Decode(&decoded)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Could not find post with supplied ID: %v", err),
		)
	}
	return &postpb.UpVoteResponse{
		Post: &postpb.Post{
			Id:    decoded.Id.Hex(),
			Title: decoded.Title,
			Votes: decoded.Votes,
		},
	}, nil
}

func (s *PostServiceServer) DownVote(ctx context.Context, request *postpb.DownVoteRequest) (*postpb.DownVoteResponse, error) {

	post := request.GetPost()

	oid, err := primitive.ObjectIDFromHex(post.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not convert the supplied post id to a MongoDB ObjectId: %v", err),
		)
	}

	filter := bson.M{"_id": oid}

	result := postDB.FindOneAndUpdate(ctx, filter, bson.M{"$inc": bson.M{"votes": -1}}, options.FindOneAndUpdate().SetReturnDocument(1))

	decoded := model.Post{}
	err = result.Decode(&decoded)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Could not find post with supplied ID: %v", err),
		)
	}
	return &postpb.DownVoteResponse{
		Post: &postpb.Post{
			Id:    decoded.Id.Hex(),
			Title: decoded.Title,
			Votes: decoded.Votes,
		},
	}, nil
}
