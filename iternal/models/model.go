package models

type Segment struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Segments []Segment `json:"segments"`
}
