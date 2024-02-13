package utils

import (
	"bytes"
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"
	"google.golang.org/api/option"
	"io"
	"log"
	"net/http"
	"os"
)

func CreateFirestoreClient(ctx context.Context) (*firestore.Client, *storage.Client) {
	debug := os.Getenv("DEBUG")
	println("DEBUG: " + debug)
	if debug == "true" {
		// Set the Firestore emulator address
		os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
		// Set the Storage emulator address
		os.Setenv("STORAGE_EMULATOR_HOST", "localhost:9199")
	}
	var saPath option.ClientOption
	if debug != "true" {
		// Use credentials file only if not in debug mode
		saPath = option.WithCredentialsFile("aitwitter-2a8ac-5d46648017ce.json")
	} else {
		// Use no authentication for emulator
		saPath = option.WithoutAuthentication()
	}

	// Initialize Firestore client
	firestoreClient, err := firestore.NewClient(ctx, "aitwitter-2a8ac", saPath)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}

	// Initialize Cloud Storage client
	var app *firebase.App
	if debug != "true" {
		// Initialize Firebase app with credentials in non-debug mode
		config := &firebase.Config{
			StorageBucket: "aitwitter-2a8ac.appspot.com",
		}
		app, err = firebase.NewApp(ctx, config, saPath)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		// Initialize Firebase app without authentication in debug mode
		app, err = firebase.NewApp(ctx, nil, option.WithoutAuthentication())
		if err != nil {
			log.Fatalln(err)
		}
	}

	storageClient, err := app.Storage(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	return firestoreClient, storageClient
}

func GenerateAIImage(prompt string) string {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error in env file")
	}

	url := "https://api.stability.ai/v1/generation/stable-diffusion-xl-1024-v1-0/text-to-image"

	body := map[string]interface{}{
		"steps":     40,
		"width":     1024,
		"height":    1024,
		"seed":      0,
		"cfg_scale": 5.0,
		"samples":   1,
		"text_prompts": []map[string]interface{}{
			{
				"text":   prompt,
				"weight": 1.0,
			},
			{
				"text":   "blurry, bad",
				"weight": -1.0,
			},
		},
	}

	jsonData, err := json.Marshal(body)

	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Example: Set additional headers
	req.Header.Set("Authorization", "Bearer "+os.Getenv("STABILITY_AI_API_KEY"))
	rClient := &http.Client{}

	response, err := rClient.Do(req)

	if response.StatusCode != http.StatusOK {
		log.Println("Response: ", response.Body)
		log.Fatalln("Error in response from Stability AI")
	}

	if err != nil {
		log.Fatalf("Error when doing request to Stability AI", err)
	}

	defer response.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln("Error reading response body", err)
	}

	base64String := gjson.GetBytes(resBody, "artifacts.0.base64").String()

	return base64String

}
