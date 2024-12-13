package repository

import (
	"context"
	"stomatology_bot/domain"

	"github.com/jackc/pgx/v5"
)

type BookingRepo struct {
	conn *pgx.Conn
}

func NewBookingRepo(conn *pgx.Conn) *BookingRepo {
	return &BookingRepo{conn: conn}
}

func (r *BookingRepo) CreateBooking(booking *domain.Booking) error {
	query := `
	INSERT INTO bookings (name, contact, datetime)
	VALUES ($1, $2, $3) 
	RETURNING id`
	row := r.conn.QueryRow(context.Background(), query, booking.Name, booking.Contact, booking.Datetime)
	err := row.Scan(&booking.ID)
	return err
}

func (r *BookingRepo) GetAllBooking() ([]domain.Booking, error) {
	var bookings []domain.Booking
	query := `
		SELECT id, name, contact, datetime FROM bookings`
	rows, err := r.conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bookingItem domain.Booking
		if err := rows.Scan(&bookingItem.ID, &bookingItem.Name, &bookingItem.Contact, bookingItem.Datetime); err != nil {
			return nil, err
		}
		bookings = append(bookings, bookingItem)
	}
	return bookings, rows.Err()
}

func (r *BookingRepo) DeleteBookingById(id int) error {
	query := `DELETE FROM bookings WHERE id = $1`
	_, err := r.conn.Exec(context.Background(), query, id)
	return err
}
