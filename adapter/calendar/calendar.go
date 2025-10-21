package calendar

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarService - сервис для работы с Google Calendar
type CalendarService struct {
	srv   *calendar.Service
	calID string
}

// NewCalendarService создает новый сервис для работы с календарем
func NewCalendarService(credentialFile, calendarID string) (*CalendarService, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If you are authorizing a service account, USE THE FOLLOWING CODE instead of the three lines above.
	config, err := google.JWTConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client := config.Client(ctx)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}

	return &CalendarService{
		srv:   srv,
		calID: calendarID,
	}, nil
}

// CreateEvent создает новое событие в календаре
func (s *CalendarService) CreateEvent(summary, description string, start, end time.Time) (string, error) {
	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: "Europe/Moscow",
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
			TimeZone: "Europe/Moscow",
		},
	}

	event, err := s.srv.Events.Insert(s.calID, event).Do()
	if err != nil {
		return "", fmt.Errorf("unable to create event: %v", err)
	}

	return event.HtmlLink, nil
}

// GetFreeSlots возвращает список свободных слотов на определенный день
func (s *CalendarService) GetFreeSlots(date time.Time) ([]time.Time, error) {
	// Устанавливаем начало и конец дня
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Запрашиваем события на этот день
	events, err := s.srv.Events.List(s.calID).
		TimeMin(startOfDay.Format(time.RFC3339)).
		TimeMax(endOfDay.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve next ten of the user's events: %v", err)
	}

	// Определяем рабочие часы (например, с 9:00 до 18:00)
	workStartHour := 9
	workEndHour := 18

	var freeSlots []time.Time

	// Генерируем все возможные слоты в течение рабочего дня
	for hour := workStartHour; hour < workEndHour; hour++ {
		slot := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, date.Location())
		isBusy := false

		// Проверяем, занят ли этот слот
		for _, item := range events.Items {
			eventStart, _ := time.Parse(time.RFC3339, item.Start.DateTime)
			eventEnd, _ := time.Parse(time.RFC3339, item.End.DateTime)

			if slot.After(eventStart) && slot.Before(eventEnd) {
				isBusy = true
				break
			}
			if slot.Equal(eventStart) {
				isBusy = true
				break
			}
		}

		if !isBusy {
			freeSlots = append(freeSlots, slot)
		}
	}

	return freeSlots, nil
}
