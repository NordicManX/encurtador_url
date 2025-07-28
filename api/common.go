package handler

import (
	"context"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// URL defines the structure for a shortened URL document in MongoDB.
type URL struct {
	ShortCode string    `bson:"_id,omitempty"`
	LongURL   string    `bson:"long_url"`
	CreatedAt time.Time `bson:"created_at"`
}

var mongoClient *mongo.Client
var urlsCollection *mongo.Collection

// once ensures the database connection is established only once.
var once sync.Once

// EnsureDBConnection establishes a thread-safe connection to MongoDB.
func EnsureDBConnection(mongoURI string) {
	once.Do(func() {
		if mongoURI == "" {
			log.Fatal("FATAL: MONGODB_URI environment variable not set.")
		}

		log.Println("Attempting to connect to MongoDB...")
		clientOptions := options.Client().ApplyURI(mongoURI)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		mongoClient, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		err = mongoClient.Ping(ctx, nil)
		if err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err)
		}
		log.Println("Successfully connected to MongoDB!")

		urlsCollection = mongoClient.Database("url_shortener").Collection("urls")
		createIndexes(ctx)
	})
}

// createIndexes creates necessary indexes for the 'urls' collection.
func createIndexes(ctx context.Context) {
	longURLIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "long_url", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := urlsCollection.Indexes().CreateOne(ctx, longURLIndexModel)
	if err != nil {
		// Don't fail if the index already exists.
		if !strings.Contains(err.Error(), "already exists") {
			log.Printf("Warning: Could not create index for long_url: %v", err)
		}
	} else {
		log.Println("Index for long_url created or already exists.")
	}
}

// generateShortCode creates a random 7-character string.
func generateShortCode() string {
	const length = 7
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// isValidURL checks if a string is a valid URL.
func isValidURL(url string) bool {
	// A simple regex to validate a URL.
	re := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/\S*)?$`)
	return re.MatchString(url)
}
