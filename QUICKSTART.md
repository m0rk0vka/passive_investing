# 🚀 Быстрый старт - Локальное тестирование

Пошаговая инструкция для запуска бота локально и проверки всего флоу.

## Шаг 1: Поднять БД

```bash
# Запустить Postgres через Podman
podman-compose up -d

# Или через Docker
docker-compose up -d

# Проверить, что БД запустилась
podman ps
# Должен быть контейнер с postgres:16
```

## Шаг 2: Применить миграции

```bash
# Подключиться к БД и применить первую миграцию
psql postgres://invest:invest@localhost:5432/invest < migrations/0001_init.sql

# Применить вторую миграцию (виртуальные портфели)
psql postgres://invest:invest@localhost:5432/invest < migrations/0002_virtual_portfolios.sql

# Проверить, что таблицы созданы
psql postgres://invest:invest@localhost:5432/invest -c "\dt"
```

Должны быть таблицы:
- `tg_user`, `tg_chat`
- `upload`, `raw_file`
- `account`, `import_batch`
- `valuation_snapshot`, `position_snapshot`
- `virtual_portfolio`, `virtual_portfolio_item`

## Шаг 3: Настроить .env

Убедись, что в `.env` файле есть:

```bash
BOT_TOKEN=your_telegram_bot_token_here
DATABASE_URL=postgres://invest:invest@localhost:5432/invest?sslmode=disable
RAW_DATA_DIR=./data
```

**Получить BOT_TOKEN:**
1. Открой [@BotFather](https://t.me/BotFather) в Telegram
2. Отправь `/newbot`
3. Следуй инструкциям
4. Скопируй токен в `.env`

## Шаг 4: Создать директорию для данных

```bash
mkdir -p ./data
```

## Шаг 5: Запустить бота

```bash
go run cmd/bot/main.go
```

Должен увидеть:
```
INFO    Database connected successfully
INFO    Parser worker started
```

## Шаг 6: Протестировать загрузку файла

1. Открой своего бота в Telegram
2. Отправь `/start` - должен ответить "Пришли мне XLSX отчёт документом"
3. Отправь XLSX файл отчета ВТБ
4. Бот должен ответить: "Файл сохранен. Имя: ... SHA256: ..."

**Что происходит под капотом:**
- Файл скачивается в `./data/`
- Создается запись в `upload` со статусом `RECEIVED`
- Сохраняется метаинформация в `raw_file`

## Шаг 7: Проверить парсинг (воркер)

Через 5-10 секунд воркер должен обработать файл. Проверь логи:

```
INFO    Upload processed successfully   {"upload_id": 1, "account_id": 1, "period": "2025-10"}
```

**Проверить в БД:**

```bash
# Проверить статус upload
psql postgres://invest:invest@localhost:5432/invest -c "SELECT id, status FROM upload;"

# Должен быть статус DONE

# Проверить созданный account
psql postgres://invest:invest@localhost:5432/invest -c "SELECT * FROM account;"

# Проверить снапшоты
psql postgres://invest:invest@localhost:5432/invest -c "SELECT * FROM valuation_snapshot;"
psql postgres://invest:invest@localhost:5432/invest -c "SELECT * FROM position_snapshot;"
```

## Шаг 8: Проверить UI

1. В Telegram отправь боту `/ui`
2. Должно открыться интерактивное меню
3. Нажми "📊 Портфели"
4. Должен увидеть свой портфель (название из отчета)
5. Открой портфель - увидишь общую сумму
6. Нажми "📋 Позиции" - увидишь список бумаг

## Шаг 9: Создать виртуальный портфель (вручную)

Если у тебя несколько аккаунтов, можно создать виртуальный портфель:

```sql
-- Подключиться к БД
psql postgres://invest:invest@localhost:5432/invest

-- Найти свой user_id
SELECT id, telegram_user_id FROM tg_user;

-- Найти account_id своих счетов
SELECT id, name FROM account;

-- Создать виртуальный портфель
INSERT INTO virtual_portfolio (id, user_id, name)
VALUES (1, YOUR_USER_ID, 'Мой общий портфель');

-- Добавить аккаунты в виртуальный портфель
INSERT INTO virtual_portfolio_item (id, virtual_portfolio_id, account_id)
VALUES 
  (1, 1, ACCOUNT_ID_1),
  (2, 1, ACCOUNT_ID_2);
```

Теперь в `/ui` → "📊 Портфели" увидишь:
- Реальные портфели (каждый счет отдельно)
- Виртуальный портфель (агрегация нескольких счетов)

## Шаг 10: Проверить агрегацию

Открой виртуальный портфель - увидишь:
- **Общую сумму** = сумма всех счетов
- **Позиции** = агрегированные по ISIN (если одна бумага в нескольких счетах, количество суммируется)

## Troubleshooting

### Бот не запускается

```bash
# Проверить, что БД доступна
psql postgres://invest:invest@localhost:5432/invest -c "SELECT 1;"

# Проверить переменные окружения
cat .env
```

### Файл не парсится

```bash
# Проверить логи воркера
# Должны быть ошибки парсинга

# Проверить статус в БД
psql postgres://invest:invest@localhost:5432/invest -c "SELECT id, status, error_message FROM upload;"
```

### UI не показывает данные

```bash
# Проверить, что снапшоты созданы
psql postgres://invest:invest@localhost:5432/invest -c "SELECT COUNT(*) FROM valuation_snapshot;"
psql postgres://invest:invest@localhost:5432/invest -c "SELECT COUNT(*) FROM position_snapshot;"

# Проверить логи бота при открытии UI
```

## Что дальше?

✅ Основной флоу работает!

Теперь можно:
1. Загружать новые отчеты - они автоматически парсятся
2. Просматривать историю по месяцам
3. Создавать виртуальные портфели для агрегации
4. Смотреть проценты распределения активов

### Планы на будущее

- [ ] UI для создания виртуальных портфелей
- [ ] Расчет доходности портфеля
- [ ] План покупок для ребалансировки
- [ ] Графики и визуализация
- [ ] Экспорт данных

## Полезные команды

```bash
# Остановить БД
podman-compose down

# Очистить БД (удалить все данные)
podman-compose down -v

# Пересоздать БД с нуля
podman-compose down -v
podman-compose up -d
psql postgres://invest:invest@localhost:5432/invest < migrations/0001_init.sql
psql postgres://invest:invest@localhost:5432/invest < migrations/0002_virtual_portfolios.sql

# Посмотреть логи БД
podman logs -f financer-postgres-1

# Подключиться к БД интерактивно
psql postgres://invest:invest@localhost:5432/invest