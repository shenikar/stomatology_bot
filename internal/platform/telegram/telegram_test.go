package telegram

import (
	"stomatology_bot/internal/booking"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock BotAPI
type MockBotAPI struct {
	mock.Mock
}

func (m *MockBotAPI) Send(c tgbot.Chattable) (tgbot.Message, error) {
	args := m.Called(c)
	return args.Get(0).(tgbot.Message), args.Error(1)
}

func (m *MockBotAPI) GetUpdatesChan(_ tgbot.UpdateConfig) tgbot.UpdatesChannel {
	// Для тестов нам не нужно получать реальные обновления
	return make(chan tgbot.Update)
}

func (m *MockBotAPI) Request(c tgbot.Chattable) (*tgbot.APIResponse, error) {
	args := m.Called(c)
	return args.Get(0).(*tgbot.APIResponse), args.Error(1)
}

// Mock CalendarService
type MockCalendarService struct {
	mock.Mock
}

func (m *MockCalendarService) GetFreeSlots(date time.Time) ([]time.Time, error) {
	args := m.Called(date)
	return args.Get(0).([]time.Time), args.Error(1)
}

func (m *MockCalendarService) CreateEvent(summary, description string, start, end time.Time) (string, string, error) {
	args := m.Called(summary, description, start, end)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockCalendarService) DeleteEvent(eventID string) error {
	args := m.Called(eventID)
	return args.Error(0)
}

func (m *MockCalendarService) IsSlotFree(start time.Time, end time.Time) (bool, error) {
	args := m.Called(start, end)
	return args.Bool(0), args.Error(1)
}

// Mock BookingRepo
type MockBookingRepo struct {
	mock.Mock
}

func (m *MockBookingRepo) CreateBooking(b *booking.Booking) error {
	args := m.Called(b)
	return args.Error(0)
}

func (m *MockBookingRepo) GetUserBookings(userID int64) ([]booking.Booking, error) {
	args := m.Called(userID)
	return args.Get(0).([]booking.Booking), args.Error(1)
}

func (m *MockBookingRepo) GetBookingByID(id int) (*booking.Booking, error) {
	args := m.Called(id)
	return args.Get(0).(*booking.Booking), args.Error(1)
}

func (m *MockBookingRepo) DeleteBookingByID(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBookingRepo) GetAllBooking() ([]booking.Booking, error) {
	args := m.Called()
	return args.Get(0).([]booking.Booking), args.Error(1)
}

func TestTgBot_handleBookCommand(t *testing.T) {
	mockAPI := new(MockBotAPI)
	bot := &TgBot{
		api:        mockAPI,
		userStates: make(map[int64]*UserState),
	}
	chatID := gofakeit.Int64()

	// Ожидаем, что будет отправлено сообщение с клавиатурой
	mockAPI.On("Send", mock.Anything).Return(tgbot.Message{}, nil).Once()

	bot.handleBookCommand(chatID)

	// Проверяем, что состояние пользователя установлено правильно
	assert.Equal(t, StateAwaitingDate, bot.userStates[chatID].State)
	mockAPI.AssertExpectations(t)
}

func TestTgBot_handleTimeSelection(t *testing.T) {
	mockAPI := new(MockBotAPI)
	bot := &TgBot{
		api:        mockAPI,
		userStates: make(map[int64]*UserState),
	}
	chatID := gofakeit.Int64()
	slot := gofakeit.Date()

	update := tgbot.Update{
		CallbackQuery: &tgbot.CallbackQuery{
			ID:      gofakeit.UUID(),
			From:    &tgbot.User{ID: chatID},
			Message: &tgbot.Message{Chat: &tgbot.Chat{ID: chatID}},
			Data:    "time_" + slot.Format(time.RFC3339),
		},
	}

	// Ожидаем, что будет отправлено сообщение с запросом имени
	mockAPI.On("Send", mock.Anything).Return(tgbot.Message{}, nil).Once()
	// Ожидаем, что будет отправлен ответ на callback query
	mockAPI.On("Request", mock.Anything).Return(&tgbot.APIResponse{}, nil).Once()

	bot.handleCallbackQuery(update)

	// Проверяем, что состояние пользователя установлено правильно
	assert.Equal(t, StateAwaitingName, bot.userStates[chatID].State)
	// Сравниваем время, обнулив наносекунды
	assert.Equal(t, slot.Truncate(time.Second), bot.userStates[chatID].TempTime.Truncate(time.Second))
	mockAPI.AssertExpectations(t)
}
