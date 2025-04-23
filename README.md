# HTTP Server for User Management (Go)

[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blue.svg)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-24.0+-blue.svg)](https://www.docker.com/)

Простой HTTP-сервер на Go для управления пользователями и их активностями (реферальные коды, задания, бонусные баллы).

## 🚀 Функционал
- **JWT-авторизация** (Middleware для некоторых эндпоинтов)
- **API Endpoints**:
  - `GET /users/{id}/status` — информация о пользователе
  - `GET /users/leaderboard` — топ пользователей по балансу
  - `POST /users/{id}/task/complete` — выполнение задания (награда в баллах)
  - `POST /users/{id}/referrer` — ввод реферального кода
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
   git clone https://github.com/Dobi-Vanish/CorrectingProject
2. Перейдите в папку deployments и запустите через makefile:
   ```bash
   cd reward-service/deployments
   make up_build
### Пример успешного запроса
 Запуск коллекции в Postman для проверки:  
 ![изображение](https://github.com/user-attachments/assets/6e12e9d7-f245-47a7-8035-56c0fb5df0cf)  
