package booking

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

type BookingRepo struct {
	conn *pgx.Conn
}

func NewBookingRepo(conn *pgx.Conn) *BookingRepo {
	return &BookingRepo{conn: conn}
}

func (r *BookingRepo) CreateBooking(booking *Booking) error {
	query := `
	INSERT INTO bookings (user_id, name, contact, datetime, event_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`
	row := r.conn.QueryRow(context.Background(), query, booking.UserID, booking.Name, booking.Contact, booking.Datetime, booking.EventID)
	err := row.Scan(&booking.ID)
	return err
}

func (r *BookingRepo) GetAllBooking() ([]Booking, error) {
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
			log.Printf("Ошибка при сканировании строки: %v", err)
			continue
		}
		bookingItem.EventID = eventID
		bookings = append(bookings, bookingItem)
	}
	return bookings, rows.Err()
}

func (r *BookingRepo) DeleteBookingById(id int) error {
	query := `DELETE FROM bookings WHERE id = $1`
	_, err := r.conn.Exec(context.Background(), query, id)
	return err
}

func (r *BookingRepo) GetUserBookings(userID int64) ([]Booking, error) {
	var bookings []Booking
	// Используем $1 вместо ?
	query := "SELECT id, user_id, name, contact, datetime, event_id FROM bookings WHERE user_id = $1"
	rows, err := r.conn.Query(context.Background(), query, userID)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var booking Booking
		var eventID *string // Используем *string для сканирования
		if err := rows.Scan(&booking.ID, &booking.UserID, &booking.Name, &booking.Contact, &booking.Datetime, &eventID); err != nil {
			log.Printf("Ошибка при сканировании строки: %v", err)
			continue
		}
		booking.EventID = eventID
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при переборе строк: %v", err)
		return nil, err
	}

	return bookings, nil
}

func (r *BookingRepo) GetBookingByID(id int) (*Booking, error) {
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
