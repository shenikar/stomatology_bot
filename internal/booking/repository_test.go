package booking

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

func TestBookingRepo_CreateBooking(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	repo := NewRepo(mock)

	eventID := gofakeit.UUID()
	booking := &Booking{
		UserID:   gofakeit.Int64(),
		Name:     gofakeit.Name(),
		Contact:  gofakeit.Phone(),
		Datetime: gofakeit.Date(),
		EventID:  &eventID,
	}

	mock.ExpectQuery(`INSERT INTO bookings`).
		WithArgs(booking.UserID, booking.Name, booking.Contact, booking.Datetime, booking.EventID).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int(gofakeit.Int64())))

	err = repo.CreateBooking(booking)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBookingRepo_GetBookingByID(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	repo := NewRepo(mock)

	bookingID := int(gofakeit.Int64())
	eventID := gofakeit.UUID()

	rows := pgxmock.NewRows([]string{"id", "user_id", "name", "contact", "datetime", "event_id"}).
		AddRow(bookingID, gofakeit.Int64(), gofakeit.Name(), gofakeit.Phone(), gofakeit.Date(), &eventID)

	mock.ExpectQuery(`SELECT id, user_id, name, contact, datetime, event_id FROM bookings WHERE id = \$1`).
		WithArgs(bookingID).
		WillReturnRows(rows)

	booking, err := repo.GetBookingByID(bookingID)
	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, bookingID, booking.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBookingRepo_DeleteBookingById(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	repo := NewRepo(mock)

	bookingID := int(gofakeit.Int64())

	mock.ExpectExec(`DELETE FROM bookings WHERE id = \$1`).
		WithArgs(bookingID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteBookingByID(bookingID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBookingRepo_GetUserBookings(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	repo := NewRepo(mock)

	userID := gofakeit.Int64()
	eventID1 := gofakeit.UUID()
	eventID2 := gofakeit.UUID()

	rows := pgxmock.NewRows([]string{"id", "user_id", "name", "contact", "datetime", "event_id"}).
		AddRow(int(gofakeit.Int64()), userID, gofakeit.Name(), gofakeit.Phone(), gofakeit.Date(), &eventID1).
		AddRow(int(gofakeit.Int64()), userID, gofakeit.Name(), gofakeit.Phone(), gofakeit.Date(), &eventID2)

	mock.ExpectQuery(`SELECT id, user_id, name, contact, datetime, event_id FROM bookings WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	bookings, err := repo.GetUserBookings(userID)
	assert.NoError(t, err)
	assert.NotNil(t, bookings)
	assert.Len(t, bookings, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBookingRepo_GetAllBooking(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	repo := NewRepo(mock)

	eventID1 := gofakeit.UUID()
	eventID2 := gofakeit.UUID()

	rows := pgxmock.NewRows([]string{"id", "name", "contact", "datetime", "event_id"}).
		AddRow(int(gofakeit.Int64()), gofakeit.Name(), gofakeit.Phone(), gofakeit.Date(), &eventID1).
		AddRow(int(gofakeit.Int64()), gofakeit.Name(), gofakeit.Phone(), gofakeit.Date(), &eventID2)

	mock.ExpectQuery(`SELECT id, name, contact, datetime, event_id FROM bookings`).
		WillReturnRows(rows)

	bookings, err := repo.GetAllBooking()
	assert.NoError(t, err)
	assert.NotNil(t, bookings)
	assert.Len(t, bookings, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBookingRepo_CreateBooking_Error(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	repo := NewRepo(mock)

	eventID := gofakeit.UUID()
	booking := &Booking{
		UserID:   gofakeit.Int64(),
		Name:     gofakeit.Name(),
		Contact:  gofakeit.Phone(),
		Datetime: gofakeit.Date(),
		EventID:  &eventID,
	}

	mock.ExpectQuery(`INSERT INTO bookings`).
		WithArgs(booking.UserID, booking.Name, booking.Contact, booking.Datetime, booking.EventID).
		WillReturnError(assert.AnError)

	err = repo.CreateBooking(booking)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBookingRepo_GetUserBookings_Error(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	repo := NewRepo(mock)
	userID := gofakeit.Int64()

	mock.ExpectQuery(`SELECT id, user_id, name, contact, datetime, event_id FROM bookings WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnError(assert.AnError)

	_, err = repo.GetUserBookings(userID)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBookingRepo_GetBookingByID_Error(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	repo := NewRepo(mock)
	bookingID := int(gofakeit.Int64())

	mock.ExpectQuery(`SELECT id, user_id, name, contact, datetime, event_id FROM bookings WHERE id = \$1`).
		WithArgs(bookingID).
		WillReturnError(assert.AnError)

	_, err = repo.GetBookingByID(bookingID)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBookingRepo_DeleteBookingById_Error(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	repo := NewRepo(mock)
	bookingID := int(gofakeit.Int64())

	mock.ExpectExec(`DELETE FROM bookings WHERE id = \$1`).
		WithArgs(bookingID).
		WillReturnError(assert.AnError)

	err = repo.DeleteBookingByID(bookingID)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
