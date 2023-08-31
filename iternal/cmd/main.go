package main

import (
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"net/http"
	"testAvito/iternal/handlers"
)

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/user", handlers.CreateUser).Methods("POST")
	router.HandleFunc("/segment", handlers.CreateSegmentHandler).Methods("POST")
	router.HandleFunc("/segment/{segment_name}", handlers.DeleteSegmentHandler).Methods("DELETE")
	router.HandleFunc("/user/{user_id}/segments", handlers.UpdateUserSegmentsHandler).Methods("POST")
	router.HandleFunc("/user/{user_id}/active_segments", handlers.GetUserActiveSegmentsHandler).Methods("GET")

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
