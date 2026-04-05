-- name: UpsertUser :exec
-- Создаёт пользователя если не существует.
-- При повторном вызове ничего не меняет.
INSERT INTO users (user_id)
VALUES (@user_id::uuid)
ON CONFLICT (user_id) DO NOTHING;
