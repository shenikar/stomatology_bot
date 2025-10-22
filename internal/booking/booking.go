package booking

import (
	"time"
)

type Booking struct {
	ID       int       `db:"id"`
	UserID   int64     `db:"user_id"`
	Name     string    `db:"name"`
	Contact  string    `db:"contact"`
	Datetime time.Time `db:"datetime"`
	EventID  *string   `db:"event_id"`
}
