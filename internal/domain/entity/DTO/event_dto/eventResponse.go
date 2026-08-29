package event_dto

import (
	"time"
)

type EventResponse struct {
	Id          int       `json:"primary_key;column:id"`
	Organizer   string    `json:"column:organizer"`
	Name        string    `json:"column:name"`
	Price       float64   `json:"column:price"`
	Address     string    `json:"column:address"`
	City        string    `json:"column:city"`
	Type        string    `json:"column:type"`
	Date        time.Time `json:"column:date"`
	Description string    `json:"column:description"`
	Passed      string    `json:"column:passed"`
	MaxPeople   int       `json:"column:max_people"`
	Media       []string  `json:"column:media"`
	Phone       string    `json:"column:phone"`
	Mail        string    `json:"column:mail"`
	Telegram    string    `json:"column:telegram"`
	Deleted     bool      `json:"column:deleted"`
}
