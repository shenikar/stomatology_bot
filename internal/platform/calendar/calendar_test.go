package calendar

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func newTestCalendarService(serverURL string) (*CalendarService, error) {
	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithEndpoint(serverURL), option.WithoutAuthentication())
	if err != nil {
		return nil, err
	}
	return &CalendarService{
		srv:           srv,
		calID:         gofakeit.UUID(),
		workStartHour: 9,
		workEndHour:   18,
	}, nil
}

func TestCalendarService_GetFreeSlots(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пример ответа от Google Calendar API
		response := `{
			"items": [
				{
					"start": {"dateTime": "2025-10-24T10:00:00+03:00"},
					"end": {"dateTime": "2025-10-24T11:00:00+03:00"}
				},
				{
					"start": {"dateTime": "2025-10-24T14:00:00+03:00"},
					"end": {"dateTime": "2025-10-24T15:00:00+03:00"}
				}
			]
		}`
		fmt.Fprintln(w, response)
	}))
	defer server.Close()

	service, err := newTestCalendarService(server.URL)
	assert.NoError(t, err)

	testDate := time.Date(2025, 10, 24, 0, 0, 0, 0, time.UTC)
	freeSlots, err := service.GetFreeSlots(testDate)

	assert.NoError(t, err)
	assert.NotNil(t, freeSlots)
	// Ожидаем, что слоты в 10 и 14 будут заняты
	// 9, 11, 12, 13, 15, 16, 17, 18 - свободны (8 слотов)
	// 9, 11, 12, 13, 15, 16, 17, 18, 19 - свободны (9 слотов)
	assert.Len(t, freeSlots, 8)
}

func TestCalendarService_CreateEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf(`{
			"id": "%s",
			"htmlLink": "%s"
		}`, gofakeit.UUID(), gofakeit.URL())
		fmt.Fprintln(w, response)
	}))
	defer server.Close()

	service, err := newTestCalendarService(server.URL)
	assert.NoError(t, err)

	now := gofakeit.Date()
	link, eventID, err := service.CreateEvent(gofakeit.Sentence(), gofakeit.Sentence(), now, now.Add(time.Hour))

	assert.NoError(t, err)
	assert.NotEmpty(t, link)
	assert.NotEmpty(t, eventID)
}

func TestCalendarService_DeleteEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// При успешном удалении Google API возвращает пустой ответ со статусом 204
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service, err := newTestCalendarService(server.URL)
	assert.NoError(t, err)

	err = service.DeleteEvent(gofakeit.UUID())
	assert.NoError(t, err)
}

func TestCalendarService_GetFreeSlots_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service, err := newTestCalendarService(server.URL)
	assert.NoError(t, err)

	testDate := time.Date(2025, 10, 24, 0, 0, 0, 0, time.UTC)
	_, err = service.GetFreeSlots(testDate)
	assert.Error(t, err)
}

func TestCalendarService_CreateEvent_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service, err := newTestCalendarService(server.URL)
	assert.NoError(t, err)

	now := gofakeit.Date()
	_, _, err = service.CreateEvent(gofakeit.Sentence(), gofakeit.Sentence(), now, now.Add(time.Hour))
	assert.Error(t, err)
}

func TestCalendarService_IsSlotFree(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"items": []}`
		fmt.Fprintln(w, response)
	}))
	defer server.Close()

	service, err := newTestCalendarService(server.URL)
	assert.NoError(t, err)

	now := gofakeit.Date()
	isFree, err := service.IsSlotFree(now, now.Add(time.Hour))
	assert.NoError(t, err)
	assert.True(t, isFree)
}

func TestCalendarService_IsSlotFree_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service, err := newTestCalendarService(server.URL)
	assert.NoError(t, err)

	now := gofakeit.Date()
	_, err = service.IsSlotFree(now, now.Add(time.Hour))
	assert.Error(t, err)
}

func TestNewCalendarService_Error(t *testing.T) {
	// Тест на ошибку чтения файла credentials
	_, err := NewCalendarService("non-existent-file.json", "test-id", 9, 18)
	assert.Error(t, err)
}

func TestCalendarService_GetFreeSlots_InvalidDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ответ не важен, так как мы ожидаем ошибку до запроса
	}))
	defer server.Close()

	service, err := newTestCalendarService(server.URL)
	assert.NoError(t, err)

	// Передаем некорректную дату (нулевое время)
	_, err = service.GetFreeSlots(time.Time{})
	// Ожидаем ошибку, так как LoadLocation вернет ошибку для нулевого времени
	assert.Error(t, err)
}

func TestCalendarService_DeleteEvent_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service, err := newTestCalendarService(server.URL)
	assert.NoError(t, err)

	err = service.DeleteEvent(gofakeit.UUID())
	assert.Error(t, err)
}