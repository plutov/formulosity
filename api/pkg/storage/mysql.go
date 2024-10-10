package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migrateMysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/plutov/formulosity/api/pkg/types"
)

type Mysql struct {
	conn *sql.DB
	addr string
}

func (m *Mysql) Init() error {
	m.addr = os.Getenv("DATABASE_URL")
	if len(m.addr) == 0 {
		return errors.New("DATABASE_URL env var is empty")
	}

	var err error
	m.conn, err = sql.Open("mysql", m.addr)
	if err != nil {
		return err
	}

	if err = m.Ping(); err != nil {
		return err
	}

	return m.Migrate()
}

func (m *Mysql) Ping() error {
	return m.conn.Ping()
}

func (m *Mysql) Close() error {
	return m.conn.Close()
}

func (m *Mysql) Migrate() error {
	migrationsDir := "file://migrations/mysql"

	driver, err := migrateMysql.WithInstance(m.conn, &migrateMysql.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("error creating migration driver: %w", err)
	}

	mig, err := migrate.NewWithDatabaseInstance(migrationsDir, "mysql", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	err = mig.Up()
	if err != nil {
		if err != migrate.ErrNoChange {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}
	return nil
}

func (m *Mysql) CreateSurvey(survey *types.Survey) error {
	insert := `INSERT INTO surveys
		(parse_status, delivery_status, error_log, name, config, url_slug)
		VALUES (?, ?, ?, ?, ?, ?);`

	_, err := m.conn.Exec(insert, survey.ParseStatus, survey.DeliveryStatus, survey.ErrorLog, survey.Name, survey.Config, survey.URLSlug)
	if err != nil {
		return fmt.Errorf("unable to create survey - %w", err)
	}

	row := m.conn.QueryRow(`SELECT id FROM surveys WHERE id = LAST_INSERT_ID();`)
	return row.Scan(&survey.ID)
}

func (m *Mysql) UpdateSurvey(survey *types.Survey) error {
	query := `UPDATE surveys
		SET parse_status=?, delivery_status=?, error_log=?, name=?, config=?, url_slug=?
		WHERE uuid=?;`

	_, err := m.conn.Exec(query, survey.ParseStatus, survey.DeliveryStatus, survey.ErrorLog, survey.Name, survey.Config, survey.URLSlug, survey.UUID)
	return err
}

func (m *Mysql) GetSurveys() ([]*types.Survey, error) {
	query := `SELECT
		s.id, s.uuid, s.created_at,
		s.parse_status, s.delivery_status,
		s.error_log, s.name, s.config, s.url_slug,
		(SELECT COUNT(*) FROM surveys_sessions WHERE survey_id = s.id AND status = ?) AS sessions_count_in_progress,
		(SELECT COUNT(*) FROM surveys_sessions WHERE survey_id = s.id AND status = ?) AS sessions_count_completed
	FROM surveys AS s;`

	rows, err := m.conn.Query(query, types.SurveySessionStatus_InProgress, types.SurveySessionStatus_Completed)
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

func (m *Mysql) GetSurveyByField(field string, value interface{}) (*types.Survey, error) {
	query := fmt.Sprintf(`SELECT
		s.id, s.uuid, s.created_at,
		s.parse_status, s.delivery_status,
		s.error_log, s.name, s.config, s.url_slug
	FROM surveys AS s
	WHERE s.%s=?;`, field)

	row := m.conn.QueryRow(query, value)
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

func (m *Mysql) UpsertSurveyQuestions(survey *types.Survey) error {
	if survey == nil || survey.Config == nil || survey.Config.Questions == nil {
		return nil
	}

	placeholders := []string{}
	values := []interface{}{}
	values = append(values, survey.ID)
	for i := range survey.Config.Questions.Questions {
		placeholders = append(placeholders, "?")
		values = append(values, survey.Config.Questions.Questions[i].ID)
	}
	// Delete removed questions
	deleteQuery := `DELETE FROM surveys_questions
	WHERE survey_id = ?
	AND question_id NOT IN (` + strings.Join(placeholders, ", ") + `);`

	_, err := m.conn.Exec(deleteQuery, values...)
	if err != nil {
		return err
	}

	placeholders = []string{}
	values = []interface{}{}
	for _, q := range survey.Config.Questions.Questions {
		placeholders = append(placeholders, "(?, ?)")
		values = append(values, survey.ID, q.ID)
	}

	insertQuery := `INSERT INTO surveys_questions
		(survey_id, question_id)
		VALUES ` + strings.Join(placeholders, ", ") +
		` ON DUPLICATE KEY UPDATE id=id;`

	_, err = m.conn.Exec(insertQuery, values...)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mysql) GetSurveyQuestions(surveyID int64) ([]types.Question, error) {
	query := `SELECT
		uuid, question_id
	FROM surveys_questions
	WHERE survey_id=?;`

	rows, err := m.conn.Query(query, surveyID)
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

func (m *Mysql) CreateSurveySession(session *types.SurveySession) error {
	insertQuery := `INSERT INTO surveys_sessions
		(status, survey_id, ip_addr)
		VALUES (?, (SELECT id FROM surveys WHERE uuid = ?), ?);`

	_, err := m.conn.Exec(insertQuery, session.Status, session.SurveyUUID, session.IPAddr)
	if err != nil {
		return fmt.Errorf("unable to create survey session")
	}

	query := `SELECT id, uuid
		FROM surveys_sessions
		WHERE id = LAST_INSERT_ID();`

	row, err := m.conn.Query(query)
	if err != nil {
		return fmt.Errorf("unable to retrieve survey session")
	}
	return row.Scan(&session.ID, &session.UUID)
}

func (m *Mysql) UpdateSurveySessionStatus(sessionUUID string, newStatus types.SurveySessionStatus) error {
	completedAt := "NULL"
	if newStatus == types.SurveySessionStatus_Completed {
		completedAt = "NOW()"
	}

	query := fmt.Sprintf(`UPDATE surveys_sessions
		SET status = ?, completed_at = %s
		WHERE uuid = ?;`, completedAt)

	_, err := m.conn.Exec(query, newStatus, sessionUUID)

	return err
}

func (m *Mysql) GetSurveySession(surveyUUID string, sessionUUID string) (*types.SurveySession, error) {
	query := `SELECT
		ss.id, ss.uuid, ss.created_at, ss.status, s.uuid
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE ss.uuid=? AND s.uuid=?;`

	row := m.conn.QueryRow(query, sessionUUID, surveyUUID)
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

func (m *Mysql) GetSurveySessionByIPAddress(surveyUUID string, ipAddr string) (*types.SurveySession, error) {
	query := `SELECT
		ss.id, ss.uuid, ss.created_at, ss.status, s.uuid
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE s.uuid=? AND ss.ip_addr=?;`

	row := m.conn.QueryRow(query, surveyUUID, ipAddr)
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

func (m *Mysql) GetSurveySessionAnswers(sessionUUID string) ([]types.QuestionAnswer, error) {
	query := `SELECT
		q.id, q.uuid, sa.answer
	FROM surveys_answers AS sa
	LEFT JOIN surveys_questions AS q ON q.id = sa.question_id
	WHERE session_id IN (SELECT id FROM surveys_sessions WHERE uuid = ?);`

	rows, err := m.conn.Query(query, sessionUUID)
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

func (m *Mysql) UpsertSurveyQuestionAnswer(sessionUUID string, questionUUID string, answer types.Answer) error {
	query := `INSERT INTO surveys_answers
		(session_id, question_id, answer)
		VALUES (
			(SELECT id FROM surveys_sessions WHERE uuid = ?),
			(SELECT id FROM surveys_questions WHERE uuid = ?),
			?
		)
		ON DUPLICATE KEY
		UPDATE answer = ?;`

	_, err := m.conn.Exec(query, sessionUUID, questionUUID, answer, answer)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mysql) GetSurveySessionsWithAnswers(surveyUUID string, filter *types.SurveySessionsFilter) ([]types.SurveySession, int, error) {
	query := fmt.Sprintf(`
	SELECT
		ss.id, ss.uuid, ss.created_at, ss.completed_at, ss.status, q.id, q.uuid, sa.answer
	FROM (
		SELECT * from surveys_sessions
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	) AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	LEFT JOIN surveys_answers AS sa ON sa.session_id = ss.id
	LEFT JOIN surveys_questions AS q ON q.id = sa.question_id
	WHERE s.uuid=?
	ORDER BY ss.%s %s
	;`, filter.SortBy, filter.Order, filter.Limit, filter.Offset, filter.SortBy, filter.Order)

	rows, err := m.conn.Query(query, surveyUUID)
	if err != nil {
		return nil, 0, err
	}

	sessions := []types.SurveySession{}
	sessionsMap := map[string]types.SurveySession{}
	for rows.Next() {
		session := types.SurveySession{}
		answer := types.QuestionAnswer{}
		var (
			questionID   sql.NullString
			questionUUID sql.NullString
		)

		err := rows.Scan(&session.ID, &session.UUID, &session.CreatedAt, &session.CompletedAt, &session.Status, &questionID, &questionUUID, &answer.AnswerBytes)
		if err != nil {
			return nil, 0, err
		}

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

	totalCount, err := m.getSurveySessionsCount(surveyUUID)
	if err != nil {
		return nil, 0, err
	}

	for i, session := range sessions {
		fullSession := sessionsMap[session.UUID]
		sessions[i].QuestionAnswers = fullSession.QuestionAnswers
	}

	return sessions, totalCount, nil
}

func (m *Mysql) getSurveySessionsCount(surveyUUID string) (int, error) {
	query := `SELECT
		COUNT(*)
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE s.uuid=?;`

	row := m.conn.QueryRow(query, surveyUUID)
	var count int
	err := row.Scan(&count)
	return count, err
}
