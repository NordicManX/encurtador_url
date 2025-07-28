package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo" // Importe mongo aqui para initMongo
)

// AS SEGUINTES DEFINIÇÕES FORAM REMOVIDAS POIS AGORA ESTÃO EM api/common.go:
// - type URL struct { ... }
// - var mongoClient *mongo.Client
// - var urlsCollection *mongo.Collection
// - func createIndexes(...)
// - func generateShortCode()
// - func isValidURL(...)

// O init() permanece aqui, mas agora chama ConnectDB do common.go
func init() {
	log.Println("Initializing MongoDB connection (shorten.go init)..")
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
	// Apenas para robustez, mas o init() é o lugar principal.

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	longURL := r.FormValue("url")
	if !isValidURL(longURL) { // isValidURL vem de common.go
		http.Error(w, "URL inválida", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existingURL URL // URL struct vem de common.go
	err := urlsCollection.FindOne(ctx, bson.M{"long_url": longURL}).Decode(&existingURL)
	if err == nil {
		vercelBaseURL := os.Getenv("BASE_URL")
		if vercelBaseURL == "" {
			vercelBaseURL = "https://" + r.Host + "/"
		}
		fmt.Fprintf(w, "%s%s", vercelBaseURL, existingURL.ShortCode)
		return
	} else if err != mongo.ErrNoDocuments {
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		log.Printf("Erro ao consultar URL existente no MongoDB: %v", err)
		return
	}

	var shortCode string
	for {
		shortCode = generateShortCode() // generateShortCode vem de common.go
		count, err := urlsCollection.CountDocuments(ctx, bson.M{"_id": shortCode})
		if err != nil {
			http.Error(w, "Erro ao verificar unicidade do short code", http.StatusInternalServerError)
			log.Printf("Erro CountDocuments: %v", err)
			return
		}
		if count == 0 {
			break
		}
	}

	newURL := URL{ // URL struct vem de common.go
		ShortCode: shortCode,
		LongURL:   longURL,
		CreatedAt: time.Now(),
	}

	_, err = urlsCollection.InsertOne(ctx, newURL)
	if err != nil {
		http.Error(w, "Erro ao salvar URL no MongoDB", http.StatusInternalServerError)
		log.Printf("Erro ao inserir URL no MongoDB: %v", err)
		return
	}

	vercelBaseURL := os.Getenv("BASE_URL")
	if vercelBaseURL == "" {
		vercelBaseURL = "https://" + r.Host + "/"
	}
	fmt.Fprintf(w, "%s%s", vercelBaseURL, shortCode)
}
