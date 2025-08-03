package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
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
	configBytes, err := json.Marshal(survey.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal survey config: %w", err)
	}

	uuid, err := db.DecodeUUID(survey.UUID)
	if err != nil {
		return fmt.Errorf("failed to decode UUID: %w", err)
	}

	return p.queries.UpdateSurvey(p.ctx, db.UpdateSurveyParams{
		ParseStatus:    db.NullSurveyParseStatuses{Valid: true, SurveyParseStatuses: db.SurveyParseStatuses(survey.ParseStatus)},
		DeliveryStatus: db.NullSurveyDeliveryStatuses{Valid: true, SurveyDeliveryStatuses: db.SurveyDeliveryStatuses(survey.DeliveryStatus)},
		ErrorLog:       pgtype.Text{Valid: true, String: survey.ErrorLog},
		Name:           survey.Name,
		Config:         configBytes,
		UrlSlug:        survey.URLSlug,
		Uuid:           uuid,
	})
}

func (p *Postgres) GetSurveys() ([]*types.Survey, error) {
	rows, err := p.queries.GetSurveys(p.ctx, db.GetSurveysParams{
		Status: db.NullSurveysSessionsStatus{
			Valid:                 true,
			SurveysSessionsStatus: db.SurveysSessionsStatus(types.SurveySessionStatus_InProgress),
		},
		Status_2: db.NullSurveysSessionsStatus{
			Valid:                 true,
			SurveysSessionsStatus: db.SurveysSessionsStatus(types.SurveySessionStatus_Completed),
		},
	})
	if err != nil {
		return nil, err
	}

	surveys := []*types.Survey{}
	for _, row := range rows {
		survey := &types.Survey{
			ID:             int64(row.ID),
			UUID:           db.EncodeUUID(row.Uuid),
			CreatedAt:      row.CreatedAt.Time,
			ParseStatus:    types.SurveyParseStatus(row.ParseStatus.SurveyParseStatuses),
			DeliveryStatus: types.SurveyDeliveryStatus(row.DeliveryStatus.SurveyDeliveryStatuses),
			ErrorLog:       row.ErrorLog.String,
			Name:           row.Name,
			URLSlug:        row.UrlSlug,
		}

		if err := json.Unmarshal(row.Config, &survey.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal survey config: %w", err)
		}

		survey.Stats.SessionsCountInProgess = int(row.SessionsCountInProgress)
		survey.Stats.SessionsCountCompleted = int(row.SessionsCountCompleted)

		totalResponses := survey.Stats.SessionsCountInProgess + survey.Stats.SessionsCountCompleted
		if totalResponses > 0 {
			survey.Stats.CompletionRate = (survey.Stats.SessionsCountCompleted * 100) / totalResponses
		}

		surveys = append(surveys, survey)
	}

	return surveys, nil
}

