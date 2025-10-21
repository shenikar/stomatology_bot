package tbot

import (
	"fmt"
	"log"
	"stomatology_bot/adapter/calendar"
	"stomatology_bot/config"
	"stomatology_bot/domain"
	"stomatology_bot/repository"
	"strings"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const slotDuration = time.Hour

type TgBot struct {
	api         *tgbot.BotAPI
	cfg         *config.Config
	repo        *repository.BookingRepo
	calendarSvc *calendar.CalendarService
}

func NewBot(api *tgbot.BotAPI, cfg *config.Config, repo *repository.BookingRepo, calendarSvc *calendar.CalendarService) *TgBot {
	return &TgBot{
		api:         api,
		cfg:         cfg,
		repo:        repo,
		calendarSvc: calendarSvc,
	}
}

func (b *TgBot) Start() {
	u := tgbot.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	// Обрабатываем обновления
	for update := range updates {
		if update.Message != nil {
			go b.processUpdate(update)
		} else if update.CallbackQuery != nil {
			go b.handleCallbackQuery(update) // Добавили обработку CallbackQuery
		}
	}
}

func (b *TgBot) processUpdate(update tgbot.Update) {
	switch update.Message.Text {
	case "/start":
		msg := tgbot.NewMessage(update.Message.Chat.ID, "Добро пожаловать! Я бот для записи в стоматологическую клинику. Используйте команду /help для получения списка доступных команд.")
		b.api.Send(msg)
	case "/help":
		msg := tgbot.NewMessage(update.Message.Chat.ID, "Список доступных команд:\n/start - Начать работу с ботом\n/help - Помощь\n/book - Записаться на приём\n/mybookings - Вывод всех записей клиента.")
		b.api.Send(msg)
	case "/book":
		b.handleBookCommand(update.Message.Chat.ID)
	case "/mybookings":
		b.handleShowAllBooking(update)
	default:
		b.sendMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /help для получения списка доступных команд.")
	}
}

func (b *TgBot) handleCallbackQuery(update tgbot.Update) {
	if update.CallbackQuery != nil {
		// Простое определение шага по префиксу данных колбэка
		if strings.HasPrefix(update.CallbackQuery.Data, "date_") {
			b.handleDateSelection(update)
		} else if strings.HasPrefix(update.CallbackQuery.Data, "time_") {
			b.handleTimeSelection(update)
		}
	}
}

func (b *TgBot) handleBookCommand(chatID int64) {
	// Предлагаем выбрать дату
	var buttons [][]tgbot.InlineKeyboardButton
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, i)
		button := tgbot.NewInlineKeyboardButtonData(date.Format("02.01.2006"), "date_"+date.Format("2006-01-02"))
		buttons = append(buttons, []tgbot.InlineKeyboardButton{button})
	}

	keyboard := tgbot.NewInlineKeyboardMarkup(buttons...)
	msg := tgbot.NewMessage(chatID, "Выберите дату для записи:")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *TgBot) handleDateSelection(update tgbot.Update) {
	dateStr := strings.TrimPrefix(update.CallbackQuery.Data, "date_")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		b.sendMessage(update.CallbackQuery.Message.Chat.ID, "Неверный формат даты.")
		return
	}

	freeSlots, err := b.calendarSvc.GetFreeSlots(date)
	if err != nil {
		log.Printf("Ошибка получения свободных слотов: %v", err)
		b.sendMessage(update.CallbackQuery.Message.Chat.ID, "Не удалось получить свободные слоты. Попробуйте позже.")
		return
	}

	if len(freeSlots) == 0 {
		b.sendMessage(update.CallbackQuery.Message.Chat.ID, "На выбранную дату нет свободных слотов.")
		return
	}

	var buttons [][]tgbot.InlineKeyboardButton
	for _, slot := range freeSlots {
		button := tgbot.NewInlineKeyboardButtonData(slot.Format("15:04"), "time_"+slot.Format(time.RFC3339))
		buttons = append(buttons, []tgbot.InlineKeyboardButton{button})
	}

	keyboard := tgbot.NewInlineKeyboardMarkup(buttons...)
	msg := tgbot.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите время для записи:")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *TgBot) handleTimeSelection(update tgbot.Update) {
	timeStr := strings.TrimPrefix(update.CallbackQuery.Data, "time_")
	slot, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		b.sendMessage(update.CallbackQuery.Message.Chat.ID, "Неверный формат времени.")
		return
	}

	chatID := update.CallbackQuery.Message.Chat.ID
	userName := update.CallbackQuery.From.UserName

	// Создаем событие в Google Calendar
	summary := fmt.Sprintf("Запись: %s", userName)
	description := fmt.Sprintf("Запись на прием от пользователя %s.", userName)
	link, err := b.calendarSvc.CreateEvent(summary, description, slot, slot.Add(slotDuration))
	if err != nil {
		log.Printf("Ошибка создания события в календаре: %v", err)
		b.sendMessage(chatID, "Не удалось создать запись. Попробуйте позже.")
		return
	}

	// Создаем запись в нашей БД
	booking := &domain.Booking{
		UserID:   chatID,
		Name:     userName,
		Contact:  "N/A", // Можно запросить дополнительно
		Datetime: slot,
	}

	if err := b.repo.CreateBooking(booking); err != nil {
		// Здесь можно было бы добавить логику отката события в календаре,
		// но для простоты пока опустим.
		b.sendMessage(chatID, "Ошибка при сохранении записи в базу данных. Попробуйте снова.")
	} else {
		response := fmt.Sprintf("Вы успешно записаны на %s.\nСсылка на событие: %s", slot.Format("02.01.2006 в 15:04"), link)
		b.sendMessage(chatID, response)
	}
}

func (b *TgBot) handleShowAllBooking(update tgbot.Update) {
	chatID := update.Message.Chat.ID
	bookings, err := b.repo.GetUserBookings(chatID) // Фильтрация на уровне БД
	if err != nil {
		log.Printf("Ошибка получения записей: %v", err)
		b.sendMessage(chatID, "Ошибка при получении записей.")
		return
	}

	if len(bookings) == 0 {
		b.sendMessage(chatID, "У вас пока нет записей.")
		return
	}

	var response strings.Builder
	for _, booking := range bookings {
		response.WriteString(fmt.Sprintf("ID: %d\nИмя: %s\nТелефон: %s\nДата/время: %s\n\n",
			booking.ID, booking.Name, booking.Contact, booking.Datetime))
	}
	b.sendMessage(chatID, response.String())
}

func (b *TgBot) sendMessage(chatID int64, text string) {
	msg := tgbot.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}
