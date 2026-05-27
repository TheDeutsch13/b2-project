# GAMEGEAR — интернет-магазин игровой периферии

Монорепозиторий учебного проекта: React SPA + два Go-микросервиса, общая библиотека `b2-common`, PostgreSQL.

**Сайт:** каталог, корзина, оформление заказа, личный кабинет, админ-панель, чат поддержки, отзывы покупателей.

## Стек

| Слой | Технологии |
|------|------------|
| **Frontend** | React 19, TypeScript, Vite, Redux Toolkit, React Router, Axios, React Hook Form + Yup, CSS Modules, Lucide |
| **Backend** | Go, Gin, PostgreSQL, pgx, golang-migrate, zap, swaggo, WebSocket |
| **Shared** | `github.com/TheDeutsch13/b2-common` (logger, env, JWT, middleware, postgres) |

## Возможности

### Покупатель (`user`)
- Регистрация и вход (JWT + refresh в `localStorage`)
- Каталог с фильтрами и поиском, карточка товара (описание, характеристики, отзывы)
- Корзина **без входа**; оформление заказа — только после авторизации
- Checkout wizard (3 шага): контакты → доставка (СДЭК ПВЗ / свой адрес) → оплата
- Личный кабинет: заказы, избранное, отзывы, настройки профиля и аватар
- Отзыв на товар после статуса заказа «Получен»
- Чат поддержки (виджет) + push-уведомления по WebSocket

### Администратор (`admin`)
- CRUD товаров, загрузка фото, управление заказами и статусами
- Пользователи и назначение ролей
- Поддержка (все треды), модерация отзывов с фильтрами
- Панель курьера (доступна и админу)

### Модератор (`moderator`)
- Пользователи и чат поддержки (без товаров/заказов/отзывов)

### Курьер (`courier`)
- Список заказов и смена статусов доставки (`/courier`)

### UI/UX
- Адаптив: десктоп ≥1280px, мобильный **375px** (breakpoint `768px`)
- Фиолетовые кнопки соцсетей в футере (YouTube, VK, Telegram, Discord)

## Структура репозитория

```text
b2-project/
├── migrations/
│   ├── auth/                 # users, profiles, refresh_tokens, roles
│   └── product/              # categories, products, orders, support, reviews (JSONB)
├── services/
│   ├── docker-compose.yml    # PostgreSQL 16 + migrate
│   ├── auth-service/         # :8081 — auth, профили, upload аватаров
│   └── product-service/      # :8082 — каталог, заказы, отзывы, support, WS
├── scripts/
│   ├── migrate.ps1           # миграции в работающий контейнер
│   ├── grant-admin.ps1       # выдать роль admin по email
│   ├── seed-products.mjs     # демо-товары
│   └── sync-carousel.mjs     # слайды из папки Carousel/
├── Carousel/                 # исходники промо-карусели
└── frontend/                 # Vite dev :5173
```

## Быстрый старт

### Требования

