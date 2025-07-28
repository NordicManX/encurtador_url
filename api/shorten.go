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
	// Garante que a conexão com o BD esteja ativa
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("ERRO: MONGODB_URI não definida.")
	}
	EnsureDBConnection(mongoURI)

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	longURL := r.FormValue("url")
	if !isValidURL(longURL) {
		http.Error(w, "URL inválida", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existingURL URL
	err := urlsCollection.FindOne(ctx, bson.M{"long_url": longURL}).Decode(&existingURL)
	if err == nil {
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "https://" + r.Host + "/"
		}
		fmt.Fprintf(w, "%s%s", baseURL, existingURL.ShortCode)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("Erro ao consultar URL: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	var shortCode string
	for {
		shortCode = generateShortCode()
		count, err := urlsCollection.CountDocuments(ctx, bson.M{"_id": shortCode})
		if err != nil {
			log.Printf("Erro CountDocuments: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
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
		log.Printf("Erro ao inserir URL: %v", err)
		http.Error(w, "Erro ao salvar URL", http.StatusInternalServerError)
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://" + r.Host + "/"
	}
	fmt.Fprintf(w, "%s%s", baseURL, shortCode)
}
