package handler // Mantenha o pacote handler

import (
	"context"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Definição da estrutura URL
type URL struct {
	ShortCode string    `bson:"_id,omitempty"`
	LongURL   string    `bson:"long_url"`
	CreatedAt time.Time `bson:"created_at"`
}

// Variáveis globais para a conexão MongoDB
var mongoClient *mongo.Client
var urlsCollection *mongo.Collection

// initMongo é agora uma função comum para inicializar o DB
func initMongo(mongoURI string) {
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
}

// createIndexes é uma função comum para criar índices
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

// generateShortCode é uma função comum para gerar shortcodes
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

// isValidURL é uma função comum para validar URLs
func isValidURL(url string) bool {
	re := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/\S*)?$`)
	return re.MatchString(url)
}
