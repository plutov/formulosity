-- name: CreateSurvey :one
INSERT INTO surveys (parse_status, delivery_status, error_log, name, config, url_slug)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: UpdateSurvey :exec
UPDATE
    surveys
SET
    parse_status = $1,
    delivery_status = $2,
    error_log = $3,
    name = $4,
    config = $5,
    url_slug = $6
WHERE
    uuid = $7;

-- name: GetSurveys :many
SELECT
    s.id,
    s.uuid,
    s.created_at,
    s.parse_status,
    s.delivery_status,
    s.error_log,
    s.name,
    s.config,
    s.url_slug,
    (
        SELECT
            COUNT(*)
        FROM
            surveys_sessions ss
        WHERE
            ss.survey_id = s.id
            AND ss.status = $1) AS sessions_count_in_progress,
    (
        SELECT
            COUNT(*)
        FROM
            surveys_sessions ss
        WHERE
            ss.survey_id = s.id
            AND ss.status = $2) AS sessions_count_completed
FROM
    surveys AS s
ORDER BY
    s.created_at DESC;

-- name: GetSurveyByUUID :one
SELECT
    s.id,
    s.uuid,
    s.created_at,
    s.parse_status,
    s.delivery_status,
    s.error_log,
    s.name,
    s.config,
    s.url_slug
FROM
    surveys AS s
WHERE
    s.uuid = $1;

-- name: GetSurveyByURLSlug :one
SELECT
    s.id,
    s.uuid,
    s.created_at,
    s.parse_status,
    s.delivery_status,
    s.error_log,
    s.name,
    s.config,
    s.url_slug
FROM
    surveys AS s
WHERE
    s.url_slug = $1;

-- name: DeleteSurveyQuestionsNotInList :exec
DELETE FROM surveys_questions
WHERE survey_id = $1
    AND question_id != ALL ($2::text[]);

-- name: UpsertSurveyQuestion :exec
INSERT INTO surveys_questions (survey_id, question_id)
    VALUES ($1, $2)
ON CONFLICT (survey_id, question_id)
    DO NOTHING;

-- name: GetSurveyQuestions :many
SELECT
    sq.uuid,
    sq.question_id
FROM
    surveys_questions sq
WHERE
    sq.survey_id = $1
ORDER BY
    sq.question_id;

-- name: CreateSurveySession :one
INSERT INTO surveys_sessions (status, survey_id, ip_addr)
    VALUES ($1, (
            SELECT
                s.id
            FROM
                surveys s
            WHERE
                s.uuid = $2), $3)
RETURNING
    id,
    uuid;

-- name: UpdateSurveySessionStatusCompleted :exec
UPDATE
    surveys_sessions
SET
    status = $1,
    completed_at = NOW()
WHERE
    uuid = $2;

-- name: UpdateSurveySessionStatus :exec
UPDATE
    surveys_sessions
SET
    status = $1
WHERE
    uuid = $2;

-- name: GetSurveySession :one
SELECT
    ss.id,
    ss.uuid,
    ss.created_at,
    ss.status,
    s.uuid AS survey_uuid
FROM
    surveys_sessions AS ss
    INNER JOIN surveys AS s ON s.id = ss.survey_id
WHERE
    ss.uuid = $1
    AND s.uuid = $2;

-- name: DeleteSurveySession :exec
DELETE FROM surveys_sessions
WHERE uuid = $1;

-- name: GetSurveySessionByIPAddress :one
SELECT
    ss.id,
    ss.uuid,
    ss.created_at,
    ss.status,
    s.uuid AS survey_uuid
FROM
    surveys_sessions AS ss
    INNER JOIN surveys AS s ON s.id = ss.survey_id
WHERE
    s.uuid = $1
    AND ss.ip_addr = $2;

-- name: GetSurveySessionAnswers :many
SELECT
    q.question_id,
    q.uuid AS question_uuid,
    sa.answer
FROM
    surveys_answers AS sa
    LEFT JOIN surveys_questions AS q ON q.id = sa.question_id
WHERE
    sa.session_id = (
        SELECT
            ss.id
        FROM
            surveys_sessions ss
        WHERE
            ss.uuid = $1)
ORDER BY
    q.question_id;

-- name: UpsertSurveyQuestionAnswer :exec
INSERT INTO surveys_answers (session_id, question_id, answer)
    VALUES ((
            SELECT
                ss.id
            FROM
                surveys_sessions ss
            WHERE
                ss.uuid = $1), (
                SELECT
                    sq.id
                FROM
                    surveys_questions sq
                WHERE
                    sq.uuid = $2), $3)
    ON CONFLICT (session_id,
        question_id)
    DO UPDATE SET
        answer = EXCLUDED.answer;

-- name: GetSurveySessionsWithAnswers :many
WITH limited_sessions AS (
    SELECT
        ss.*
    FROM
        surveys_sessions ss
    WHERE
        ss.survey_id = (
            SELECT
                s.id
            FROM
                surveys s
            WHERE
                s.uuid = $1)
        ORDER BY
            ss.created_at DESC
        LIMIT $2 OFFSET $3
)
SELECT
    ss.id,
    ss.uuid,
    ss.created_at,
    ss.completed_at,
    ss.status,
    q.question_id,
    q.uuid AS question_uuid,
    sa.answer,
    w.response_status,
    w.response
FROM
    limited_sessions AS ss
    LEFT JOIN surveys_answers AS sa ON sa.session_id = ss.id
    LEFT JOIN surveys_questions AS q ON q.id = sa.question_id
    LEFT JOIN surveys_webhook_responses AS w ON w.session_id = ss.id
ORDER BY
    ss.created_at DESC;

-- name: GetSurveySessionsCount :one
SELECT
    COUNT(*)
FROM
    surveys_sessions AS ss
    INNER JOIN surveys AS s ON s.id = ss.survey_id
WHERE
    s.uuid = $1;

-- name: StoreWebhookResponse :exec
INSERT INTO surveys_webhook_responses (created_at, session_id, response_status, response)
    VALUES ($1, $2, $3, $4);

