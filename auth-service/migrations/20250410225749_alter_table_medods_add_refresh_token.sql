-- +goose Up
ALTER TABLE medods
ADD COLUMN refresh_token TEXT,
ADD COLUMN refresh_token_expires TIMESTAMP;

CREATE INDEX idx_medods_refresh_token ON medods(refresh_token);
CREATE INDEX idx_medods_refresh_token_expires ON medods(refresh_token_expires);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
ALTER TABLE medods
DROP COLUMN refresh_token,
DROP COLUMN refresh_token_expires;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd