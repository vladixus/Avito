package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"
	"testAvito/iternal/models"
)

var DB *sql.DB

func init() {

	// загрузка файла
	if err := godotenv.Load(".env"); err != nil {
		log.Print("No .env file found")
	}
	user, exists := os.LookupEnv("USER")
	passw, exists := os.LookupEnv("PASSWORD")
	dbname, exists := os.LookupEnv("DBNAME")
	host, exists := os.LookupEnv("HOST")

	if !exists {
		log.Fatal("error reading env file")
	}
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=5432 dbname=%s sslmode=disable", user, passw, host, dbname)
	var err error
	DB, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД, %s", err)
	}

	//Проверка и создание таблиц
	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL
	
)`)
	if err != nil {
		log.Fatalf("Error creating table users, %s", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS segments  (
	id SERIAL PRIMARY KEY,
    name_segment VARCHAR(255) NOT NULL
	
)`)
	if err != nil {
		log.Fatalf("Error creating table segments, %s", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS user_segments  (
	user_id INT REFERENCES users(id),
    segment_id INT REFERENCES segments(id),
    PRIMARY KEY (user_id, segment_id)
	
)`)
	if err != nil {
		log.Fatalf("Error creating table user_segments, %s", err)
	}

}

// Создание пользователя
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var users struct {
		Name []string `json:"name"`
	}
	var canceldUsers struct {
		Name []string `json:"name"`
	}
	var copy = false

	err := json.NewDecoder(r.Body).Decode(&users)
	if err != nil {
		log.Println(err)
		http.Error(w, "Не распарсилось", http.StatusBadRequest)
		return
	}

	for _, Name := range users.Name {
		err = DB.QueryRow("SELECT username FROM users WHERE username=$1", Name).Scan(&Name)
		if err == nil {
			copy = true
			log.Println(err)
			canceldUsers.Name = append(canceldUsers.Name, Name)
			continue
			return
		}

		_, err = DB.Exec("INSERT INTO users (username) VALUES ($1)", Name)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
	}
	if copy {
		w.Header().Set("Content-Type", "application/json")
		response, err := json.Marshal(canceldUsers)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Дублированыые пользователи:\n"))
		w.Write(response)
	}
	w.WriteHeader(http.StatusCreated)
}

// Создание сегмента
func CreateSegmentHandler(w http.ResponseWriter, r *http.Request) {
	var segmentName struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&segmentName)
	if err != nil {
		log.Println(err)
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	// Проверка существования сегмента с таким именем
	var count int
	segmentName.Name = strings.ToUpper(segmentName.Name)
	err = DB.QueryRow("SELECT COUNT(*) FROM segments WHERE name_segment = $1", segmentName.Name).Scan(&count)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Если сегмент с таким именем уже существует, вернуть ошибку
	if count > 0 {
		http.Error(w, "Сегмент с данным именем уже существует", http.StatusConflict)
		return
	}

	// Вставка нового сегмента
	_, err = DB.Exec("INSERT INTO segments (name_segment) VALUES ($1)", segmentName.Name)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusNotModified)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Удаление сегмента
func DeleteSegmentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	segmentName := vars["segment_name"]
	segmentName = strings.ToUpper(segmentName)
	_, err := DB.Exec("DELETE FROM segments WHERE name_segment = $1", segmentName)
	if err != nil {
		log.Println(err)
		http.Error(w, "Пользователь состоит в сегмента, удалите сперва связь", http.StatusNotModified)
		return
	}

	w.WriteHeader(http.StatusGone)
}

// Изменение сегмента пользователя
func UpdateUserSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Если сегмент с таким именем уже существует, вернуть ошибку
	if count == 0 {
		http.Error(w, "Пользователя с данным id не существует", http.StatusConflict)
		return
	}

	var request struct {
		Add    []string `json:"add"`
		Remove []string `json:"remove"`
	}

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	for _, segmentName := range request.Add {
		if segmentName == "" {
			continue
		}
		segmentName = strings.ToUpper(segmentName)
		var count int
		err = DB.QueryRow("SELECT COUNT(*) FROM segments WHERE name_segment = $1", segmentName).Scan(&count)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		if count == 0 {
			http.Error(w, "Сегмент с данным именем не существует "+segmentName, http.StatusConflict)
			return
		}

		_, err = DB.Exec("INSERT INTO user_segments (user_id, segment_id) SELECT $1, id FROM segments WHERE name_segment = $2", userID, segmentName)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	for _, segmentName := range request.Remove {
		segmentName = strings.ToUpper(segmentName)
		var count int
		err = DB.QueryRow("SELECT COUNT(*) FROM segments WHERE name_segment = $1", segmentName).Scan(&count)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		if count == 0 {
			http.Error(w, "Сегмент с данным именем не существует "+segmentName, http.StatusConflict)
			return
		}
		segmentName = strings.ToUpper(segmentName)
		_, err = DB.Exec("DELETE FROM user_segments WHERE user_id = $1 AND segment_id = (SELECT id FROM segments WHERE name_segment = $2)", userID, segmentName)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Активные сегменты пользователя
func GetUserActiveSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	rows, err := DB.Query("SELECT segments.id, segments.name_segment FROM user_segments JOIN segments ON user_segments.segment_id = segments.id WHERE user_segments.user_id = $1", userID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var segments []models.Segment
	for rows.Next() {
		var segment models.Segment
		err := rows.Scan(&segment.ID, &segment.Name)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
		segments = append(segments, segment)
	}

	response, err := json.Marshal(segments)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
