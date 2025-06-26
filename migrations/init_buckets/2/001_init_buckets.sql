-- +goose Up
CREATE SCHEMA bucket_3;
CREATE SCHEMA bucket_4;

-- +goose Down
DROP SCHEMA bucket_3;
DROP SCHEMA bucket_4;
