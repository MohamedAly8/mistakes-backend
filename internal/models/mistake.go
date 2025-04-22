package models

type Mistake struct {
	ID         int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category   string `json:"category"`
}
