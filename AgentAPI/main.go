package main

import (
	tweet "AgentAPI/controllers/tweets"
	user "AgentAPI/controllers/users"
	middleware "AgentAPI/middleware"
	"AgentAPI/utils"
	"context"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	ctx := context.Background()

	client, storageClient := utils.CreateFirestoreClient(ctx)
	defer client.Close()
	debug := os.Getenv("DEBUG") == "true"
	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.Default()
	r.Use(middleware.FirestoreClientMiddleware(client, storageClient))
	// Tweet Routes
	r.GET("/tweets/", tweet.GetTweetTimeline)
	r.GET("/tweets/:id/", tweet.GetTweet)
	r.POST("/tweet/", tweet.CreateTweet)
	r.POST("/image-tweet/", tweet.TweetImage)

	// User Routes
	r.POST("/user/", user.CreateUser)

	r.Run(":3000")
}
