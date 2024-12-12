package domain

import (
	"time"
)

type Booking struct {
	ID       int       `db:"id"`
	Name     string    `db:"name"`
	Contact  string    `db:"contact"`
	Datetime time.Time `db:"datetime"`
}
