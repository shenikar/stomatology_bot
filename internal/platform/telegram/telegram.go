package telegram

import (
	"fmt"
	"log"
	"stomatology_bot/configs"
	"stomatology_bot/internal/booking"
	"stomatology_bot/internal/platform/calendar"
	"strconv"
	"strings"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const slotDuration = time.Hour

// Состояния пользователя
const (
	StateDefault         = ""
	StateAwaitingDate    = "awaiting_date"
	StateAwaitingTime    = "awaiting_time"
	StateAwaitingName    = "awaiting_name"
	StateAwaitingContact = "awaiting_contact"
)

type UserState struct {
	State     string
	TempTime  time.Time // Для хранения выбранной даты/времени
	TempName  string    // Для хранения имени
	TempEvent string    // Для хранения ID события календаря
}

type TgBot struct {
	api         *tgbot.BotAPI
	cfg         *configs.Config
	repo        *booking.BookingRepo
	calendarSvc *calendar.CalendarService
	userStates  map[int64]*UserState
}

func NewBot(api *tgbot.BotAPI, cfg *configs.Config, repo *booking.BookingRepo, calendarSvc *calendar.CalendarService) *TgBot {
	return &TgBot{
		api:         api,
		cfg:         cfg,
		repo:        repo,
		calendarSvc: calendarSvc,
		userStates:  make(map[int64]*UserState),
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
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	state, exists := b.userStates[chatID]

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start", "help":
			b.userStates[chatID] = &UserState{State: StateDefault}
			b.sendMainMenu(chatID)
		default:
			b.sendMessage(chatID, "Неизвестная команда. Используйте /help для получения списка доступных команд.")
		}
		return
	}

	if exists {
		switch state.State {
		case StateAwaitingName:
			b.handleNameInput(update)
			return
		case StateAwaitingContact:
			b.handleContactInput(update)
			return
		}
	}

	b.sendMessage(chatID, "Пожалуйста, используйте кнопки или команды.")
}

func (b *TgBot) handleCallbackQuery(update tgbot.Update) {
	if update.CallbackQuery == nil {
		return
	}

	// Отправляем "typing" статус
	callback := tgbot.NewCallback(update.CallbackQuery.ID, "")
	b.api.Request(callback)

	chatID := update.CallbackQuery.Message.Chat.ID

	// Разбор данных колбэка
	data := update.CallbackQuery.Data
	switch {
	case data == "book":
		b.handleBookCommand(chatID)
	case data == "my_bookings":
		b.handleShowAllBooking(update) // Передаем весь update
	case strings.HasPrefix(data, "date_"):
		b.handleDateSelection(update)
	case strings.HasPrefix(data, "time_"):
		b.handleTimeSelection(update)
	case strings.HasPrefix(data, "cancel_"):
		b.handleCancelBooking(update)
	default:
		b.sendMessage(chatID, "Неизвестное действие.")
	}
}

func (b *TgBot) handleBookCommand(chatID int64) {
	b.userStates[chatID] = &UserState{State: StateAwaitingDate}
	// Предлагаем выбрать дату
	var buttons [][]tgbot.InlineKeyboardButton
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, i)
		// Пропускаем воскресенье
		if date.Weekday() == time.Sunday {
			continue
		}
		button := tgbot.NewInlineKeyboardButtonData(date.Format("02.01.2006"), "date_"+date.Format("2006-01-02"))
		buttons = append(buttons, []tgbot.InlineKeyboardButton{button})
	}

	keyboard := tgbot.NewInlineKeyboardMarkup(buttons...)
	msg := tgbot.NewMessage(chatID, "Выберите дату для записи:")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *TgBot) handleDateSelection(update tgbot.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	b.userStates[chatID] = &UserState{State: StateAwaitingTime}

	dateStr := strings.TrimPrefix(update.CallbackQuery.Data, "date_")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		b.sendMessage(chatID, "Неверный формат даты.")
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
	chatID := update.CallbackQuery.Message.Chat.ID
	timeStr := strings.TrimPrefix(update.CallbackQuery.Data, "time_")
	slot, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		b.sendMessage(chatID, "Неверный формат времени.")
		return
	}

	// Сохраняем выбранное время и переходим к запросу имени
	b.userStates[chatID] = &UserState{
		State:    StateAwaitingName,
		TempTime: slot,
	}

	b.sendMessage(chatID, "Пожалуйста, введите ваше Имя и Фамилию.")
}

func (b *TgBot) handleNameInput(update tgbot.Update) {
	chatID := update.Message.Chat.ID
	name := update.Message.Text

	state, ok := b.userStates[chatID]
	if !ok || state.State != StateAwaitingName {
		b.sendMessage(chatID, "Произошла ошибка состояния. Пожалуйста, начните заново с /start.")
		return
	}

	// Сохраняем имя и переходим к запросу контакта
	state.State = StateAwaitingContact
	state.TempName = name
	b.userStates[chatID] = state

	b.sendMessage(chatID, "Спасибо! Теперь, пожалуйста, введите ваш номер телефона для связи.")
}

