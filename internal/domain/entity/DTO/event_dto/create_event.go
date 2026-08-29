package event_dto

import "time"

type EventAdd struct {
	Name        string    `form:"name"`
	Price       float64   `form:"price"`
	Address     string    `form:"address"`
	City        string    `form:"city"`
	Latitude    float32   `form:"latitude"`
	Longitude   float32   `form:"longitude"`
	Type        string    `form:"type"`
	Date        time.Time `form:"date" time_format:"2006-01-02T15:04:05Z07:00"`
	Description string    `form:"description"`
	Passed      string    `form:"passed"`
	MaxPeople   int       `form:"max_people"`
	Phone       string    `form:"phone"`
	Mail        string    `form:"mail"`
	Telegram    string    `form:"telegram"`
}
