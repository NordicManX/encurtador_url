package main

import (
	"database/sql"
	"fmt"
	"html/template" // Usado para servir o arquivo HTML
	"log"
	"net/http"
	"regexp" // Para validar URLs
	// Para conversão de string para int (usado em decodeBase62, embora não diretamente no handler)
	"strings" // Para manipulação de strings (trimPrefix)

	_ "github.com/mattn/go-sqlite3" // Driver do SQLite3. O '_' indica que o pacote é importado apenas pelos seus efeitos colaterais (registro do driver).
)

// URL representa a estrutura de uma URL em nosso sistema.
type URL struct {
	ID        int64
	LongURL   string
	ShortCode string
}

var db *sql.DB // Variável global para a conexão com o banco de dados.

const (
	// base62Charset contém os 62 caracteres usados para codificar os IDs do banco de dados em códigos curtos.
	base62Charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// baseURL é a base da URL para as URLs encurtadas. Altere para o seu domínio se estiver em produção.
	baseURL = "http://localhost:8080/"
)

// initDB inicializa a conexão com o banco de dados SQLite e cria a tabela 'urls' se ela não existir.
func initDB(dataSourceName string) {
	var err error
	db, err = sql.Open("sqlite3", dataSourceName) // Abre a conexão com o banco de dados.
	if err != nil {
		log.Fatalf("Erro ao abrir o banco de dados: %v", err) // Se houver erro, encerra o programa.
	}

	// SQL para criar a tabela 'urls'.
	// id: Chave primária auto-incrementável (irá gerar o ID numérico que usamos para o shortcode).
	// long_url: A URL original longa, deve ser única para evitar duplicatas.
	// short_code: O código curto gerado, também deve ser único.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		long_url TEXT NOT NULL UNIQUE,
		short_code TEXT UNIQUE
	);`
	_, err = db.Exec(createTableSQL) // Executa a query para criar a tabela.
	if err != nil {
		log.Fatalf("Erro ao criar a tabela: %v", err) // Se houver erro, encerra o programa.
	}
	log.Println("Banco de dados SQLite inicializado e tabela 'urls' verificada.")
}

// encodeBase62 converte um ID numérico (gerado pelo banco de dados) para uma string Base62.
// Isso cria códigos curtos e únicos a partir de IDs incrementais.
func encodeBase62(id int64) string {
	if id == 0 {
		return string(base62Charset[0]) // Retorna '0' para o ID 0.
	}

	var result []byte
	for id > 0 {
		remainder := id % 62                              // Obtém o resto da divisão por 62.
		result = append(result, base62Charset[remainder]) // Adiciona o caractere correspondente.
		id /= 62                                          // Divide o ID por 62 para a próxima iteração.
	}

	// A string é construída de trás para frente, então precisamos invertê-la.
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

// decodeBase62 converte uma string Base62 de volta para um ID numérico.
// Esta função é útil para buscar a URL longa a partir do shortcode.
func decodeBase62(shortCode string) (int64, error) {
	var id int64 = 0
	base := int64(len(base62Charset))
	for _, char := range shortCode {
		idx := strings.IndexRune(base62Charset, char) // Encontra o índice do caractere no charset.
		if idx == -1 {
			return 0, fmt.Errorf("caractere inválido na string Base62: %c", char)
		}
		id = id*base + int64(idx) // Converte o caractere para seu valor numérico e adiciona ao ID.
	}
	return id, nil
}

// isValidURL verifica se uma string tem o formato básico de uma URL válida.
// Usa uma expressão regular simples. Em um ambiente de produção, seria mais robusta.
func isValidURL(url string) bool {
	re := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/\S*)?$`)
	return re.MatchString(url)
}

