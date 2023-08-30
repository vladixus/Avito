package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"
)

var db *sql.DB

type Segment struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Segments []Segment `json:"segments"`
}

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
	fmt.Println(users)
	for _, Name := range users.Name {
		err = db.QueryRow("SELECT username FROM users WHERE username=$1", Name).Scan(&Name)
		if err == nil {
			copy = true
			log.Println(err)
			canceldUsers.Name = append(canceldUsers.Name, Name)
			continue
			return
		}

		_, err = db.Exec("INSERT INTO users (username) VALUES ($1)", Name)
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
	err = db.QueryRow("SELECT COUNT(*) FROM segments WHERE name_segment = $1", segmentName.Name).Scan(&count)
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
	_, err = db.Exec("INSERT INTO segments (name_segment) VALUES ($1)", segmentName.Name)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func DeleteSegmentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	segmentName := vars["segment_name"]

	_, err := db.Exec("DELETE FROM segments WHERE name_segment = $1", segmentName)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateUserSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
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
		segmentName = strings.ToUpper(segmentName)
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM segments WHERE name_segment = $1", segmentName).Scan(&count)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		// Если сегмент с таким именем уже существует, вернуть ошибку
		if count == 0 {
			http.Error(w, "Сегмент с данным именем не существует", http.StatusConflict)
			return
		}

		_, err = db.Exec("INSERT INTO user_segments (user_id, segment_id) SELECT $1, id FROM segments WHERE name_segment = $2", userID, segmentName)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера(Insert)", http.StatusInternalServerError)
			return
		}
	}

	for _, segmentName := range request.Remove {
		_, err = db.Exec("DELETE FROM user_segments WHERE user_id = $1 AND segment_id = (SELECT id FROM segments WHERE name_segment = $2)", userID, segmentName)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func GetUserActiveSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	rows, err := db.Query("SELECT segments.id, segments.name_segment FROM user_segments JOIN segments ON user_segments.segment_id = segments.id WHERE user_segments.user_id = $1", userID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var segments []Segment
	for rows.Next() {
		var segment Segment
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

func init() {
	// loads values from .env into the system
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
	db, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД, %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL
	
)`)
	if err != nil {
		log.Fatalf("Error creating table users, %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS segments  (
	id SERIAL PRIMARY KEY,
    name_segment VARCHAR(255) NOT NULL
	
)`)
	if err != nil {
		log.Fatalf("Error creating table segments, %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS user_segments  (
	user_id INT REFERENCES users(id),
    segment_id INT REFERENCES segments(id),
    PRIMARY KEY (user_id, segment_id)
	
)`)
	if err != nil {
		log.Fatalf("Error creating table user_segments, %s", err)
	}
}

func main() {
	user, exists := os.LookupEnv("USER")
	passw, exists := os.LookupEnv("PASSWORD")
	dbname, exists := os.LookupEnv("DBNAME")
	host, exists := os.LookupEnv("HOST")

	if !exists {
		log.Fatal("error reading env file")
	}
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=5432 dbname=%s sslmode=disable", user, passw, host, dbname)
	var err error
	db, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/user", CreateUser).Methods("POST")
	router.HandleFunc("/segment", CreateSegmentHandler).Methods("POST")
	router.HandleFunc("/segment/{segment_name}", DeleteSegmentHandler).Methods("DELETE")
	router.HandleFunc("/user/{user_id}/segments", UpdateUserSegmentsHandler).Methods("POST")
	router.HandleFunc("/user/{user_id}/active_segments", GetUserActiveSegmentsHandler).Methods("GET")

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
