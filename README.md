# Stomatology Bot

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-28.4-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

Telegram-бот для записи на приём в стоматологическую клинику. Бот интегрирован с Google Calendar для управления расписанием.

## 🚀 Функционал

-   Запись на приём на 30 дней вперёд.
-   Выбор доступной даты и времени.
-   Запрос имени, фамилии и номера телефона.
-   Валидация номера телефона.
-   Просмотр своих записей.
-   Отмена записи.
-   Уведомление администратора о новых записях со ссылкой на событие в Google Calendar.
-   Разграничение доступа: клиенты не видят ссылки на события.

## 🛠️ Установка и запуск

Проект использует Docker и Docker Compose для простоты развертывания.

1.  **Клонируйте репозиторий:**
    ```bash
    git clone https://github.com/shenikar/stomatology_bot.git
    cd stomatology_bot
    ```

2.  **Настройте Google Calendar API:**
    Следуйте инструкциям в разделе [Настройка Google Calendar API](#-настройка-google-calendar-api), чтобы получить файл `credentials.json` и ID календаря. Поместите `credentials.json` в корень проекта.

3.  **Создайте и настройте файл `.env`:**
    Скопируйте `.env.example` в новый файл с именем `.env`:
    ```bash
    cp .env.example .env
    ```
    Затем откройте `.env` и заполните необходимые значения (секретные ключи и ID).

4.  **Запустите проект:**
    ```bash
    docker-compose up --build
    ```
    Бот будет запущен, а база данных PostgreSQL развёрнута в Docker-контейнере. Миграции применятся автоматически при старте.

---

## 🗓️ Настройка Google Calendar API

Для работы с Google Calendar бот использует **сервисный аккаунт**. Это специальный тип аккаунта Google, который принадлежит приложению, а не конкретному пользователю.

### Шаг 1: Создание проекта в Google Cloud Console

1.  Перейдите в [Google Cloud Console](https://console.cloud.google.com/).
2.  Создайте новый проект (или выберите существующий).

### Шаг 2: Включение Google Calendar API

1.  В меню навигации выберите **APIs & Services > Library**.
2.  Найдите **Google Calendar API** и включите (Enable) его для вашего проекта.

### Шаг 3: Создание сервисного аккаунта

1.  Перейдите в **APIs & Services > Credentials**.
2.  Нажмите **Create Credentials** и выберите **Service account**.
3.  Заполните имя сервисного аккаунта (например, `stomatology-bot-account`) и нажмите **Create and Continue**.
4.  На шаге **Grant this service account access to project** можно ничего не выбирать и нажать **Continue**.
5.  На шаге **Grant users access to this service account** тоже можно ничего не выбирать и нажать **Done**.

### Шаг 4: Получение файла `credentials.json`

1.  После создания сервисного аккаунта вы вернётесь на страницу **Credentials**.
2.  Найдите в списке ваш сервисный аккаунт и нажмите на него.
3.  Перейдите на вкладку **Keys**.
4.  Нажмите **Add Key > Create new key**.
5.  Выберите тип ключа **JSON** и нажмите **Create**.
6.  Файл с ключами будет автоматически скачан. **Переименуйте его в `credentials.json`** и поместите в корень проекта.

    > ⚠️ **Внимание!** Этот файл содержит приватные ключи. Никогда не добавляйте его в систему контроля версий (он уже добавлен в `.gitignore`).

### Шаг 5: Настройка доступа к календарю

1.  Откройте [Google Calendar](https://calendar.google.com/).
2.  Создайте новый календарь для записей (или используйте существующий).
3.  В настройках календаря (`Settings and sharing`) найдите раздел **Share with specific people or groups**.
4.  Нажмите **Add people and groups** и вставьте `client_email` из вашего `credentials.json` (это email вашего сервисного аккаунта).
5.  В выпадающем списке **Permissions** выберите **Make changes to events**.
6.  Скопируйте **Calendar ID** из раздела **Integrate calendar**. Он понадобится для переменной `CALENDAR_ID` в `.env`.

---

## 📂 Структура проекта

-   `cmd/bot/main.go`: Точка входа в приложение.
-   `configs/`: Конфигурация приложения.
-   `internal/`: Внутренняя логика проекта, не предназначенная для импорта извне.
    -   `booking/`: Логика, связанная с записями (модель, репозиторий).
    -   `logger/`: Настройка логгера.
    -   `platform/`: Взаимодействие с внешними сервисами.
        -   `calendar/`: Клиент для Google Calendar.
        -   `database/`: Подключение к БД.
        -   `telegram/`: Логика Telegram-бота.
-   `migrations/`: Миграции базы данных.
-   `Dockerfile`: Инструкции для сборки Docker-образа.
-   `docker-compose.yml`: Конфигурация для запуска проекта с помощью Docker Compose.

---

## 📜 Лицензия

Этот проект распространяется под лицензией MIT. Подробности смотрите в файле [LICENSE](LICENSE).
