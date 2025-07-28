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

func Handler(w http.ResponseWriter, r *http.Request) {
	// Garante que a conexão com o BD esteja ativa
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("ERRO: MONGODB_URI não definida.")
	}
	EnsureDBConnection(mongoURI)

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
		http.NotFound(w, r)
		return
	} else if err != nil {
		log.Printf("Erro ao buscar long_url: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, result.LongURL, http.StatusMovedPermanently)
}
