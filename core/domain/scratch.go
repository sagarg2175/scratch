package domain

import (
	"time"
)

type Scratch struct {
	Id              *int       `json:"id" db:"id"`
	Name            *string    `json:"name" db:"name"`
	LastUpdatedDate *time.Time `json:"lastupdated" db:"lastupdated"`
	Password        *string    `json:"password" db:"password"`
}
