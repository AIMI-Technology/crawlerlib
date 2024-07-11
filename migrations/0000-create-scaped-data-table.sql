-- File: create_scraped_data_table.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE scraped_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    text TEXT NOT NULL,
    scraped_at TIMESTAMP NOT NULL,
    published_at TIMESTAMP,
    url VARCHAR NOT NULL UNIQUE,
    source_country CHAR(2),
    content_country CHAR(2)
);

CREATE INDEX idx_scraped_data_url ON scraped_data (url);
