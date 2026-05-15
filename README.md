# Purchases API

Backend-сервис для управления покупками: пользователи, товары, заказы и история покупок.

Идея проекта — создать backend, который позволяет вести список покупок и работать с ним с разных устройств.

---

## 🎯 Цель проекта

- управление пользователями и ролями
- управление товарами и категориями
- создание заказов
- хранение истории покупок
- подготовка базы под статистику

---

## 🧱 Модель данных

### Пользователи
- id
- name
- password (хешируется)
- email
- role (user/admin)
- status (active/blocked)
- created_at
- updated_at

### Категории
- id
- name

### Товары
- id
- title
- unit
- category_id

### Алиасы товаров
- id
- product_id
- alias

### Магазины
- id
- name

### Заказы
- id
- user_id
- store_id
- created_at

### Позиции заказа
- id
- order_id
- product_id
- count

### История покупок
- id
- user_id
- external_id
- store_id
- total_sum
- purchased_at
- raw_qr

### Позиции покупки
- id
- purchase_id
- product_id
- raw_name
- count
- price

---

## 🧠 Архитектура

Проект разделён на слои:

- HTTP handlers (обработка запросов)
- Service layer (бизнес-логика)
- Storage layer (работа с БД)
- Middleware (логирование, авторизация)

---

## 🔐 Авторизация

- JWT токены
- роли:
  - user
  - admin
- доступ через middleware
- userID и role передаются через context

---

## ⚙️ Возможности

### Реализовано / в работе

- регистрация пользователей
- CRUD пользователей
- CRUD товаров
- CRUD категорий
- CRUD магазинов
- создание заказов
- управление позициями заказа
- история покупок

---

## 🧩 Архитектурные решения

- разделение public/private API
- context-based авторизация
- транзакции в service layer
- обработка ошибок доменного уровня
- логирование через middleware

---

## 🚧 Планы развития

- статистика покупок
- фильтры и пагинация
- Redis кеширование
- расширение policy слоя (права доступа)
- event-based история покупок

---

## 🛠 Технологии

- Go
- PostgreSQL
- chi router
- JWT
- bcrypt
- slog (логирование)