// shortenURLHandler lida com as requisições POST para encurtar uma URL.
func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// longURL é extraída do formulário HTML.
	longURL := r.FormValue("url")
	if !isValidURL(longURL) {
		http.Error(w, "URL inválida", http.StatusBadRequest)
		return
	}

	// 1. Tenta buscar a URL longa no banco de dados para ver se já foi encurtada.
	var existingShortCode string
	err := db.QueryRow("SELECT short_code FROM urls WHERE long_url = ?", longURL).Scan(&existingShortCode)
	if err == nil {
		// Se a URL já existe, retorna o código curto já existente.
		fmt.Fprintf(w, "%s%s", baseURL, existingShortCode)
		return
	} else if err != sql.ErrNoRows {
		// Lida com outros erros de banco de dados.
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		log.Printf("Erro ao consultar URL existente: %v", err)
		return
	}

	// 2. Se a URL não existe, insere a URL longa no banco de dados.
	res, err := db.Exec("INSERT INTO urls (long_url) VALUES (?)", longURL)
	if err != nil {
		http.Error(w, "Erro ao salvar URL", http.StatusInternalServerError)
		log.Printf("Erro ao inserir URL: %v", err)
		return
	}

	// Obtém o ID auto-incrementado da URL recém-inserida.
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Erro ao obter ID", http.StatusInternalServerError)
		log.Printf("Erro ao obter LastInsertId: %v", err)
		return
	}

	// 3. Converte o ID para Base62 para gerar o código curto.
	shortCode := encodeBase62(id)

	// 4. Atualiza o registro no banco de dados com o short_code gerado.
	_, err = db.Exec("UPDATE urls SET short_code = ? WHERE id = ?", shortCode, id)
	if err != nil {
		http.Error(w, "Erro ao atualizar short code", http.StatusInternalServerError)
		log.Printf("Erro ao atualizar short code: %v", err)
		return
	}

	// Retorna a URL curta completa para o frontend.
	fmt.Fprintf(w, "%s%s", baseURL, shortCode)
}

// redirectURLHandler lida com as requisições GET para redirecionar URLs curtas para as longas.
func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	// Extrai o código curto da URL (ex: /abc123 -> abc123).
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" {
		http.NotFound(w, r) // Se não há código, retorna 404.
		return
	}

	var longURL string
	// Busca a URL longa no banco de dados usando o short_code.
	err := db.QueryRow("SELECT long_url FROM urls WHERE short_code = ?", shortCode).Scan(&longURL)
	if err == sql.ErrNoRows {
		http.NotFound(w, r) // Se o código curto não for encontrado, retorna 404.
		return
	} else if err != nil {
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		log.Printf("Erro ao buscar long_url: %v", err)
		return
	}

	// Redireciona o navegador para a URL longa.
	http.Redirect(w, r, longURL, http.StatusMovedPermanently) // Status 301 para redirecionamento permanente.
}

func main() {
	// Inicializa o banco de dados. O arquivo será criado ou aberto em 'shortener.db' no mesmo diretório do executável.
	initDB("./shortener.db")
	defer db.Close() // Garante que a conexão com o banco de dados seja fechada ao final da execução.

	// Handler para servir arquivos estáticos (HTML, CSS, JS) da pasta 'static'.
	// http.StripPrefix remove "/static/" da URL para que http.FileServer encontre os arquivos corretamente.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Handler para a URL raiz ("/"). Serve o index.html.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// Se a URL não for a raiz, tenta tratar como um shortcode para redirecionamento.
			redirectURLHandler(w, r)
			return
		}
		// Carrega o template HTML e o executa para servir a página.
		tmpl, err := template.ParseFiles("./static/index.html")
		if err != nil {
			log.Printf("Erro ao carregar template: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil) // Executa o template, enviando a página HTML para o navegador.
	})

	// Handler para o endpoint de encurtar URLs (recebe POST do formulário).
	http.HandleFunc("/shorten", shortenURLHandler)

	log.Println("Servidor iniciado na porta :8080")
	// Inicia o servidor HTTP na porta 8080. log.Fatal fará o programa sair se houver um erro.
	log.Fatal(http.ListenAndServe(":8080", nil))
}
