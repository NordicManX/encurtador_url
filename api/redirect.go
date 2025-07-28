package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo" // Importe mongo aqui para initMongo
)

// AS SEGUINTES DEFINIÇÕES FORAM REMOVIDAS POIS AGORA ESTÃO EM api/common.go:
// - type URL struct { ... }
// - var mongoClient *mongo.Client
// - var urlsCollection *mongo.Collection

// O init() permanece aqui, mas agora chama ConnectDB do common.go
func init() {
	log.Println("Initializing MongoDB connection (redirect.go init)..")
	mongoURI := os.Getenv("MONGODB_URI") // Lê do ambiente da Vercel
	if mongoURI == "" {
		log.Fatal("ERRO: MONGODB_URI não definida no ambiente da Vercel.")
	}
	ConnectDB(mongoURI) // Chama a função de common.go
}

// Handler é a função de entrada para a função serverless da Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Não precisamos mais verificar mongoClient == nil aqui,
	// pois o init() deve ter garantido a conexão na cold start.

	shortCode := r.URL.Query().Get("shortCode")

	if shortCode == "" {
		shortCode = strings.TrimPrefix(r.URL.Path, "/")
		if strings.Contains(shortCode, "/") {
			http.NotFound(w, r)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result URL // URL struct vem de common.go
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
