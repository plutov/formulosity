package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/plutov/formulosity/api/pkg/types"
)

type Sqlite struct {
	conn *sql.DB
	addr string
}

func (p *Sqlite) Init() error {
	p.addr = os.Getenv("DATABASE_URL")
	if len(p.addr) == 0 {
		return errors.New("DATABASE_URL env var is empty")
	}

	if _, err := os.Stat(p.addr); err != nil {
		if _, err := os.Create(p.addr); err != nil {
			return fmt.Errorf("unable to create db file: %w", err)
		}
	}

	var err error
	p.conn, err = sql.Open("sqlite3", p.addr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}

	if err = p.Ping(); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}

	return p.Migrate()
}

func (p *Sqlite) Ping() error {
	return p.conn.Ping()
}

func (p *Sqlite) Close() error {
	return p.conn.Close()
}

func (p *Sqlite) Migrate() error {
	migrationsDir := "file://migrations/sqlite"

	driver, err := migratepg.WithInstance(p.conn, &migratepg.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("error creating migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsDir, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	err = m.Up()
	if err != nil {
		if err != migrate.ErrNoChange {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}
	return nil
}

func (p *Sqlite) CreateSurvey(survey *types.Survey) error {
	query := `INSERT INTO surveys
		(parse_status, delivery_status, error_log, name, config, url_slug, uuid, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`

	survey.UUID = uuid.New().String()
	configBytes, _ := json.Marshal(survey.Config)
	createdAtStr := time.Now().UTC().Format(types.DateTimeFormat)

	_, err := p.conn.Exec(query, survey.ParseStatus, survey.DeliveryStatus, survey.ErrorLog, survey.Name, string(configBytes), survey.URLSlug, survey.UUID, createdAtStr)
	return err
}

func (p *Sqlite) UpdateSurvey(survey *types.Survey) error {
	query := `UPDATE surveys
		SET parse_status=$1, delivery_status=$2, error_log=$3, name=$4, config=$5, url_slug=$6
		WHERE uuid=$7;`

	configBytes, _ := json.Marshal(survey.Config)
	_, err := p.conn.Exec(query, survey.ParseStatus, survey.DeliveryStatus, survey.ErrorLog, survey.Name, string(configBytes), survey.URLSlug, survey.UUID)
	return err
}

func (p *Sqlite) GetSurveys() ([]*types.Survey, error) {
	query := `SELECT
		s.id, s.uuid, s.created_at,
		s.parse_status, s.delivery_status,
		s.error_log, s.name, s.config, s.url_slug,
		(SELECT COUNT(*) FROM surveys_sessions WHERE survey_id = s.id AND status = $1) AS sessions_count_in_progress,
		(SELECT COUNT(*) FROM surveys_sessions WHERE survey_id = s.id AND status = $2) AS sessions_count_completed
	FROM surveys AS s;`

	rows, err := p.conn.Query(query, types.SurveySessionStatus_InProgress, types.SurveySessionStatus_Completed)
	if err != nil {
		return nil, err
	}

	surveys := []*types.Survey{}
	for rows.Next() {
		survey := &types.Survey{}
		var (
			createdAtStr sql.NullString
			configStr    sql.NullString
		)

		err := rows.Scan(&survey.ID, &survey.UUID, &createdAtStr,
			&survey.ParseStatus, &survey.DeliveryStatus, &survey.ErrorLog,
			&survey.Name, &configStr, &survey.URLSlug,
			&survey.Stats.SessionsCountInProgess, &survey.Stats.SessionsCountCompleted)
		if err != nil {
			return nil, err
		}

		survey.CreatedAt, _ = time.Parse(types.DateTimeFormat, createdAtStr.String)
		json.Unmarshal([]byte(configStr.String), &survey.Config)

		totalResponses := survey.Stats.SessionsCountInProgess + survey.Stats.SessionsCountCompleted
		if totalResponses > 0 {
			survey.Stats.CompletionRate = (survey.Stats.SessionsCountCompleted * 100) / totalResponses
		}

		surveys = append(surveys, survey)
	}

	return surveys, nil
}

func (p *Sqlite) GetSurveyByField(field string, value interface{}) (*types.Survey, error) {
	query := fmt.Sprintf(`SELECT
		s.id, s.uuid, s.created_at,
		s.parse_status, s.delivery_status,
		s.error_log, s.name, s.config, s.url_slug
	FROM surveys AS s
	WHERE s.%s=$1;`, field)

	var (
		createdAtStr sql.NullString
		configStr    sql.NullString
	)

	row := p.conn.QueryRow(query, value)
	survey := &types.Survey{}
	err := row.Scan(&survey.ID, &survey.UUID, &createdAtStr,
		&survey.ParseStatus, &survey.DeliveryStatus, &survey.ErrorLog,
		&survey.Name, &configStr, &survey.URLSlug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	survey.CreatedAt, _ = time.Parse(types.DateTimeFormat, createdAtStr.String)
	json.Unmarshal([]byte(configStr.String), &survey.Config)

	return survey, nil
}

func (p *Sqlite) UpsertSurveyQuestions(survey *types.Survey) error {
	if survey == nil || survey.Config == nil || survey.Config.Questions == nil {
		return nil
	}

	placeholders := []string{}
	values := []interface{}{}
	values = append(values, survey.UUID)
	for i := range survey.Config.Questions.Questions {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		values = append(values, survey.Config.Questions.Questions[i].ID)
	}
	// Delete removed questions
	deleteQuery := `DELETE FROM surveys_questions
	WHERE survey_id = (SELECT id FROM surveys where uuid = $1)
	AND question_id NOT IN (` + strings.Join(placeholders, ", ") + `);`

	_, err := p.conn.Exec(deleteQuery, values...)
	if err != nil {
		return err
	}

	for _, q := range survey.Config.Questions.Questions {
		insertQuery := `INSERT INTO surveys_questions
		(survey_id, question_id, uuid)
		VALUES ((SELECT id FROM surveys where uuid = $1), $2, $3)
		ON CONFLICT (survey_id, question_id)
		DO NOTHING
		;`

		uuidStr := uuid.New().String()
		_, err := p.conn.Exec(insertQuery, survey.UUID, q.ID, uuidStr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Sqlite) GetSurveyQuestions(surveyID int64) ([]types.Question, error) {
	query := `SELECT
		uuid, question_id
	FROM surveys_questions
	WHERE survey_id=$1;`

	rows, err := p.conn.Query(query, surveyID)
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

func (p *Sqlite) CreateSurveySession(session *types.SurveySession) error {
	query := `INSERT INTO surveys_sessions
		(status, survey_id, ip_addr, uuid, created_at)
		VALUES ($1, (SELECT id FROM surveys WHERE uuid = $2), $3, $4, $5);`

	session.UUID = uuid.New().String()
	createdAtStr := time.Now().UTC().Format(types.DateTimeFormat)
	_, err := p.conn.Exec(query, session.Status, session.SurveyUUID, session.IPAddr, session.UUID, createdAtStr)

	return err
}

func (p *Sqlite) UpdateSurveySessionStatus(sessionUUID string, newStatus types.SurveySessionStatus) error {
	completedAt := ""
	if newStatus == types.SurveySessionStatus_Completed {
		completedAt = time.Now().UTC().Format(types.DateTimeFormat)
	}

	query := `UPDATE surveys_sessions
		SET status = $1, completed_at = $2
		WHERE uuid = $3;`

	_, err := p.conn.Exec(query, newStatus, completedAt, sessionUUID)

	return err
}

func (p *Sqlite) GetSurveySession(surveyUUID string, sessionUUID string) (*types.SurveySession, error) {
	query := `SELECT
		ss.id, ss.uuid, ss.created_at, ss.status, s.uuid
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE ss.uuid=$1 AND s.uuid=$2;`

	row := p.conn.QueryRow(query, sessionUUID, surveyUUID)
	session := &types.SurveySession{}
	var createdAtStr sql.NullString
	err := row.Scan(&session.ID, &session.UUID, &createdAtStr, &session.Status, &session.SurveyUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	session.CreatedAt, _ = time.Parse(types.DateTimeFormat, createdAtStr.String)

	return session, nil
}

func (p *Sqlite) GetSurveySessionByIPAddress(surveyUUID string, ipAddr string) (*types.SurveySession, error) {
	query := `SELECT
		ss.id, ss.uuid, ss.created_at, ss.status, s.uuid
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE s.uuid=$1 AND ss.ip_addr=$2;`

	row := p.conn.QueryRow(query, surveyUUID, ipAddr)
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

func (p *Sqlite) GetSurveySessionAnswers(sessionUUID string) ([]types.QuestionAnswer, error) {
	query := `SELECT
		q.id, q.uuid, sa.answer
	FROM surveys_answers AS sa
	LEFT JOIN surveys_questions AS q ON q.id = sa.question_id
	WHERE session_id = (SELECT id FROM surveys_sessions WHERE uuid = $1);`

	rows, err := p.conn.Query(query, sessionUUID)
	if err != nil {
		return nil, err
	}

	answers := []types.QuestionAnswer{}
	for rows.Next() {
		answer := types.QuestionAnswer{}
		var (
			questionID   sql.NullString
			questionUUID sql.NullString
			answerStr    sql.NullString
		)
		err := rows.Scan(&questionID, &questionUUID, &answerStr)
		if err != nil {
			return nil, err
		}

		if !questionID.Valid || !questionUUID.Valid {
			continue
		}

		answer.QuestionID = questionID.String
		answer.QuestionUUID = questionUUID.String
		answer.AnswerBytes = []byte(answerStr.String)
		answers = append(answers, answer)
	}

	return answers, nil
}

func (p *Sqlite) UpsertSurveyQuestionAnswer(sessionUUID string, questionUUID string, answer types.Answer) error {
	query := `INSERT INTO surveys_answers
		(session_id, question_id, answer, uuid, created_at)
		VALUES (
			(SELECT id FROM surveys_sessions WHERE uuid = $1),
			(SELECT id FROM surveys_questions WHERE uuid = $2),
			$3, $4, $5
		)
		ON CONFLICT (session_id, question_id)
		DO UPDATE SET answer = EXCLUDED.answer;`

	answerBytes, _ := json.Marshal(answer)
	uuidStr := uuid.New().String()
	createdAtStr := time.Now().UTC().Format(types.DateTimeFormat)

	_, err := p.conn.Exec(query, sessionUUID, questionUUID, string(answerBytes), uuidStr, createdAtStr)
	if err != nil {
		return err
	}

	return nil
}

func (p *Sqlite) GetSurveySessionsWithAnswers(surveyUUID string, filter *types.SurveySessionsFilter) ([]types.SurveySession, int, error) {
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

	rows, err := p.conn.Query(query, surveyUUID)
	if err != nil {
		return nil, 0, err
	}

	sessions := []types.SurveySession{}
	sessionsMap := map[string]types.SurveySession{}
	for rows.Next() {
		session := types.SurveySession{}
		answer := types.QuestionAnswer{}
		var (
			questionID     sql.NullString
			questionUUID   sql.NullString
			createdAtStr   sql.NullString
			completedAtStr sql.NullString
			answerStr      sql.NullString
			httpStatusCode sql.NullInt16
			httpResponse   sql.NullString
		)

		err := rows.Scan(&session.ID, &session.UUID, &createdAtStr, &completedAtStr, &session.Status, &questionID, &questionUUID, &answerStr, &httpStatusCode, &httpResponse)
		if err != nil {
			return nil, 0, err
		}

		session.CreatedAt, _ = time.Parse(types.DateTimeFormat, createdAtStr.String)
		completedAt, completedAtErr := time.Parse(types.DateTimeFormat, completedAtStr.String)
		if completedAtErr == nil {
			session.CompletedAt = &completedAt
		}
		answer.AnswerBytes = []byte(answerStr.String)
		// fmt.Println(httpStatusCode)
		// fmt.Println(httpResponse)

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

func (p *Sqlite) getSurveySessionsCount(surveyUUID string) (int, error) {
	query := `SELECT
		COUNT(*)
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE s.uuid=$1;`

	row := p.conn.QueryRow(query, surveyUUID)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (p *Sqlite) StoreWebhookResponse(sessionId int, responseStatus int, response string) error {
	query := `INSERT INTO surveys_webhook_responses
		(created_at, session_id, response_status, response)
		VALUES ($1, $2, $3, $4);`

	createdAtStr := time.Now().UTC().Format(types.DateTimeFormat)

	_, err := p.conn.Exec(query, createdAtStr, sessionId, responseStatus, response)
	return err
}