func (b *TgBot) handleContactInput(update tgbot.Update) {
	chatID := update.Message.Chat.ID
	contact := update.Message.Text

	// Валидация номера телефона
	if !strings.HasPrefix(contact, "+7") || len(contact) != 12 {
		b.sendMessage(chatID, "Неверный формат номера. Пожалуйста, введите номер в формате +7XXXXXXXXXX (12 цифр).")
		return // Оставляем пользователя в том же состоянии, чтобы он мог повторить ввод
	}

	state, ok := b.userStates[chatID]
	if !ok || state.State != StateAwaitingContact {
		b.sendMessage(chatID, "Произошла ошибка состояния. Пожалуйста, начните заново с /start.")
		return
	}

	userName := state.TempName // Используем сохраненное имя
	slot := state.TempTime

	// Повторная проверка, свободен ли слот
	isFree, err := b.calendarSvc.IsSlotFree(slot, slot.Add(slotDuration))
	if err != nil {
		log.Printf("Ошибка проверки доступности слота: %v", err)
		b.sendMessage(chatID, "Произошла ошибка. Попробуйте снова.")
		return
	}
	if !isFree {
		b.sendMessage(chatID, "К сожалению, этот слот только что заняли. Пожалуйста, выберите другое время.")
		b.userStates[chatID] = &UserState{State: StateDefault} // Сброс состояния
		return
	}

	// Создаем событие в Google Calendar
	summary := fmt.Sprintf("Запись: %s", userName)
	description := fmt.Sprintf("Запись на прием от пользователя %s.\nКонтакт: %s", userName, contact)
	link, eventID, err := b.calendarSvc.CreateEvent(summary, description, slot, slot.Add(slotDuration))
	if err != nil {
		log.Printf("Ошибка создания события в календаре: %v", err)
		b.sendMessage(chatID, "Не удалось создать запись. Попробуйте позже.")
		return
	}

	// Создаем запись в нашей БД
	booking := &booking.Booking{
		UserID:   chatID,
		Name:     userName,
		Contact:  contact,
		Datetime: slot,
		EventID:  &eventID,
	}

	if err := b.repo.CreateBooking(booking); err != nil {
		log.Printf("Ошибка сохранения записи в БД: %v. Откатываем событие в календаре.", err)
		if delErr := b.calendarSvc.DeleteEvent(eventID); delErr != nil {
			log.Printf("КРИТИЧЕСКАЯ ОШИБКА: не удалось откатить событие %s в календаре: %v", eventID, delErr)
			b.sendMessage(chatID, "Произошла критическая ошибка. Пожалуйста, свяжитесь с администратором.")
		} else {
			log.Printf("Событие %s успешно удалено из календаря.", eventID)
			b.sendMessage(chatID, "Ошибка при сохранении записи в базу данных. Попробуйте снова.")
		}
	} else {
		response := fmt.Sprintf("Вы успешно записаны на %s.\nСсылка на событие: %s", slot.Format("02.01.2006 в 15:04"), link)
		b.sendMessage(chatID, response)
	}

	// Сбрасываем состояние пользователя
	delete(b.userStates, chatID)
}

func (b *TgBot) handleShowAllBooking(update tgbot.Update) {
	var chatID int64
	if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
	} else {
		return // Не можем определить чат
	}

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
		loc, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			log.Printf("Ошибка загрузки часового пояса: %v", err)
			// В случае ошибки выводим как есть (в UTC)
			response.WriteString(fmt.Sprintf("ID: %d\nИмя: %s\nТелефон: %s\nДата/время: %s\n\n",
				booking.ID, booking.Name, booking.Contact, booking.Datetime))
			continue
		}
		localTime := booking.Datetime.In(loc)
		response.WriteString(fmt.Sprintf("ID: %d\nИмя: %s\nТелефон: %s\nДата/время: %s\n\n",
			booking.ID, booking.Name, booking.Contact, localTime.Format("02.01.2006 15:04")))

		// Добавляем кнопку отмены для каждой записи
		keyboard := tgbot.NewInlineKeyboardMarkup(
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData("Отменить запись", fmt.Sprintf("cancel_%d", booking.ID)),
			),
		)
		msg := tgbot.NewMessage(chatID, response.String())
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
		response.Reset() // Очищаем builder для следующей записи
	}
}

func (b *TgBot) handleCancelBooking(update tgbot.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	bookingIDStr := strings.TrimPrefix(update.CallbackQuery.Data, "cancel_")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		log.Printf("Ошибка конвертации ID записи: %v", err)
		b.sendMessage(chatID, "Некорректный ID записи.")
		return
	}

	// 1. Получаем запись из БД, чтобы узнать event_id
	bookingToCancel, err := b.repo.GetBookingByID(bookingID)
	if err != nil {
		log.Printf("Ошибка получения записи для отмены: %v", err)
		b.sendMessage(chatID, "Не удалось найти указанную запись.")
		return
	}

	// 2. Удаляем событие из Google Calendar (если оно есть)
	if bookingToCancel.EventID != nil {
		if err := b.calendarSvc.DeleteEvent(*bookingToCancel.EventID); err != nil {
			log.Printf("Ошибка удаления события из календаря: %v", err)
			b.sendMessage(chatID, "Ошибка при отмене записи в календаре. Пожалуйста, попробуйте еще раз.")
			// Не продолжаем, если не удалось удалить из календаря
			return
		}
	}

	// 3. Удаляем запись из нашей БД
	if err := b.repo.DeleteBookingById(bookingID); err != nil {
		log.Printf("КРИТИЧЕСКАЯ ОШИБКА: событие в календаре удалено, но не удалось удалить запись из БД: %v", err)
		b.sendMessage(chatID, "Произошла критическая ошибка при отмене. Пожалуйста, свяжитесь с администратором.")
		return
	}

	b.sendMessage(chatID, "Ваша запись успешно отменена.")
}

func (b *TgBot) sendMessage(chatID int64, text string) {
	msg := tgbot.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func (b *TgBot) sendMainMenu(chatID int64) {
	msg := tgbot.NewMessage(chatID, "Добро пожаловать! Выберите действие:")
	keyboard := tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData("Записаться на приём", "book"),
			tgbot.NewInlineKeyboardButtonData("Мои записи", "my_bookings"),
		),
	)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}