func (p *Postgres) GetSurveyByField(field string, value interface{}) (*types.Survey, error) {
	var survey *types.Survey
	var err error

	switch field {
	case "uuid":
		uuid, err := db.DecodeUUID(value.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to decode UUID: %w", err)
		}
		row, err := p.queries.GetSurveyByUUID(p.ctx, uuid)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		survey = &types.Survey{
			ID:             int64(row.ID),
			UUID:           db.EncodeUUID(row.Uuid),
			CreatedAt:      row.CreatedAt.Time,
			ParseStatus:    types.SurveyParseStatus(row.ParseStatus.SurveyParseStatuses),
			DeliveryStatus: types.SurveyDeliveryStatus(row.DeliveryStatus.SurveyDeliveryStatuses),
			ErrorLog:       row.ErrorLog.String,
			Name:           row.Name,
			URLSlug:        row.UrlSlug,
		}
		if err := json.Unmarshal(row.Config, &survey.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal survey config: %w", err)
		}
	case "url_slug":
		row, err := p.queries.GetSurveyByURLSlug(p.ctx, value.(string))
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		survey = &types.Survey{
			ID:             int64(row.ID),
			UUID:           db.EncodeUUID(row.Uuid),
			CreatedAt:      row.CreatedAt.Time,
			ParseStatus:    types.SurveyParseStatus(row.ParseStatus.SurveyParseStatuses),
			DeliveryStatus: types.SurveyDeliveryStatus(row.DeliveryStatus.SurveyDeliveryStatuses),
			ErrorLog:       row.ErrorLog.String,
			Name:           row.Name,
			URLSlug:        row.UrlSlug,
		}
		if err := json.Unmarshal(row.Config, &survey.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal survey config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported field: %s", field)
	}

	return survey, err
}

func (p *Postgres) UpsertSurveyQuestions(survey *types.Survey) error {
	if survey == nil || survey.Config == nil || survey.Config.Questions == nil {
		return nil
	}

	questionIds := make([]string, len(survey.Config.Questions.Questions))
	for i, q := range survey.Config.Questions.Questions {
		questionIds[i] = q.ID
	}

	// Delete removed questions
	err := p.queries.DeleteSurveyQuestionsNotInList(p.ctx, db.DeleteSurveyQuestionsNotInListParams{
		SurveyID: int32(survey.ID),
		Column2:  questionIds,
	})
	if err != nil {
		return err
	}

	// Insert/update questions
	for _, q := range survey.Config.Questions.Questions {
		err := p.queries.UpsertSurveyQuestion(p.ctx, db.UpsertSurveyQuestionParams{
			SurveyID:   int32(survey.ID),
			QuestionID: q.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Postgres) GetSurveyQuestions(surveyID int64) ([]types.Question, error) {
	rows, err := p.queries.GetSurveyQuestions(p.ctx, int32(surveyID))
	if err != nil {
		return []types.Question{}, err
	}

	questions := []types.Question{}
	for _, row := range rows {
		question := types.Question{
			UUID: db.EncodeUUID(row.Uuid),
			ID:   row.QuestionID,
		}
		questions = append(questions, question)
	}

	return questions, nil
}

func (p *Postgres) CreateSurveySession(session *types.SurveySession) error {
	surveyUUID, err := db.DecodeUUID(session.SurveyUUID)
	if err != nil {
		return fmt.Errorf("failed to decode survey UUID: %w", err)
	}

	row, err := p.queries.CreateSurveySession(p.ctx, db.CreateSurveySessionParams{
		Status: db.NullSurveysSessionsStatus{Valid: true, SurveysSessionsStatus: db.SurveysSessionsStatus(session.Status)},
		Uuid:   surveyUUID,
		IpAddr: pgtype.Text{Valid: true, String: session.IPAddr},
	})
	if err != nil {
		return err
	}

	session.ID = int64(row.ID)
	session.UUID = db.EncodeUUID(row.Uuid)
	return nil
}

func (p *Postgres) UpdateSurveySessionStatus(sessionUUID string, newStatus types.SurveySessionStatus) error {
	uuid, err := db.DecodeUUID(sessionUUID)
	if err != nil {
		return fmt.Errorf("failed to decode session UUID: %w", err)
	}

	if newStatus == types.SurveySessionStatus_Completed {
		return p.queries.UpdateSurveySessionStatusCompleted(p.ctx, db.UpdateSurveySessionStatusCompletedParams{
			Status: db.NullSurveysSessionsStatus{Valid: true, SurveysSessionsStatus: db.SurveysSessionsStatus(newStatus)},
			Uuid:   uuid,
		})
	} else {
		return p.queries.UpdateSurveySessionStatus(p.ctx, db.UpdateSurveySessionStatusParams{
			Status: db.NullSurveysSessionsStatus{Valid: true, SurveysSessionsStatus: db.SurveysSessionsStatus(newStatus)},
			Uuid:   uuid,
		})
	}
}

func (p *Postgres) GetSurveySession(surveyUUID string, sessionUUID string) (*types.SurveySession, error) {
	sessionUUIDPg, err := db.DecodeUUID(sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session UUID: %w", err)
	}

	surveyUUIDPg, err := db.DecodeUUID(surveyUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode survey UUID: %w", err)
	}

	row, err := p.queries.GetSurveySession(p.ctx, db.GetSurveySessionParams{
		Uuid:   sessionUUIDPg,
		Uuid_2: surveyUUIDPg,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	session := &types.SurveySession{
		ID:         int64(row.ID),
		UUID:       db.EncodeUUID(row.Uuid),
		CreatedAt:  row.CreatedAt.Time,
		Status:     types.SurveySessionStatus(row.Status.SurveysSessionsStatus),
		SurveyUUID: db.EncodeUUID(row.SurveyUuid),
	}

	return session, nil
}

func (p *Postgres) DeleteSurveySession(sessionUUID string) error {
	uuid, err := db.DecodeUUID(sessionUUID)
	if err != nil {
		return fmt.Errorf("failed to decode session UUID: %w", err)
	}

	return p.queries.DeleteSurveySession(p.ctx, uuid)
}

func (p *Postgres) GetSurveySessionByIPAddress(surveyUUID string, ipAddr string) (*types.SurveySession, error) {
	surveyUUIDPg, err := db.DecodeUUID(surveyUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode survey UUID: %w", err)
	}

	row, err := p.queries.GetSurveySessionByIPAddress(p.ctx, db.GetSurveySessionByIPAddressParams{
		Uuid:   surveyUUIDPg,
		IpAddr: pgtype.Text{Valid: true, String: ipAddr},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	session := &types.SurveySession{
		ID:         int64(row.ID),
		UUID:       db.EncodeUUID(row.Uuid),
		CreatedAt:  row.CreatedAt.Time,
		Status:     types.SurveySessionStatus(row.Status.SurveysSessionsStatus),
		SurveyUUID: db.EncodeUUID(row.SurveyUuid),
	}

	return session, nil
}

func (p *Postgres) GetSurveySessionAnswers(sessionUUID string) ([]types.QuestionAnswer, error) {
	sessionUUIDPg, err := db.DecodeUUID(sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session UUID: %w", err)
	}

	rows, err := p.queries.GetSurveySessionAnswers(p.ctx, sessionUUIDPg)
	if err != nil {
		return nil, err
	}

	answers := []types.QuestionAnswer{}
	for _, row := range rows {
		answer := types.QuestionAnswer{
			QuestionID:   row.QuestionID.String,
			QuestionUUID: db.EncodeUUID(row.QuestionUuid),
			AnswerBytes:  row.Answer,
		}
		answers = append(answers, answer)
	}

	return answers, nil
}

func (p *Postgres) UpsertSurveyQuestionAnswer(sessionUUID string, questionUUID string, answer types.Answer) error {
	sessionUUIDPg, err := db.DecodeUUID(sessionUUID)
	if err != nil {
		return fmt.Errorf("failed to decode session UUID: %w", err)
	}

	questionUUIDPg, err := db.DecodeUUID(questionUUID)
	if err != nil {
		return fmt.Errorf("failed to decode question UUID: %w", err)
	}

	// Convert answer to bytes via JSON marshaling
	answerBytes, err := json.Marshal(answer)
	if err != nil {
		return fmt.Errorf("failed to marshal answer: %w", err)
	}

	return p.queries.UpsertSurveyQuestionAnswer(p.ctx, db.UpsertSurveyQuestionAnswerParams{
		Uuid:   sessionUUIDPg,
		Uuid_2: questionUUIDPg,
		Answer: answerBytes,
	})
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
	surveyUUIDPg, err := db.DecodeUUID(surveyUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to decode survey UUID: %w", err)
	}

	count, err := p.queries.GetSurveySessionsCount(p.ctx, surveyUUIDPg)
	return int(count), err
}

func (p *Postgres) StoreWebhookResponse(sessionId int, responseStatus int, response string) error {
	now := time.Now().UTC()
	return p.queries.StoreWebhookResponse(p.ctx, db.StoreWebhookResponseParams{
		CreatedAt:      pgtype.Timestamp{Time: now, Valid: true},
		SessionID:      int32(sessionId),
		ResponseStatus: int32(responseStatus),
		Response:       pgtype.Text{String: response, Valid: true},
	})
}
