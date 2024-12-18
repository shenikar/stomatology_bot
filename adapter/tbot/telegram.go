package tbot

import (
	"fmt"
	"log"
	"stomatology_bot/config"
	"stomatology_bot/domain"
	"stomatology_bot/repository"
	"strings"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserState struct {
	Step string
	Date string
	Time string
}

type TgBot struct {
	api      *tgbot.BotAPI
	cfg      *config.Config
	repo     *repository.BookingRepo
	userStep map[int64]*UserState // Храним состояние пользователя
}

func NewBot(api *tgbot.BotAPI, cfg *config.Config, repo *repository.BookingRepo) *TgBot {
	return &TgBot{
		api:      api,
		cfg:      cfg,
		repo:     repo,
		userStep: make(map[int64]*UserState),
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
		msg := tgbot.NewMessage(update.Message.Chat.ID, "Список доступных команд:\n/start - Начать работу с ботом\n/help - Помощь\n/book - Записаться на приём\n/mybookings - Вывод всех записей клиента.\n/cancel - Отмена записи.")
		b.api.Send(msg)
	case "/book":
		b.sendCalendar(update.Message.Chat.ID) // Отправляем календарь для выбора даты
	case "/mybookings":
		b.handleShowAllBooking(update)
	case "/cancel":
		b.handleCancel(update)
	default:
		b.sendMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /help для получения списка доступных команд.")
	}
}

func (b *TgBot) handleCallbackQuery(update tgbot.Update) {
	if update.CallbackQuery != nil {
		chatID := update.CallbackQuery.Message.Chat.ID
		state, exists := b.userStep[chatID]
		if !exists {
			return
		}

		// Обрабатываем выбор даты
		if state.Step == "selecting_date" {
			b.handleDateSelection(update)
		} else if state.Step == "selecting_time" {
			b.handleTimeSelection(update)
		}
	}
}

func (b *TgBot) sendCalendar(chatID int64) {
	// Генерация списка дат для выбора (например, на ближайшие 7 дней)
	var buttons [][]tgbot.InlineKeyboardButton
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, i)
		button := tgbot.NewInlineKeyboardButtonData(date.Format("2006-01-02"), date.Format("2006-01-02"))
		buttons = append(buttons, []tgbot.InlineKeyboardButton{button})
	}

	// Создаем клавиатуру с кнопками
	keyboard := tgbot.NewInlineKeyboardMarkup(buttons...)
	msg := tgbot.NewMessage(chatID, "Выберите дату для записи:")
	msg.ReplyMarkup = keyboard

	// Отправляем сообщение с календарем
	b.api.Send(msg)

	// Устанавливаем состояние пользователя
	b.userStep[chatID] = &UserState{
		Step: "selecting_date",
	}
}

func (b *TgBot) handleDateSelection(update tgbot.Update) {
	date := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

	// Сохраняем выбранную дату
	state := b.userStep[chatID]
	state.Date = date

	// Генерация кнопок для выбора времени
	var buttons [][]tgbot.InlineKeyboardButton
	for i := 9; i <= 17; i++ { // Рабочие часы с 9:00 до 17:00
		timeSlot := fmt.Sprintf("%02d:00", i)
		button := tgbot.NewInlineKeyboardButtonData(timeSlot, timeSlot)
		buttons = append(buttons, []tgbot.InlineKeyboardButton{button})
	}

	// Создаем клавиатуру с кнопками для времени
	keyboard := tgbot.NewInlineKeyboardMarkup(buttons...)
	msg := tgbot.NewMessage(chatID, "Выберите время для записи:")
	msg.ReplyMarkup = keyboard

	// Отправляем сообщение с предложением выбрать время
	b.api.Send(msg)

	// Обновляем шаг
	state.Step = "selecting_time"
}

func (b *TgBot) handleTimeSelection(update tgbot.Update) {
	timeSelected := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

	// Получаем дату из состояния
	state := b.userStep[chatID]
	date := state.Date

	// Сохраняем время
	state.Time = timeSelected

	// Составляем полную дату и время
	datetime := fmt.Sprintf("%s %s", date, timeSelected)
	parsedDatetime, err := time.Parse("2006-01-02 15:04", datetime)
	if err != nil {
		b.sendMessage(chatID, "Ошибка формата времени. Попробуйте снова.")
		return
	}

	// Создаем запись в базе данных
	booking := &domain.Booking{
		Name:     "Имя пользователя", // Получите имя пользователя, если необходимо
		Contact:  "Контакт",          // Получите контакт пользователя, если необходимо
		Datetime: parsedDatetime,
	}

	if err := b.repo.CreateBooking(booking); err != nil {
		b.sendMessage(chatID, "Ошибка при создании записи. Попробуйте снова.")
	} else {
		b.sendMessage(chatID, fmt.Sprintf("Запись успешно создана! Номер записи: %d", booking.ID))
	}

	// Сбрасываем состояние пользователя
	delete(b.userStep, chatID)
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

func (b *TgBot) handleCancel(update tgbot.Update) {
	chatID := update.Message.Chat.ID
	b.sendMessage(chatID, "Введите ID записи, которую хотите отменить.")
	b.waitForInput(chatID, "cancel", func(input string) {
		var bookingID int
		if _, err := fmt.Sscanf(input, "%d", &bookingID); err != nil {
			b.sendMessage(chatID, "Ошибка ввода. Убедитесь, что указали корректный ID.")
			return
		}
		if err := b.repo.DeleteBookingById(bookingID); err != nil {
			b.sendMessage(chatID, "Ошибка при удалении записи. Проверьте корректность ID.")
		} else {
			b.sendMessage(chatID, "Запись успешно отменена.")
		}
	})
}

func (b *TgBot) sendMessage(chatID int64, text string) {
	msg := tgbot.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func (b *TgBot) waitForInput(chatID int64, step string, handler func(input string)) {
	// Устанавливаем состояние для пользователя
	b.userStep[chatID] = &UserState{
		Step: step,
	}

	// Получаем обновления только через один канал
	u := tgbot.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			if update.Message != nil && update.Message.Chat.ID == chatID {
				// Проверяем состояние пользователя
				if currentStep, ok := b.userStep[chatID]; ok && currentStep.Step == step {
					// Вызываем обработчик для ввода
					handler(update.Message.Text)
					// Сбрасываем состояние
					delete(b.userStep, chatID)
					return
				}
			}
		}
	}()
}
