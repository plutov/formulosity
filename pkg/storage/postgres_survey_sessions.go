package storage

import (
	"database/sql"
	"fmt"

	"github.com/plutov/formulosity/pkg/types"
)

func (p *Postgres) CreateSurveySession(session *types.SurveySession) error {
	query := `INSERT INTO surveys_sessions
		(status, survey_id, ip_addr)
		VALUES ($1, (SELECT id FROM surveys WHERE uuid = $2), $3)
		RETURNING id, uuid;`

	row := p.conn.QueryRow(query, session.Status, session.SurveyUUID, session.IPAddr)
	if row == nil {
		return fmt.Errorf("unable to create survey session")
	}

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

	_, err := p.conn.Exec(query, newStatus, sessionUUID)

	return err
}

func (p *Postgres) GetSurveySession(surveyUUID string, sessionUUID string) (*types.SurveySession, error) {
	query := `SELECT
		ss.id, ss.uuid, ss.created_at, ss.status, s.uuid
	FROM surveys_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	WHERE ss.uuid=$1 AND s.uuid=$2;`

	row := p.conn.QueryRow(query, sessionUUID, surveyUUID)
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

func (p *Postgres) GetSurveySessionByIPAddress(surveyUUID string, ipAddr string) (*types.SurveySession, error) {
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

func (p *Postgres) GetSurveySessionAnswers(sessionUUID string) ([]types.QuestionAnswer, error) {
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

	_, err := p.conn.Exec(query, sessionUUID, questionUUID, answer)
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
		ss.id, ss.uuid, ss.created_at, ss.completed_at, ss.status, q.id, q.uuid, sa.answer
	FROM limited_sessions AS ss
	INNER JOIN surveys AS s ON s.id = ss.survey_id
	LEFT JOIN surveys_answers AS sa ON sa.session_id = ss.id
	LEFT JOIN surveys_questions AS q ON q.id = sa.question_id
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

	row := p.conn.QueryRow(query, surveyUUID)
	var count int
	err := row.Scan(&count)
	return count, err
}
