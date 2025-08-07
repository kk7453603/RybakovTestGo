-- Инициализация базы данных
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создание индексов для оптимизации
CREATE INDEX IF NOT EXISTS idx_currencies_symbol ON currencies(symbol);
CREATE INDEX IF NOT EXISTS idx_currency_prices_symbol ON currency_prices(symbol);
CREATE INDEX IF NOT EXISTS idx_currency_prices_timestamp ON currency_prices(timestamp);
CREATE INDEX IF NOT EXISTS idx_currency_prices_symbol_timestamp ON currency_prices(symbol, timestamp);

-- Создание функции для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';