- Docker Desktop
- Go 1.22+
- Node.js 20+
- Локально клонированный [b2-common](https://github.com/TheDeutsch13/b2-common) рядом с репозиторием (`../../../b2-common` в `go.mod`)

### 1. База данных и миграции

```powershell
cd services
docker compose up -d postgres
docker compose up migrate-auth migrate-product
```

Или из корня (если Postgres уже запущен):

```powershell
.\scripts\migrate.ps1
```

### 2. Backend

В **двух** терминалах, одинаковый `JWT_SECRET`:

```powershell
# Терминал 1 — auth-service
cd services/auth-service
$env:JWT_SECRET="gamegear-dev-secret"
go run ./cmd/api

# Терминал 2 — product-service
cd services/product-service
$env:JWT_SECRET="gamegear-dev-secret"
go run ./cmd/api
```

### 3. Frontend

```powershell
cd frontend
npm install
npm run dev
```

Откройте http://localhost:5173

Опционально — демо-товары:

```powershell
cd frontend
npm run seed:products
```

### 4. Первый администратор

```powershell
.\scripts\grant-admin.ps1 -Email "your@email.com"
```

Или SQL:

```sql
UPDATE users SET role = 'admin' WHERE email = 'your@email.com';
```

После смены роли в БД нужно **выйти и войти снова** (роль зашита в JWT).

## Роли

| Роль | Маршруты | Доступ |
|------|----------|--------|
| `user` | `/`, `/catalog`, `/cart`, `/checkout`, `/profile/*` | Покупки, отзывы, поддержка |
| `admin` | `/admin`, `/courier` | Полная админка |
| `moderator` | `/admin?tab=users`, `support` | Пользователи + поддержка |
| `courier` | `/courier` | Заказы на доставку |

Роль назначает администратор в админке (вкладка «Пользователи») или через SQL / `grant-admin.ps1`.

## Соответствие ТЗ (frontend)

| № | Требование | Реализация |
|---|------------|------------|
| 1 | Auth + JWT + маршруты | `/login`, `/register`, `ProtectedRoute`, Axios interceptors |
| 2 | Роли (≥2) | 4 роли, разный UI и маршруты |
| 3 | Глобальное состояние | Redux Toolkit: auth, cart, favorites, notifications, profile, support |
| 4 | Адаптив | 1280px / 375px, breakpoint 768px |
| 5 | Сложная форма | Checkout wizard, RHF + Yup |
| 6 | WebSocket | Уведомления о заказах + сообщения поддержки |

## API (кратко)

Прокси Vite: `/api/auth` → `:8081`, остальное `/api/*` → `:8082`.

### Auth Service — http://localhost:8081

| Method | Endpoint | Описание |
|--------|----------|----------|
| POST | `/api/auth/register` | Регистрация |
| POST | `/api/auth/login` | Вход |
| POST | `/api/auth/refresh` | Refresh token |
| GET | `/api/auth/me` | Текущий пользователь |
| GET/PATCH | `/api/auth/profile` | Профиль |
| POST | `/api/auth/upload` | Аватар |
| GET | `/api/auth/users` | Список пользователей (staff) |
| PATCH | `/api/auth/users/:id/role` | Смена роли (admin) |
| GET | `/api/auth/users/public` | Публичные профили по id |
| GET | `/swagger/index.html` | Swagger |

### Product Service — http://localhost:8082

| Method | Endpoint | Описание |
|--------|----------|----------|
| GET | `/api/products`, `/api/products/:id` | Каталог |
| GET | `/api/categories` | Категории |
| POST | `/api/orders` | Создать заказ |
| GET | `/api/orders/my` | Мои заказы |
| GET | `/api/orders` | Все заказы (admin) |
| PATCH | `/api/orders/:id/status` | Статус (admin) |
| GET/PATCH | `/api/courier/orders` | Панель курьера |
| PUT/DELETE | `/api/products/:id/reviews` | Отзыв покупателя |
| GET | `/api/reviews/my` | Мои отзывы |
| GET | `/api/reviews` | Все отзывы (admin, фильтры) |
| GET/POST | `/api/support/*` | Чат поддержки |
| GET | `/api/cdek/points` | Пункты СДЭК |
| POST | `/api/upload` | Фото товара (admin) |
| GET | `/ws/notifications?token=` | WebSocket |
| GET | `/swagger/index.html` | Swagger |

Параметры фильтра отзывов (admin): `rating`, `product_id`, `q`.

## Статусы заказа

`pending` → `confirmed` → `shipped` → `delivered` → `received` (можно оставить отзыв)  
Отмена: `cancelled` (остаток на складе восстанавливается)

## Скрипты

| Скрипт | Назначение |
|--------|------------|
| `scripts/migrate.ps1` | Применить миграции auth + product |
| `scripts/migrate-reset.ps1` | Сброс БД (осторожно) |
| `scripts/grant-admin.ps1` | Назначить admin по email |
| `scripts/seed-products.mjs` | Заполнить каталог демо-товарами |
| `scripts/sync-carousel.mjs` | Синхронизация слайдов из `Carousel/` |

## Переменные окружения

| Переменная | Сервис | По умолчанию |
|------------|--------|--------------|
| `PORT` | auth / product | `8081` / `8082` |
| `DATABASE_URL` | оба | `postgres://b2user:b2password@localhost:5432/b2db?sslmode=disable` |
| `JWT_SECRET` | оба | `gamegear-dev-secret-change-me` (должен совпадать!) |
| `UPLOAD_DIR` | оба | `./uploads` |

## Тесты

```powershell
cd services/auth-service
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out

cd ../product-service
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Демо-сценарий для защиты

1. Запуск: Postgres → миграции → auth + product + `npm run dev`
2. Регистрация → каталог → корзина без входа → вход → checkout (3 шага)
3. `grant-admin.ps1` → админка: товар, заказ, смена статуса → toast по WebSocket
4. Пользователь: заказ «Получен» → отзыв на товаре
5. Админ: вкладка «Отзывы» с фильтрами
6. Чат поддержки: пользователь + модератор в двух вкладках
7. Мобильная вёрстка 375px в DevTools
8. Swagger + покрытие тестами

## b2-common

Локальная разработка:

```go
replace github.com/TheDeutsch13/b2-common => ../../../b2-common
```

После изменений в `b2-common`: `go mod tidy` в обоих сервисах.

## Лицензия

Учебный проект.
