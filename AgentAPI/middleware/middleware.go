package middleware

import (
	"cloud.google.com/go/firestore"
	"firebase.google.com/go/storage"
	"github.com/gin-gonic/gin"
)

func FirestoreClientMiddleware(client *firestore.Client, storageClient *storage.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("firestore", client)
		c.Set("storageClient", storageClient)
		c.Next()
	}
}
