package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/plutov/formulosity/api/pkg/db"
	"github.com/plutov/formulosity/api/pkg/types"
)

type Postgres struct {
	conn    *pgx.Conn
	queries *db.Queries
	addr    string
	ctx     context.Context
}

func (p *Postgres) Init() error {
	p.ctx = context.Background()

	p.addr = os.Getenv("DATABASE_URL")
	if len(p.addr) == 0 {
		return errors.New("DATABASE_URL env var is empty")
	}

	var err error
	p.conn, err = pgx.Connect(context.Background(), p.addr)
	if err != nil {
		log.Fatalf("cannot connect to postgres: %v", err)
	}

	p.queries = db.New(p.conn)

	if err = p.Ping(); err != nil {
		return err
	}

	return p.Migrate()
}

func (p *Postgres) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.conn.Ping(ctx)
}

func (p *Postgres) Close() error {
	return p.conn.Close(context.Background())
}

func (p *Postgres) Migrate() error {
	migrationsDir := "file://migrations"

	m, err := migrate.New(migrationsDir, p.addr)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return nil
}

func (p *Postgres) CreateSurvey(survey *types.Survey) error {
	configBytes, err := json.Marshal(survey.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal survey config: %w", err)
	}

	surveyDb, err := p.queries.CreateSurvey(context.Background(), db.CreateSurveyParams{
		ParseStatus: db.NullSurveyParseStatuses{
			Valid:               true,
			SurveyParseStatuses: db.SurveyParseStatuses(survey.ParseStatus),
		},
		DeliveryStatus: db.NullSurveyDeliveryStatuses{
			Valid:                  true,
			SurveyDeliveryStatuses: db.SurveyDeliveryStatuses(survey.DeliveryStatus),
		},
		ErrorLog: pgtype.Text{
			Valid:  true,
			String: survey.ErrorLog,
		},
		Name:    survey.Name,
		Config:  configBytes,
		UrlSlug: survey.URLSlug,
	})
	if err != nil {
		return fmt.Errorf("failed to create survey: %w", err)
	}

	survey.ID = int64(surveyDb.ID)
	survey.UUID = db.EncodeUUID(surveyDb.Uuid)
	return nil
}

func (p *Postgres) UpdateSurvey(survey *types.Survey) error {
	query := `UPDATE surveys
		SET parse_status=$1, delivery_status=$2, error_log=$3, name=$4, config=$5, url_slug=$6
		WHERE uuid=$7;`

	_, err := p.conn.Exec(p.ctx, query, survey.ParseStatus, survey.DeliveryStatus, survey.ErrorLog, survey.Name, survey.Config, survey.URLSlug, survey.UUID)
	return err
}

func (p *Postgres) GetSurveys() ([]*types.Survey, error) {
	query := `SELECT
		s.id, s.uuid, s.created_at,
		s.parse_status, s.delivery_status,
		s.error_log, s.name, s.config, s.url_slug,
		(SELECT COUNT(*) FROM surveys_sessions WHERE survey_id = s.id AND status = $1) AS sessions_count_in_progress,
		(SELECT COUNT(*) FROM surveys_sessions WHERE survey_id = s.id AND status = $2) AS sessions_count_completed
	FROM surveys AS s;`

	rows, err := p.conn.Query(p.ctx, query, types.SurveySessionStatus_InProgress, types.SurveySessionStatus_Completed)
	if err != nil {
		return nil, err
	}

	surveys := []*types.Survey{}
	for rows.Next() {
		survey := &types.Survey{}

		err := rows.Scan(&survey.ID, &survey.UUID, &survey.CreatedAt,
			&survey.ParseStatus, &survey.DeliveryStatus, &survey.ErrorLog,
			&survey.Name, &survey.Config, &survey.URLSlug,
			&survey.Stats.SessionsCountInProgess, &survey.Stats.SessionsCountCompleted)
		if err != nil {
			return nil, err
		}

		totalResponses := survey.Stats.SessionsCountInProgess + survey.Stats.SessionsCountCompleted
		if totalResponses > 0 {
			survey.Stats.CompletionRate = (survey.Stats.SessionsCountCompleted * 100) / totalResponses
		}

		surveys = append(surveys, survey)
	}

	return surveys, nil
}

