-- +goose Up
CREATE TABLE IF NOT EXISTS medods(
    id serial PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    password VARCHAR(255) NOT NULL,
    active INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE UNIQUE INDEX idx_medods_email ON medods(email);
    CREATE INDEX idx_medods_created_at ON medods(created_at);
    CREATE INDEX idx_medods_inactive ON medods(email) WHERE active = 0;
    CREATE INDEX idx_medods_name ON medods(first_name, last_name);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS medods;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
