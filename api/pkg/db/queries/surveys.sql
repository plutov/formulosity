-- name: CreateSurvey :one
INSERT INTO surveys
(parse_status, delivery_status, error_log, name, config, url_slug)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSurveys :many
SELECT *
FROM surveys
ORDER BY id DESC;
