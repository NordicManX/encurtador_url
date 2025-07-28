package handler

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os" // Para ler variáveis de ambiente da Vercel
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Remova a variável global 'baseURL' daqui e obtenha-a do ambiente
// const baseURL = "http://localhost:8080/"

type URL struct {
	ShortCode string    `bson:"_id,omitempty"`
	LongURL   string    `bson:"long_url"`
	CreatedAt time.Time `bson:"created_at"`
}

var mongoClient *mongo.Client
var urlsCollection *mongo.Collection

// init() é executado uma vez por cold start da função
func init() {
	log.Println("Initializing MongoDB connection for shorten.go...")
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
	log.Println("Conectado ao MongoDB Atlas (shorten.go)!")

	urlsCollection = mongoClient.Database("url_shortener").Collection("urls")
	createIndexes(ctx) // Garante que os índices existam
}

func createIndexes(ctx context.Context) {
	longURLIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "long_url", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := urlsCollection.Indexes().CreateOne(ctx, longURLIndexModel)
	if err != nil {
		// Em serverless, erros de índice podem ser mais comuns no init
		// Apenas logar, não fatalizar, se o índice já existir
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

// Handler é a função de entrada para a função serverless da Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
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
		// Obtém a BASE_URL do ambiente da Vercel
		vercelBaseURL := os.Getenv("BASE_URL")
		if vercelBaseURL == "" {
			vercelBaseURL = "https://" + r.Host + "/" // Fallback para o host da requisição
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
		shortCode = generateShortCode()
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

	newURL := URL{
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

	// Obtém a BASE_URL do ambiente da Vercel para a resposta
	vercelBaseURL := os.Getenv("BASE_URL")
	if vercelBaseURL == "" {
		vercelBaseURL = "https://" + r.Host + "/"
	}
	fmt.Fprintf(w, "%s%s", vercelBaseURL, shortCode)
}
