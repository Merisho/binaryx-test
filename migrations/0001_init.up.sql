BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE,
    password TEXT,
    first_name TEXT,
    last_name TEXT
);

CREATE TABLE IF NOT EXISTS user_wallets (
    user_id UUID,
    wallet TEXT UNIQUE,
    currency VARCHAR(16),
    PRIMARY KEY (user_id, wallet)
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    currency VARCHAR(16),
    to_wallet TEXT,
    from_wallet TEXT,
    amount TEXT,
    fee TEXT,
    timestamp TIMESTAMP DEFAULT NOW()
);
CREATE INDEX transactions_to_index ON transactions (to_wallet);
CREATE INDEX transactions_from_index ON transactions (from_wallet);

COMMIT;
