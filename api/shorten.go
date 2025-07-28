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

func Handler(w http.ResponseWriter, r *http.Request) {
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

	var existingURL URL
	err := urlsCollection.FindOne(ctx, bson.M{"long_url": longURL}).Decode(&existingURL)
	if err == nil {
		baseURL := getBaseURL(r)
		fmt.Fprintf(w, "%s%s", baseURL, existingURL.ShortCode)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("Error finding document: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

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

	_, err = urlsCollection.InsertOne(ctx, newURL)
	if err != nil {
		log.Printf("Error inserting document: %v", err)
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	baseURL := getBaseURL(r)
	fmt.Fprintf(w, "%s%s", baseURL, shortCode)
}

func getBaseURL(r *http.Request) string {
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		return baseURL
	}
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/", scheme, r.Host)
}
