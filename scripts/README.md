# Скрипты GAMEGEAR

Все команды запускаются **из корня репозитория** `b2-project`, если не указано иное.

Перед большинством скриптов для БД должен быть запущен PostgreSQL:

```powershell
cd services
docker compose up -d postgres
```

---

## Сводная таблица

| Скрипт | Когда использовать |
|--------|-------------------|
| `migrate.ps1` | Первый запуск, после `git pull` с новыми миграциями |
| `migrate-reset.ps1` | Полностью очистить БД (только разработка) |
| `migrate-fix-auth.ps1` | Ошибка миграций auth: «таблица уже существует» |
| `grant-admin.ps1` | Сделать пользователя администратором |
| `seed-products.mjs` | Заполнить каталог демо-товарами |
| `sync-carousel.mjs` | Обновить слайды карусели на главной |
| `sync-carousel.ps1` | То же, обёртка для PowerShell |

---

## migrate.ps1 — миграции БД

Применяет SQL-миграции **auth** и **product** в Docker-контейнер `b2-postgres` через образ `migrate/migrate`.

### Использование

```powershell
.\scripts\migrate.ps1
```

Откат на одну версию назад (редко нужен):

```powershell
.\scripts\migrate.ps1 -Direction down
```

### Что делает

1. Проверяет, что контейнер `b2-postgres` запущен.
2. Показывает текущую версию миграций для auth и product.
3. Выполняет `up` для `migrations/auth` и `migrations/product`.

### Когда запускать

- Первый раз после клонирования проекта.
- После обновления репозитория, если появились новые файлы в `migrations/`.
- После `migrate-reset.ps1`.

### Альтернатива (через docker compose)

```powershell
cd services
docker compose up migrate-auth migrate-product
```

### Ошибки

| Сообщение | Решение |
|-----------|---------|
| `Container b2-postgres is not running` | `cd services` → `docker compose up -d postgres` |
| `Migration failed` | Смотрите текст ошибки; для auth см. `migrate-fix-auth.ps1` |

---

## migrate-reset.ps1 — сброс всей БД

**Удаляет все таблицы и данные** в базе `b2db`. Только для локальной разработки.

### Использование

```powershell
.\scripts\migrate-reset.ps1
```

Скрипт спросит подтверждение `yes`.

Без вопроса (осторожно):

```powershell
.\scripts\migrate-reset.ps1 -Force
```

### После сброса

```powershell
.\scripts\migrate.ps1
```

Затем заново: регистрация, `grant-admin.ps1`, при необходимости `seed-products.mjs`.

---

## migrate-fix-auth.ps1 — починка миграций auth

Если `migrate.ps1` падает с ошибками вроде **«relation users already exists»** или **«role already exists»** — таблицы уже есть, но migrate «не знает», что миграции применены.

### Использование

```powershell
.\scripts\migrate-fix-auth.ps1
.\scripts\migrate.ps1
```

Скрипт помечает auth-миграции как применённые до версии **3**. Если у вас миграций больше (например `000004`, `000005`), после `migrate.ps1` должны примениться только новые.

---

## grant-admin.ps1 — роль администратора

Назначает роль `admin` пользователю по email.

### Использование

```powershell
.\scripts\grant-admin.ps1 -Email "your@email.com"
```

### Требования

- Пользователь уже **зарегистрирован** на сайте.
- PostgreSQL доступен на `localhost:5432` (скрипт также умеет выполнить SQL через docker, используя `host.docker.internal`).
- Опционально: установлен `psql` — тогда SQL выполнится автоматически.

### После скрипта

**Обязательно выйти из аккаунта и войти снова** — роль хранится в JWT, старый токен останется с `user`.

### Вручную (SQL)

```sql
UPDATE users SET role = 'admin' WHERE email = 'your@email.com';
```

---

## seed-products.mjs — демо-товары в каталог

Создаёт тестовые товары через API (нужен запущенный **auth-service** и **product-service** и аккаунт **admin**).

### Использование

```powershell
cd frontend
$env:SEED_EMAIL="admin@test.com"
$env:SEED_PASSWORD="ваш_пароль"
npm run seed:products
```

