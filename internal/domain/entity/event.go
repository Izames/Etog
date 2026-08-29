package entity

import (
	"time"

	"github.com/pgvector/pgvector-go"
	"gorm.io/datatypes"
)

type Event struct {
	Id          int             `gorm:"primary_key;column:id"`
	Organizer   int             `gorm:"column:organizer"`
	Name        string          `gorm:"column:name"`
	Price       float64         `gorm:"column:price"`
	Address     string          `gorm:"column:address"`
	City        string          `gorm:"column:city"`
	Cords       pgvector.Vector `gorm:"column:cords"`
	Type        string          `gorm:"column:type"`
	Date        time.Time       `gorm:"column:date"`
	Description string          `gorm:"column:description"`
	Passed      string          `gorm:"column:passed"`
	MaxPeople   int             `gorm:"column:max_people"`
	Media       datatypes.JSON  `gorm:"column:media"`
	Contact     int             `gorm:"column:contact"`
	Deleted     bool            `gorm:"column:deleted"`
}
