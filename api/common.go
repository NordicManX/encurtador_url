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

type URL struct {
	ShortCode string    `bson:"_id,omitempty"`
	LongURL   string    `bson:"long_url"`
	CreatedAt time.Time `bson:"created_at"`
}

var mongoClient *mongo.Client
var urlsCollection *mongo.Collection
var once sync.Once // Garante que a conexão seja estabelecida apenas uma vez

// EnsureDBConnection garante que a conexão com o DB esteja ativa.
// Usa sync.Once para ser "thread-safe" e eficiente.
func EnsureDBConnection(mongoURI string) {
	once.Do(func() {
		log.Println("Iniciando conexão com o MongoDB...")
		clientOptions := options.Client().ApplyURI(mongoURI)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		mongoClient, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
		}

		err = mongoClient.Ping(ctx, nil)
		if err != nil {
			log.Fatalf("Erro ao pingar o MongoDB: %v", err)
		}
		log.Println("Conectado ao MongoDB Atlas!")

		urlsCollection = mongoClient.Database("url_shortener").Collection("urls")
		createIndexes(ctx)
	})
}

func createIndexes(ctx context.Context) {
	longURLIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "long_url", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := urlsCollection.Indexes().CreateOne(ctx, longURLIndexModel)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") &&
			!strings.Contains(err.Error(), "collection already has an index") {
			log.Printf("Erro ao criar índice para long_url: %v", err)
		}
	} else {
		log.Println("Índice para long_url criado/verificado.")
	}
}

func generateShortCode() string {
	const length = 7
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func isValidURL(url string) bool {
	re := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/\S*)?$`)
	return re.MatchString(url)
}