func (p *Postgres) GetSurveyByField(field string, value interface{}) (*types.Survey, error) {
	query := fmt.Sprintf(`SELECT
		s.id, s.uuid, s.created_at,
		s.parse_status, s.delivery_status,
		s.error_log, s.name, s.config, s.url_slug
	FROM surveys AS s
	WHERE s.%s=$1;`, field)

	row := p.conn.QueryRow(p.ctx, query, value)
	survey := &types.Survey{}
	err := row.Scan(&survey.ID, &survey.UUID, &survey.CreatedAt,
		&survey.ParseStatus, &survey.DeliveryStatus, &survey.ErrorLog,
		&survey.Name, &survey.Config, &survey.URLSlug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return survey, nil
}

func (p *Postgres) UpsertSurveyQuestions(survey *types.Survey) error {
	if survey == nil || survey.Config == nil || survey.Config.Questions == nil {
		return nil
	}

	placeholders := []string{}
	values := []interface{}{}
	values = append(values, survey.ID)
	for i := range survey.Config.Questions.Questions {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		values = append(values, survey.Config.Questions.Questions[i].ID)
	}
	// Delete removed questions
	deleteQuery := `DELETE FROM surveys_questions
	WHERE survey_id=$1
	AND question_id NOT IN (` + strings.Join(placeholders, ", ") + `);`

	_, err := p.conn.Exec(p.ctx, deleteQuery, values...)
	if err != nil {
		return err
	}

	for _, q := range survey.Config.Questions.Questions {
		insertQuery := `INSERT INTO surveys_questions
		(survey_id, question_id)
		VALUES ($1, $2)
		ON CONFLICT (survey_id, question_id)
		DO NOTHING
		;`

		_, err := p.conn.Exec(p.ctx, insertQuery, survey.ID, q.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Postgres) GetSurveyQuestions(surveyID int64) ([]types.Question, error) {
	query := `SELECT
		uuid, question_id
	FROM surveys_questions
	WHERE survey_id=$1;`

	rows, err := p.conn.Query(p.ctx, query, surveyID)
	if err != nil {
		return []types.Question{}, err
	}

	questions := []types.Question{}
	for rows.Next() {
		question := types.Question{}

		err := rows.Scan(&question.UUID, &question.ID)
		if err != nil {
			return []types.Question{}, err
		}

		questions = append(questions, question)
	}

	return questions, nil
}

func (p *Postgres) CreateSurveySession(session *types.SurveySession) error {
	query := `INSERT INTO surveys_sessions
		(status, survey_id, ip_addr)
		VALUES ($1, (SELECT id FROM surveys WHERE uuid = $2), $3)
		RETURNING id, uuid;`

	row := p.conn.QueryRow(p.ctx, query, session.Status, session.SurveyUUID, session.IPAddr)
	return row.Scan(&session.ID, &session.UUID)
}

func (p *Postgres) UpdateSurveySessionStatus(sessionUUID string, newStatus types.SurveySessionStatus) error {
	completedAt := "NULL"
	if newStatus == types.SurveySessionStatus_Completed {
		completedAt = "NOW()"
	}

	query := fmt.Sprintf(`UPDATE surveys_sessions
		SET status = $1, completed_at = %s
		WHERE uuid = $2;`, completedAt)

	_, err := p.conn.Exec(p.ctx, query, newStatus, sessionUUID)

	return err
}

func (p *Postgres) GetSurveySession(surveyUUID string, sessionUUID string) (*types.SurveySession, error) {
	query := `SELECT
		ss.id, ss.uuid, ss.created_at, ss.status, s.uuid
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE ss.uuid=$1 AND s.uuid=$2;`

	row := p.conn.QueryRow(p.ctx, query, sessionUUID, surveyUUID)
	session := &types.SurveySession{}
	err := row.Scan(&session.ID, &session.UUID, &session.CreatedAt, &session.Status, &session.SurveyUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return session, nil
}

func (p *Postgres) DeleteSurveySession(sessionUUID string) error {
	query := `DELETE
	FROM surveys_sessions
	WHERE uuid=$1;`

	_, err := p.conn.Exec(p.ctx, query, sessionUUID)
	return err
}

func (p *Postgres) GetSurveySessionByIPAddress(surveyUUID string, ipAddr string) (*types.SurveySession, error) {
	query := `SELECT
		ss.id, ss.uuid, ss.created_at, ss.status, s.uuid
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE s.uuid=$1 AND ss.ip_addr=$2;`

	row := p.conn.QueryRow(p.ctx, query, surveyUUID, ipAddr)
	session := &types.SurveySession{}
	err := row.Scan(&session.ID, &session.UUID, &session.CreatedAt, &session.Status, &session.SurveyUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return session, nil
}

func (p *Postgres) GetSurveySessionAnswers(sessionUUID string) ([]types.QuestionAnswer, error) {
	query := `SELECT
		q.id, q.uuid, sa.answer
	FROM surveys_answers AS sa
	LEFT JOIN surveys_questions AS q ON q.id = sa.question_id
	WHERE session_id = (SELECT id FROM surveys_sessions WHERE uuid = $1);`

	rows, err := p.conn.Query(p.ctx, query, sessionUUID)
	if err != nil {
		return nil, err
	}

	answers := []types.QuestionAnswer{}
	for rows.Next() {
		answer := types.QuestionAnswer{}

		err := rows.Scan(&answer.QuestionID, &answer.QuestionUUID, &answer.AnswerBytes)
		if err != nil {
			return nil, err
		}

		answers = append(answers, answer)
	}

	return answers, nil
}

func (p *Postgres) UpsertSurveyQuestionAnswer(sessionUUID string, questionUUID string, answer types.Answer) error {
	query := `INSERT INTO surveys_answers
		(session_id, question_id, answer)
		VALUES (
			(SELECT id FROM surveys_sessions WHERE uuid = $1),
			(SELECT id FROM surveys_questions WHERE uuid = $2),
			$3
		)
		ON CONFLICT (session_id, question_id)
		DO UPDATE SET answer = EXCLUDED.answer;`

	_, err := p.conn.Exec(p.ctx, query, sessionUUID, questionUUID, answer)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetSurveySessionsWithAnswers(surveyUUID string, filter *types.SurveySessionsFilter) ([]types.SurveySession, int, error) {
	query := fmt.Sprintf(`WITH limited_sessions AS (
		SELECT * from surveys_sessions
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	)
	SELECT
		ss.id, ss.uuid, ss.created_at, ss.completed_at, ss.status, q.id, q.uuid, sa.answer, w.response_status, w.response
	FROM limited_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	LEFT JOIN surveys_answers AS sa ON sa.session_id = ss.id
	LEFT JOIN surveys_questions AS q ON q.id = sa.question_id
	LEFT JOIN surveys_webhook_responses AS w ON w.session_id = ss.id
	WHERE s.uuid=$1
	ORDER BY ss.%s %s
	;`, filter.SortBy, filter.Order, filter.Limit, filter.Offset, filter.SortBy, filter.Order)

	rows, err := p.conn.Query(p.ctx, query, surveyUUID)
	if err != nil {
		return nil, 0, err
	}

	sessions := []types.SurveySession{}
	sessionsMap := map[string]types.SurveySession{}
	for rows.Next() {
		session := types.SurveySession{}
		webhookData := types.WebhookData{}
		answer := types.QuestionAnswer{}
		var (
			questionID   sql.NullString
			questionUUID sql.NullString
		)

		err := rows.Scan(&session.ID, &session.UUID, &session.CreatedAt, &session.CompletedAt, &session.Status, &questionID, &questionUUID, &answer.AnswerBytes, &webhookData.StatusCode, &webhookData.Response)
		if err != nil {
			return nil, 0, err
		}

		session.WebhookData = webhookData

		if _, ok := sessionsMap[session.UUID]; !ok {
			session.QuestionAnswers = []types.QuestionAnswer{}
			sessionsMap[session.UUID] = session
			sessions = append(sessions, session)
		}

		if questionID.Valid && questionUUID.Valid {
			answer.QuestionID = questionID.String
			answer.QuestionUUID = questionUUID.String

			sessionCopy := sessionsMap[session.UUID]
			sessionCopy.QuestionAnswers = append(sessionCopy.QuestionAnswers, answer)
			sessionsMap[session.UUID] = sessionCopy
		}
	}

	totalCount, err := p.getSurveySessionsCount(surveyUUID)
	if err != nil {
		return nil, 0, err
	}

	for i, session := range sessions {
		fullSession := sessionsMap[session.UUID]
		sessions[i].QuestionAnswers = fullSession.QuestionAnswers
	}

	return sessions, totalCount, nil
}

func (p *Postgres) getSurveySessionsCount(surveyUUID string) (int, error) {
	query := `SELECT
		COUNT(*)
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE s.uuid=$1;`

	row := p.conn.QueryRow(p.ctx, query, surveyUUID)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (p *Postgres) StoreWebhookResponse(sessionId int, responseStatus int, response string) error {
	query := `INSERT INTO surveys_webhook_responses
		(created_at, session_id, response_status, response)
		VALUES ($1, $2, $3, $4);`

	createdAtStr := time.Now().UTC().Format(types.DateTimeFormat)
	_, err := p.conn.Exec(p.ctx, query, createdAtStr, sessionId, responseStatus, response)
	return err
}
