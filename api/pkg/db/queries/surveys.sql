-- name: CreateSurvey :one
INSERT INTO surveys
(parse_status, delivery_status, error_log, name, config, url_slug)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSurvey :one
SELECT *
FROM surveys
WHERE uuid = $1 
LIMIT 1;

-- name: GetSurveys :many
SELECT *
FROM surveys;
