package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os" // Para usar os.Getenv para ler variáveis de ambiente
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv" // NOVO: Importar godotenv

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

const (
	baseURL = "http://localhost:8080/" // Para desenvolvimento local
)

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

func createIndexes(ctx context.Context) {
	longURLIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "long_url", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := urlsCollection.Indexes().CreateOne(ctx, longURLIndexModel)
	if err != nil {
		log.Printf("Erro ao criar índice para long_url: %v", err)
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

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprintf(w, "%s%s", baseURL, existingURL.ShortCode)
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

	fmt.Fprintf(w, "%s%s", baseURL, shortCode)
}

func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
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
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		log.Printf("Erro ao buscar long_url no MongoDB: %v", err)
		return
	}

	http.Redirect(w, r, result.LongURL, http.StatusMovedPermanently)
}

func main() {

	err := godotenv.Load(".env")
	if err != nil {

		log.Println("Aviso: Arquivo .env não encontrado ou erro ao carregar:", err)
		log.Println("Tentando ler variáveis de ambiente do sistema.")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("ERRO: Variável de ambiente MONGODB_URI não definida. Por favor, defina-a no arquivo .env ou no ambiente do sistema.")
	}

	initMongo(mongoURI)
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			redirectURLHandler(w, r)
			return
		}
		tmpl, err := template.ParseFiles("./static/index.html")
		if err != nil {
			log.Printf("Erro ao carregar template: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/shorten", shortenURLHandler)

	log.Println("Servidor iniciado na porta :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
