package main // Define o nome do pacote

// Importa as bibliotecas necessárias
import (
	"database/sql"  // Usado para interagir com o banco de dados
	"encoding/json" // Usado para codificar e decodificar JSON

	// Usado para formatar strings e imprimir no console
	"io/ioutil" // Usado para operações de E/S
	"log"       // Usado para registrar erros
	"net/http"  // Usado para lidar com solicitações HTTP
	"strconv"   // Usado para conversões de string

	_ "github.com/mattn/go-sqlite3" // Driver SQLite para interagir com o banco de dados SQLite
)

// Define a estrutura da mensagem
type Message struct {
	ID   int    `json:"id"`      // ID da mensagem
	Text string `json:"message"` // Texto da mensagem
}

var db *sql.DB // Variável global para a conexão do banco de dados

// Função init é chamada automaticamente antes da função main
func init() {
	var err error
	db, err = sql.Open("sqlite3", "./messages.db") // Abre uma conexão com o banco de dados SQLite
	if err != nil {
		log.Fatal(err) // Registra o erro e para a execução
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY, message TEXT)") // Cria a tabela de mensagens se ela não existir
	if err != nil {
		log.Fatal(err) // Registra o erro e para a execução"
	}
}

// Função para lidar com as solicitações HTTP para mensagens
func messagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method { // Verifica o método da solicitação HTTP
	case "GET": // Se for uma solicitação GET
		rows, err := db.Query("SELECT id, message FROM messages") // Consulta todas as mensagens do banco de dados
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // Retorna um erro 500 se houver um problema
			return
		}
		defer rows.Close() // Garante que os resultados da consulta sejam fechados após o término da função

		var messages []Message // Inicializa uma fatia para armazenar as mensagens
		for rows.Next() {      // Itera sobre cada linha nos resultados da consulta
			var msg Message                     // Inicializa uma nova mensagem
			err = rows.Scan(&msg.ID, &msg.Text) // Atribui os valores da linha à mensagem
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError) // Retorna um erro 500 se houver um problema
				return
			}
			messages = append(messages, msg) // Adiciona a mensagem à fatia de mensagens
		}

		json.NewEncoder(w).Encode(messages) // Codifica as mensagens como JSON e escreve na resposta
	case "POST": // Se for uma solicitação POST
		body, err := ioutil.ReadAll(r.Body) // Lê o corpo da solicitação
		if err != nil {
			http.Error(w, "Erro ao ler o corpo da solicitação", http.StatusBadRequest) // Retorna um erro 400 se houver um problema
			return
		}
		var msg Message                  // Inicializa uma nova mensagem
		err = json.Unmarshal(body, &msg) // Decodifica o corpo da solicitação em uma mensagem
		if err != nil {
			http.Error(w, "Erro ao analisar o corpo da solicitação", http.StatusBadRequest) // Retorna um erro 400 se houver um problema
			return
		}

		result, err := db.Exec("INSERT INTO messages (message) VALUES (?)", msg.Text) // Insere a mensagem no banco de dados
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // Retorna um erro 500 se houver um problema
			return
		}
		id, err := result.LastInsertId() // Obtém o ID da última mensagem inserida
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // Retorna um erro 500 se houver um problema
			return
		}
		msg.ID = int(id) // Atribui o ID à mensagem

		w.WriteHeader(http.StatusCreated) // Define o status da resposta como 201 Created
		json.NewEncoder(w).Encode(msg)    // Codifica a mensagem como JSON e escreve na resposta
	case "PUT": // Se for uma solicitação PUT
		body, err := ioutil.ReadAll(r.Body) // Lê o corpo da solicitação
		if err != nil {
			http.Error(w, "Erro ao ler o corpo da solicitação", http.StatusBadRequest) // Retorna um erro 400 se houver um problema
			return
		}
		var msg Message                  // Inicializa uma nova mensagem
		err = json.Unmarshal(body, &msg) // Decodifica o corpo da solicitação em uma mensagem
		if err != nil {
			http.Error(w, "Erro ao analisar o corpo da solicitação", http.StatusBadRequest) // Retorna um erro 400 se houver um problema
			return
		}

		_, err = db.Exec("UPDATE messages SET message = ? WHERE id = ?", msg.Text, msg.ID) // Atualiza a mensagem no banco de dados
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // Retorna um erro 500 se houver um problema
			return
		}

		json.NewEncoder(w).Encode(msg) // Codifica a mensagem como JSON e escreve na resposta
	case "DELETE": // Se for uma solicitação DELETE
		id, err := strconv.Atoi(r.URL.Query().Get("id")) // Obtém o ID da URL da solicitação
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest) // Retorna um erro 400 se houver um problema
			return
		}

		_, err = db.Exec("DELETE FROM messages WHERE id = ?", id) // Exclui a mensagem do banco de dados
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // Retorna um erro 500 se houver um problema
			return
		}

		w.WriteHeader(http.StatusOK)               // Define o status da resposta como 200 OK
		json.NewEncoder(w).Encode(Message{ID: id}) // Codifica a mensagem como JSON e escreve na resposta
	default:
		http.Error(w, "Método não suportado", http.StatusMethodNotAllowed) // Retorna um erro 405 se o método da solicitação não for suportado
	}
}

// Função principal que inicia o servidor
func main() {
	http.HandleFunc("/messages", messagesHandler) // Define o manipulador para a rota "/messages"
	http.ListenAndServe(":8080", nil)             // Inicia o servidor na porta 8080
}
