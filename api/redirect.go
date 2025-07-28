package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// NO LONGER NEEDED: type URL struct { ... }
// NO LONGER NEEDED: var mongoClient *mongo.Client
// NO LONGER NEEDED: var urlsCollection *mongo.Collection

// Handler é a função de entrada para a função serverless da Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Inicializa a conexão com o banco de dados se ainda não estiver inicializada
	if mongoClient == nil { // Verifica se o cliente já foi inicializado
		log.Println("Initializing MongoDB connection (redirect.go Handler)..")
		mongoURI := os.Getenv("MONGODB_URI")
		if mongoURI == "" {
			http.Error(w, "Erro interno do servidor: MONGODB_URI não definida.", http.StatusInternalServerError)
			log.Fatal("ERRO FATAL: MONGODB_URI não definida no ambiente da Vercel.")
			return
		}
		ConnectDB(mongoURI) // Chama a função de common.go
	}

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
