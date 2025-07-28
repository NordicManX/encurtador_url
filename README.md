# 🔗 Encurtador de URL com Go e Vercel

![Badge de Status do Deploy na Vercel](https://therealsujitk-vercel-badge.vercel.app/?app=encurtador-url-nordicmanx)
![Badge da Linguagem Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Badge do MongoDB](https://img.shields.io/badge/MongoDB-47A248?style=for-the-badge&logo=mongodb&logoColor=white)
![Badge da Licença MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)

Um projeto de encurtador de URLs simples, rápido e eficiente, construído com Go para o backend, MongoDB como banco de dados e implantado como Funções Serverless na Vercel.

---

### ✨ Visão Geral

![Imagem da interface do encurtador de URL](https://i.imgur.com/8YgL4r6.png)

---

## 🚀 Tecnologias Utilizadas

Este projeto foi construído utilizando as seguintes tecnologias:

* **Backend:** [Go (Golang)](https://go.dev/)
* **Banco de Dados:** [MongoDB Atlas](https://www.mongodb.com/atlas)
* **Frontend:** HTML5, CSS3 com [Bootstrap 5](https://getbootstrap.com/) e JavaScript puro.
* **Plataforma de Deploy:** [Vercel](https://vercel.com/)

---

## 📋 Funcionalidades

* **Encurtar URLs:** Transforma URLs longas em links curtos e fáceis de compartilhar.
* **Redirecionamento Rápido:** Redireciona os links curtos para as URLs originais de forma eficiente.
* **Interface Limpa:** Uma interface de usuário simples e intuitiva para encurtar os links.
* **API Simples:** Um endpoint de API para criar os links (`/api/shorten`).
* **Serverless:** Arquitetura de baixo custo e alta escalabilidade, ideal para a Vercel.

---

## 🛠️ Como Executar Localmente

Para executar este projeto em sua máquina local, siga os passos abaixo.

### Pré-requisitos

* [Go](https://go.dev/doc/install) (versão 1.18 ou superior)
* [MongoDB](https://www.mongodb.com/try/download/community) rodando localmente ou uma conta no [MongoDB Atlas](https://www.mongodb.com/cloud/atlas/register)
* [Vercel CLI](https://vercel.com/docs/cli) (opcional, para testes locais)

### Passos

1.  **Clone o repositório:**
    ```bash
    git clone [https://github.com/NordicManX/encurtador_url.git](https://github.com/NordicManX/encurtador_url.git)
    cd encurtador_url
    ```

2.  **Crie o arquivo de ambiente:**
    Crie um arquivo chamado `.env` na raiz do projeto e adicione sua string de conexão do MongoDB:
    ```.env
    MONGODB_URI="sua-string-de-conexao-do-mongodb-aqui"
    BASE_URL="http://localhost:3000/"
    ```

3.  **Instale as dependências do Go:**
    ```bash
    go mod tidy
    ```

4.  **Execute o projeto com a Vercel CLI:**
    Este é o método recomendado, pois simula o ambiente da Vercel.
    ```bash
    vercel dev
    ```
    A aplicação estará disponível em `http://localhost:3000`.

---



## 👤 Autor

Feito com ❤️ por **NordicManX**.

[![GitHub](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/NordicManX)

---

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.
