package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/plutov/formulosity/api/pkg/types"
)

func (p *Postgres) CreateSurvey(survey *types.Survey) error {
	query := `INSERT INTO surveys
		(parse_status, delivery_status, error_log, name, config, url_slug)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;`

	row := p.conn.QueryRow(query, survey.ParseStatus, survey.DeliveryStatus, survey.ErrorLog, survey.Name, survey.Config, survey.URLSlug)
	if row == nil {
		return fmt.Errorf("unable to create survey")
	}

	if err := row.Scan(&survey.ID); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) UpdateSurvey(survey *types.Survey) error {
	query := `UPDATE surveys
		SET parse_status=$1, delivery_status=$2, error_log=$3, name=$4, config=$5, url_slug=$6
		WHERE uuid=$7;`

	_, err := p.conn.Exec(query, survey.ParseStatus, survey.DeliveryStatus, survey.ErrorLog, survey.Name, survey.Config, survey.URLSlug, survey.UUID)
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

	rows, err := p.conn.Query(query, types.SurveySessionStatus_InProgress, types.SurveySessionStatus_Completed)
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

	row := p.conn.QueryRow(query, value)
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

	_, err := p.conn.Exec(deleteQuery, values...)
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

		_, err := p.conn.Exec(insertQuery, survey.ID, q.ID)
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
