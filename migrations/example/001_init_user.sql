-- +goose Up
CREATE TABLE "user" (
    user_id uuid,
    name text
);

-- +goose Down
DROP TABLE "user";
