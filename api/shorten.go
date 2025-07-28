package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler processes the request to shorten a URL.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Ensure the database connection is active.
	EnsureDBConnection(os.Getenv("MONGODB_URI"))

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	longURL := r.FormValue("url")
	if !isValidURL(longURL) {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the URL has already been shortened.
	var existingURL URL
	err := urlsCollection.FindOne(ctx, bson.M{"long_url": longURL}).Decode(&existingURL)
	if err == nil {
		// If it exists, return the existing short URL.
		baseURL := getBaseURL(r)
		fmt.Fprintf(w, "%s%s", baseURL, existingURL.ShortCode)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("Error finding document: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Generate a unique short code.
	var shortCode string
	for {
		shortCode = generateShortCode()
		count, err := urlsCollection.CountDocuments(ctx, bson.M{"_id": shortCode})
		if err != nil {
			log.Printf("Error counting documents: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		if count == 0 {
			break
		}
	}

	newURL := URL{
		ShortCode: shortCode,
		LongURL:   longURL,
		CreatedAt: time.Now(),
	}

	// Insert the new URL into the database.
	_, err = urlsCollection.InsertOne(ctx, newURL)
	if err != nil {
		log.Printf("Error inserting document: %v", err)
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	// Return the new short URL.
	baseURL := getBaseURL(r)
	fmt.Fprintf(w, "%s%s", baseURL, shortCode)
}

// getBaseURL determines the base URL for the response.
func getBaseURL(r *http.Request) string {
	// Use BASE_URL from environment variables if available.
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		return baseURL
	}
	// Otherwise, construct it from the request.
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/", scheme, r.Host)
}
