package handler

import (
	"context"
	"log"
	"net/http"
	"os" // Para ler variáveis de ambiente da Vercel
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type URL struct { // Re-defina a struct URL aqui, pois cada função é um pacote 'main' separado
	ShortCode string    `bson:"_id,omitempty"`
	LongURL   string    `bson:"long_url"`
	CreatedAt time.Time `bson:"created_at"`
}

var mongoClient *mongo.Client
var urlsCollection *mongo.Collection

// init() é executado uma vez por cold start da função
func init() {
	log.Println("Initializing MongoDB connection for redirect.go...")
	mongoURI := os.Getenv("MONGODB_URI") // Lê do ambiente da Vercel
	if mongoURI == "" {
		log.Fatal("ERRO: MONGODB_URI não definida no ambiente da Vercel.")
	}

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
	log.Println("Conectado ao MongoDB Atlas (redirect.go)!")

	urlsCollection = mongoClient.Database("url_shortener").Collection("urls")
	// Não precisa criar índices aqui, já foi feito em shorten.go
}

// Handler é a função de entrada para a função serverless da Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// O shortCode virá da URL via regex no vercel.json
	// A Vercel coloca as named capture groups em r.URL.Query()
	shortCode := r.URL.Query().Get("shortCode") // Nome da captura no vercel.json

	if shortCode == "" { // Fallback ou para casos sem match de regex
		// Tenta extrair do path se a regex falhar ou para outros cenários
		shortCode = strings.TrimPrefix(r.URL.Path, "/")
		if strings.Contains(shortCode, "/") { // Se ainda tiver barras, não é um shortcode simples
			http.NotFound(w, r)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result URL
	err := urlsCollection.FindOne(ctx, bson.M{"_id": shortCode}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		log.Printf("Erro ao buscar long_url no MongoDB: %v", err)
		return
	}

	http.Redirect(w, r, result.LongURL, http.StatusMovedPermanently)
}
