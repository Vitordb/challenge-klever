package main

import (
	"log"
	"net/http"

	postpb "github.com/Vitordb/crypto/cryptopb"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := postpb.NewPostServiceClient(conn)

	r := gin.Default()

	r.GET("/crypto/:id", func(c *gin.Context) {
		id := c.Param("id")

		req := &postpb.GetPostRequest{Id: id}
		res, err := client.GetPost(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, res.Post)
	})

	r.GET("/crypto/", func(c *gin.Context) {

		req := &postpb.ListPostsRequest{}
		res, err := client.ListPosts(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	//create
	r.POST("/crypto", func(c *gin.Context) {
		post := postpb.Post{}

		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Missing values"})
			return
		}

		req := &postpb.CreatePostRequest{
			Post: &post,
		}
		res, err := client.CreatePost(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, res.Post)
	})

	r.DELETE("/crypto/:id", func(c *gin.Context) {
		id := c.Param("id")

		req := &postpb.DeletePostRequest{Id: id}
		res, err := client.DeletePost(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"Id:":           id,
			"Post deleted:": res.Success,
		})

	})

	r.PUT("/crypto/", func(c *gin.Context) {
		post := postpb.Post{}

		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Missing values"})
			return
		}

		req := &postpb.UpdatePostRequest{
			Post: &post,
		}
		res, err := client.UpdatePost(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, res.Post)
	})

	r.PUT("/crypto/upvote", func(c *gin.Context) {
		post := postpb.Post{}
		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Missing values"})
			return
		}
		req := &postpb.UpVoteRequest{
			Post: &post,
		}

		res, err := client.UpVote(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, res.Post)
	})

	r.PUT("/crypto/downvote", func(c *gin.Context) {
		post := postpb.Post{}

		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Missing values"})
			return
		}

		req := &postpb.DownVoteRequest{
			Post: &post,
		}

		res, err := client.DownVote(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, res.Post)
	})

	if err := r.Run(":8052"); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
