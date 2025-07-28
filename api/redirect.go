package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler processes the request to redirect a short URL to its original long URL.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Ensure the database connection is active.
	EnsureDBConnection(os.Getenv("MONGODB_URI"))

	shortCode := r.URL.Query().Get("shortCode")
	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result URL
	err := urlsCollection.FindOne(ctx, bson.M{"_id": shortCode}).Decode(&result)

	if err == mongo.ErrNoDocuments {
		// If the short code is not found, return a 404.
		http.NotFound(w, r)
		return
	} else if err != nil {
		// For any other error, log it and return a server error.
		log.Printf("Error finding document for redirect: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Redirect to the long URL.
	http.Redirect(w, r, result.LongURL, http.StatusMovedPermanently)
}
