CREATE TABLE IF NOT EXISTS pvz (
                                   id UUID PRIMARY KEY,
                                   registration_date TIMESTAMP NOT NULL DEFAULT NOW(),
    city VARCHAR(50) NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань'))
    );

CREATE TABLE IF NOT EXISTS reception (
                                         id UUID PRIMARY KEY,
                                         date_time TIMESTAMP NOT NULL DEFAULT NOW(),
    pvz_id UUID NOT NULL REFERENCES pvz(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('in_progress', 'close'))
    );

CREATE TABLE IF NOT EXISTS product (
                                       id UUID PRIMARY KEY,
                                       date_time TIMESTAMP NOT NULL DEFAULT NOW(),
    type VARCHAR(20) NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    reception_id UUID NOT NULL REFERENCES reception(id)
    );

CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY,
                                     email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('employee', 'moderator'))
    );

CREATE INDEX IF NOT EXISTS idx_reception_pvz_id ON reception(pvz_id);
CREATE INDEX IF NOT EXISTS idx_reception_status ON reception(status);
CREATE INDEX IF NOT EXISTS idx_product_reception_id ON product(reception_id);
CREATE INDEX IF NOT EXISTS idx_reception_date_time ON reception(date_time);