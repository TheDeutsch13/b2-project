# GAMEGEAR — интернет-магазин игровой периферии

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
| `migrate.ps1` | Миграции auth + product в Docker Postgres |
| `migrate-reset.ps1` | Полный сброс БД (только dev) |
| `migrate-fix-auth.ps1` | Починка «таблица уже существует» для auth |
| `grant-admin.ps1` | Назначить `admin` по email |
| `seed-products.mjs` | Демо-товары через API (`npm run seed:products`) |
| `sync-carousel.mjs` | Слайды карусели из папки `Carousel/` |
| `sync-carousel.ps1` | Обёртка для `sync-carousel.mjs` |

**Быстро:**

```powershell
.\scripts\migrate.ps1
.\scripts\grant-admin.ps1 -Email "you@mail.com"
cd frontend
$env:SEED_EMAIL="you@mail.com"; $env:SEED_PASSWORD="пароль"; npm run seed:products
```

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