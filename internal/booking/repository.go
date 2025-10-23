package booking

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
)

type DBConnection interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Repo struct {
	conn DBConnection
}

func NewRepo(conn DBConnection) *Repo {
	return &Repo{conn: conn}
}

func (r *Repo) CreateBooking(booking *Booking) error {
	query := `
	INSERT INTO bookings (user_id, name, contact, datetime, event_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`
	row := r.conn.QueryRow(context.Background(), query, booking.UserID, booking.Name, booking.Contact, booking.Datetime, booking.EventID)
	err := row.Scan(&booking.ID)
	return err
}

func (r *Repo) GetAllBooking() ([]Booking, error) {
	var bookings []Booking
	query := `
		SELECT id, name, contact, datetime, event_id FROM bookings`
	rows, err := r.conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bookingItem Booking
		var eventID *string
		if err := rows.Scan(&bookingItem.ID, &bookingItem.Name, &bookingItem.Contact, &bookingItem.Datetime, &eventID); err != nil {
			logrus.WithError(err).Error("Failed to scan row in GetAllBooking")
			continue
		}
		bookingItem.EventID = eventID
		bookings = append(bookings, bookingItem)
	}
	return bookings, rows.Err()
}

func (r *Repo) DeleteBookingByID(id int) error {
	query := `DELETE FROM bookings WHERE id = $1`
	_, err := r.conn.Exec(context.Background(), query, id)
	return err
}

func (r *Repo) GetUserBookings(userID int64) ([]Booking, error) {
	var bookings []Booking
	// Используем $1 вместо ?
	query := "SELECT id, user_id, name, contact, datetime, event_id FROM bookings WHERE user_id = $1"
	rows, err := r.conn.Query(context.Background(), query, userID)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).Error("Failed to query user bookings")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var booking Booking
		var eventID *string
		if err := rows.Scan(&booking.ID, &booking.UserID, &booking.Name, &booking.Contact, &booking.Datetime, &eventID); err != nil {
			logrus.WithError(err).Error("Failed to scan row in GetUserBookings")
			continue
		}
		booking.EventID = eventID
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		logrus.WithError(err).Error("Error iterating over user bookings rows")
		return nil, err
	}

	return bookings, nil
}

func (r *Repo) GetBookingByID(id int) (*Booking, error) {
	var booking Booking
	query := "SELECT id, user_id, name, contact, datetime, event_id FROM bookings WHERE id = $1"
	var eventID *string
	err := r.conn.QueryRow(context.Background(), query, id).Scan(&booking.ID, &booking.UserID, &booking.Name, &booking.Contact, &booking.Datetime, &eventID)
	if err != nil {
		return nil, err
	}
	booking.EventID = eventID
	return &booking, nil
}

func (r *Repo) GetUpcomingBookings(from, to time.Time) ([]Booking, error) {
	var bookings []Booking
	query := "SELECT id, user_id, name, contact, datetime, event_id FROM bookings WHERE datetime >= $1 AND datetime < $2 AND user_id IS NOT NULL"
	rows, err := r.conn.Query(context.Background(), query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var booking Booking
		var eventID *string
		if err := rows.Scan(&booking.ID, &booking.UserID, &booking.Name, &booking.Contact, &booking.Datetime, &eventID); err != nil {
			logrus.WithError(err).Error("Failed to scan row in GetUpcomingBookings")
			continue
		}
		booking.EventID = eventID
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}