Или напрямую:

```powershell
$env:SEED_EMAIL="admin@test.com"
$env:SEED_PASSWORD="ваш_пароль"
node scripts/seed-products.mjs
```

### Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `SEED_EMAIL` | — | Email admin (обязательно) |
| `SEED_PASSWORD` | — | Пароль (обязательно) |
| `AUTH_URL` | `http://localhost:8081` | Auth service |
| `PRODUCT_URL` | `http://localhost:8082` | Product service |
| `SEED_PER_CATEGORY` | `8` | Сколько товаров на категорию |

### Пример: меньше товаров

```powershell
$env:SEED_PER_CATEGORY="3"
node scripts/seed-products.mjs
```

### Что создаётся

По умолчанию **8 товаров** в каждой категории: Мыши, Коврики, Клавиатуры, Аксессуары — с брендами, характеристиками и ценами.

### Ошибки

| Сообщение | Решение |
|-----------|---------|
| `Задайте SEED_EMAIL и SEED_PASSWORD` | Указать переменные |
| `401` / login failed | Неверный пароль или пользователь не admin |
| `Категории не найдены` | Запустить product-service и `migrate.ps1` |

---

## sync-carousel.mjs — карусель на главной

Копирует слайды из папки `Carousel/` в `frontend/public/carousel/` и генерирует `frontend/src/data/carousel.generated.ts`.

### Структура папки Carousel

```text
Carousel/
├── 1.png          # картинка слайда (или .jpg, .webp, .gif)
├── 1.txt          # текст и ссылка (опционально)
├── 2.png
├── 2.txt
└── ...
```

Имя файла — **номер слайда**: `1`, `2`, `3`…

### Формат 1.txt

```text
Заголовок слайда
Подзаголовок или описание (можно несколько строк)
product: 5
```

или ссылка:

```text
Новая мышь
link: /catalog
```

или на конкретный товар:

```text
Скидка
/product/12
```

Поддерживаемые строки ссылки:

- `product: 12` или `товар: 12`
- `link: /catalog` или `link: https://...`
- `/product/12`

Первая строка без спецформата — **заголовок**, остальные — **описание**.

### Использование

```powershell
node scripts/sync-carousel.mjs
```

Ссылка по умолчанию для слайдов без `link` / `product`:

```powershell
node scripts/sync-carousel.mjs /catalog
```

### Автоматически

Скрипт вызывается при:

- `npm run dev`
- `npm run build`

через Vite-плагин и `package.json` → `sync:carousel`.

При изменении файлов в `Carousel/` в dev-режиме страница перезагружается.

### Обёртка PowerShell

```powershell
.\scripts\sync-carousel.ps1
.\scripts\sync-carousel.ps1 -Link "/catalog"
```

---

## Типичные сценарии

### Первый запуск на новом ПК

```powershell
cd services
docker compose up -d postgres
cd ..
.\scripts\migrate.ps1

# auth + product + frontend (3 терминала)
.\scripts\grant-admin.ps1 -Email "you@mail.com"
# войти на сайте, выйти, войти снова

cd frontend
$env:SEED_EMAIL="you@mail.com"
$env:SEED_PASSWORD="пароль"
npm run seed:products
```

### Обновили проект с GitHub

```powershell
git pull
.\scripts\migrate.ps1
# перезапустить auth-service и product-service
```

### «Всё сломалось», чистый старт

```powershell
.\scripts\migrate-reset.ps1
.\scripts\migrate.ps1
# заново регистрация, grant-admin, seed
```

### Поменяли баннеры на главной

1. Положить `N.png` и `N.txt` в `Carousel/`.
2. `node scripts/sync-carousel.mjs` или просто перезапустить `npm run dev`.

---

## npm-команды во frontend (связанные со скриптами)

| Команда | Скрипт |
|---------|--------|
| `npm run sync:carousel` | `sync-carousel.mjs` |
| `npm run seed:products` | `seed-products.mjs` |
| `npm run dev` | sync карусели + Vite |
| `npm run build` | sync карусели + production build |
