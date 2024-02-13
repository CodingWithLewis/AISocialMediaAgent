package tweet

import (
	"AgentAPI/utils"
	"cloud.google.com/go/firestore"
	"context"
	"encoding/base64"
	"firebase.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"log"
	"net/http"
	"os"
	"time"
)

type Parent struct {
	Id       string `json:"id" firestore:"id,omitempty"`
	Username string `json:"username" firestore:"username,omitempty"`
}

type Image struct {
	Id  string `json:"id" firestore:"id,omitempty"`
	Src string `json:"src" firestore:"src,omitempty"`
	Alt string `json:"alt" firestore:"alt,omitempty"`
}

type Tweet struct {
	Id            string     `json:"id" firestore:"id,omitempty"`
	Text          string     `json:"text" binding:"required" firestore:"text"`
	Parent        *Parent    `json:"parent" firestore:"parent"`
	CreatedBy     string     `json:"createdBy" binding:"required" firestore:"createdBy"`
	CreatedAt     time.Time  `json:"createdAt" binding:"required" firestore:"createdAt"`
	CreatedByName string     `json:"createdByName" firestore:"createdByName,omitempty"`
	UpdatedAt     *time.Time `json:"updatedAt,omitempty" firestore:"updatedAt,omitempty"`
	UserReplies   int        `json:"userReplies" firestore:"userReplies"`
	UserRetweets  []string   `json:"userRetweets" binding:"required" firestore:"userRetweets"`
	UserLikes     []string   `json:"userLikes" binding:"required" firestore:"userLikes"`
	Images        []Image    `json:"images" firestore:"images"`
}

type ImageUploadRequest struct {
	ImagePrompt string `json:"imagePrompt"`
	Tweet       Tweet  `json:"tweet"`
}

func UpdateTimeAndId(data *Tweet) {

	// Update the data with the current time
	data.Id = uuid.New().String()
	data.CreatedAt = time.Now()
	data.UpdatedAt = &data.CreatedAt
}

func findPathToRootTweet(ctx context.Context, client *firestore.Client, tweetID string) ([]Tweet, error) {
	var tweets []Tweet

	for {
		doc, err := client.Collection("tweets").Doc(tweetID).Get(ctx)
		if err != nil {
			return nil, err // Handle error appropriately
		}

		var t Tweet
		err = doc.DataTo(&t)
		if err != nil {
			return nil, err
		}

		// Prepend the tweet to the slice to maintain the order from the given tweet to the root
		tweets = append([]Tweet{t}, tweets...)

		if t.Parent.Id == "" {
			// Reached the root tweet
			break
		} else {
			// Update tweetID to the ID of the parent for the next iteration
			tweetID = t.Parent.Id
		}
	}

	return tweets, nil
}
func GetTweetTimeline(c *gin.Context) {
	client, _ := c.MustGet("firestore").(*firestore.Client)

	docs := client.Collection("tweets").
		Where("parent", "==", nil).
		Where("id", "!=", "").
		Limit(5).
		Documents(c)

	var data []Tweet

	for {
		doc, err := docs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalln("error:", err)

		}

		var t Tweet
		err = doc.DataTo(&t)
		if err != nil {
			log.Fatalln("error:", err)
		}
		t.Id = doc.Ref.ID
		data = append(data, t)
	}

	c.JSON(http.StatusOK, data)
}

func GetTweet(c *gin.Context) {
	client, _ := c.MustGet("firestore").(*firestore.Client)

	tweetId := c.Param("id")
	pathToRoot, err := findPathToRootTweet(c, client, tweetId)

	if err != nil {
		log.Printf("Error finding path to root tweet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding path to root tweet"})
		return
	}

	// Directly return the path to root tweets as JSON
	c.JSON(http.StatusOK, pathToRoot)
}

func CreateTweet(c *gin.Context) {
	client, ok := c.MustGet("firestore").(*firestore.Client)
	if !ok || client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get Firestore client"})
		return
	}

	var data Tweet
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	UpdateTimeAndId(&data)

	// Query the name of the person submitting this request
	user, err := client.Collection("users").Doc(data.CreatedBy).Get(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	data.CreatedByName = user.Data()["name"].(string)

	// If this tweet has a parent, increment the number of replies
	if data.Parent != nil {
		parentDoc, err := client.Collection("tweets").Doc(data.Parent.Id).Get(c)
		if err != nil {
			log.Printf("Error getting parent tweet: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting parent tweet"})
			return
		}
		var parent Tweet
		if err = parentDoc.DataTo(&parent); err != nil {
			log.Printf("Error converting parent tweet to struct: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error converting parent tweet to struct"})
			return
		}
		parent.UserReplies++

		// Ensure UserRetweets is an empty slice instead of nil if there are no retweets
		if parent.UserRetweets == nil {
			parent.UserRetweets = []string{}
		}

		// Similarly, ensure UserLikes is an empty slice instead of nil if there are no likes
		if parent.UserLikes == nil {
			parent.UserLikes = []string{}
		}
		if _, err = client.Collection("tweets").Doc(data.Parent.Id).Set(c, parent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating parent tweet"})
			return
		}
	}

	// Insert the data into Firestore
	if _, err = client.Collection("tweets").Doc(data.Id).Set(c, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tweet created"})
}

func TweetImage(c *gin.Context) {
	client, _ := c.MustGet("firestore").(*firestore.Client)
	storageClient, _ := c.MustGet("storageClient").(*storage.Client)

	var data = ImageUploadRequest{}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//	Make image request to Stability AI API
	base64String := utils.GenerateAIImage(data.ImagePrompt)
	if base64String == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate image. No Base64String"})
		return
	}
	// Decode String
	imageData, err := base64.StdEncoding.DecodeString(base64String)

	filename := uuid.New().String() + ".jpg"

	// Create image path
	imagePath := data.Tweet.CreatedBy + "/images/" + filename

	if err != nil {
		log.Fatalln("Could not decode base64 string", err)
	}

	bucketName := os.Getenv("BUCKET_NAME")
	if os.Getenv("DEBUG") == "true" {
		bucketName = "aitwitter-dev.appspot.com"
	}
	bucket, err := storageClient.Bucket(bucketName)

	if err != nil {
		log.Fatalln("Could not get default bucket", err)
	}

	wc := bucket.Object(imagePath).NewWriter(c)

	if _, err = wc.Write(imageData); err != nil {
		log.Fatalf("error writing to Firebase Storage: %v", err)
	}
	if err := wc.Close(); err != nil {
		log.Fatalf("error closing Firebase Storage writer: %v", err)
	}
	imgURL := wc.Attrs().MediaLink

	// Create Tweet with Image Path
	data.Tweet.Images = append(data.Tweet.Images, Image{
		Id:  filename,
		Src: imgURL,
		Alt: data.ImagePrompt,
	})

	UpdateTimeAndId(&data.Tweet)

	_, err = client.Collection("tweets").Doc(data.Tweet.Id).Set(c, data.Tweet)

	if err != nil {
		log.Fatalln("Error creating tweet with image", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "tweet created"})

}
