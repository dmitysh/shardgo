-- +goose Up
CREATE SCHEMA bucket_1;
CREATE SCHEMA bucket_2;

-- +goose Down
DROP SCHEMA bucket_1;
DROP SCHEMA bucket_2;
