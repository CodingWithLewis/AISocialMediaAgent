package user

import (
	"cloud.google.com/go/firestore"
	"encoding/base64"
	"firebase.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"strings"
)

type User struct {
	ID          string   `json:"id" firestore:"id"`
	Bio         *string  `json:"bio" firestore:"bio"`
	Name        string   `json:"name" firestore:"name"`
	Theme       string   `json:"theme" firestore:"theme"`
	Accent      string   `json:"accent" firestore:"accent"`
	Website     *string  `json:"website" firestore:"website"`
	Location    *string  `json:"location" firestore:"location"`
	Username    string   `json:"username" firestore:"username"`
	PhotoBase64 string   `json:"photoBase64"`
	Verified    bool     `json:"verified" firestore:"verified"`
	Following   []string `json:"following" firestore:"following"`
	Followers   []string `json:"followers" firestore:"followers"`
}

func CreateUser(c *gin.Context) {
	var data = User{}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, _ := c.MustGet("firestore").(*firestore.Client)
	storageClient, _ := c.MustGet("storageClient").(*storage.Client)

	// Save image in Cloud Storage
	imageData, err := base64.StdEncoding.DecodeString(data.PhotoBase64[strings.IndexByte(data.PhotoBase64, ',')+1:])

	filename := uuid.New().String() + ".jpg"

	imagePath := data.ID + "/images/" + filename

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

	// Get the URL of the image
	imgURL := wc.Attrs().MediaLink

	_, err = client.Collection("users").Doc(data.ID).Set(c, map[string]interface{}{
		"id": data.ID, "bio": data.Bio,
		"name":        data.Name,
		"theme":       data.Theme,
		"accent":      data.Accent,
		"website":     data.Website,
		"username":    data.Username,
		"photoBase64": data.PhotoBase64,
		"verified":    data.Verified,
		"following":   data.Following,
		"followers":   data.Followers,
		"createdAt":   firestore.ServerTimestamp,
		"updatedAt":   firestore.ServerTimestamp,
		"photoURL":    imgURL,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})

	return
}
