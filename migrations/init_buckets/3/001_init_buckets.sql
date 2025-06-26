-- +goose Up
CREATE SCHEMA bucket_5;
CREATE SCHEMA bucket_6;

-- +goose Down
DROP SCHEMA bucket_5;
DROP SCHEMA bucket_6;
