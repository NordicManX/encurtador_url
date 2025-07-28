package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os" // Para ler variáveis de ambiente da Vercel
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// NO LONGER NEEDED: // Remova a variável global 'baseURL' daqui e obtenha-a do ambiente
// NO LONGER NEEDED: type URL struct { ... }
// NO LONGER NEEDED: var mongoClient *mongo.Client
// NO LONGER NEEDED: var urlsCollection *mongo.Collection
// NO LONGER NEEDED: func createIndexes(ctx context.Context) { ... }
// NO LONGER NEEDED: func generateShortCode() string { ... }
// NO LONGER NEEDED: func isValidURL(url string) bool { ... }

// Handler é a função de entrada para a função serverless da Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Inicializa a conexão com o banco de dados se ainda não estiver inicializada
	// Isso é importante porque cada "cold start" de uma função serverless
	// precisa garantir que a conexão está ativa.
	if mongoClient == nil { // Verifica se o cliente já foi inicializado
		log.Println("Initializing MongoDB connection (shorten.go Handler)..")
		mongoURI := os.Getenv("MONGODB_URI")
		if mongoURI == "" {
			http.Error(w, "Erro interno do servidor: MONGODB_URI não definida.", http.StatusInternalServerError)
			log.Fatal("ERRO FATAL: MONGODB_URI não definida no ambiente da Vercel.")
			return
		}
		ConnectDB(mongoURI) // Chama a função de common.go
	}

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
