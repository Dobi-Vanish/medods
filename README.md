# HTTP Server for providing JWT tokens (Go)

[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blue.svg)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-24.0+-blue.svg)](https://www.docker.com/)

Простой HTTP-сервер на Go для генерации, хранения и проверки JWT-токенов.

## 🚀 Функционал
- **JWT-авторизация** (Middleware для некоторых эндпоинтов)
- **API Endpoints**:
  - `GET /users/{id}/status` — информация о пользователе
  - `GET /users/leaderboard` — список пользователей
  - `GET /refresh/{id}` - обновление токенов
  - `GET /provide/{id}` - предоставление токенов
  - `POST /authenticate` - аутентификация пользователя
  - `POST /registrate` - регистрация пользователя
- **Хранилище**: PostgreSQL с миграциями (`goose`)
- **Docker-сборка**: Готовый `docker-compose.yml` для развертывания

## 📦 Установка
### Предварительные требования
- Go 1.21+
- PostgreSQL 15+
- Docker 24.0+

### Запуск локально
1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/Dobi-Vanish/medods
2. Перейдите в папку deployments и запустите через makefile:
   ```bash
   cd auth-service/deployments
   make up_build
### Пример успешного запроса
 Запуск коллекции в Postman для проверки:  
 ![изображение](https://github.com/user-attachments/assets/94ad1fcc-4806-4f57-adb0-1713358d33ea)  
 ### Примечание
 Для начала необходимо зарегестрировать нового пользователя, а затем аутентифицироваться за него, чтобы получить токены и было понятно, на какого пользователя сохранять токены в БД.  
 Также в задании было указано что "формат передачи base64", как я понял, это формат передачи токена пользователю, но по этой причине он содержит в себе IP пользователя. Прошу указать если надо как-либо исправить.

