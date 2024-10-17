package storage

import "github.com/plutov/formulosity/api/pkg/types"

type Interface interface {
	Init() error
	Ping() error
	Close() error
	Migrate() error

	CreateSurvey(survey *types.Survey) error
	UpdateSurvey(survey *types.Survey) error
	GetSurveys() ([]*types.Survey, error)
	GetSurveyByField(field string, value interface{}) (*types.Survey, error)

	CreateSurveySession(session *types.SurveySession) error
	UpdateSurveySessionStatus(sessionUUID string, newStatus types.SurveySessionStatus) error
	GetSurveySessionByIPAddress(surveyUUID string, ipAddr string) (*types.SurveySession, error)
	GetSurveySession(surveyUUID string, sessionUUID string) (*types.SurveySession, error)
	UpsertSurveyQuestions(survey *types.Survey) error
	GetSurveyQuestions(surveyID int64) ([]types.Question, error)
	GetSurveySessionsWithAnswers(surveyUUID string, filter *types.SurveySessionsFilter) ([]types.SurveySession, int, error)
	GetSurveySessionAnswers(sessionUUID string) ([]types.QuestionAnswer, error)
	UpsertSurveyQuestionAnswer(sessionUUID string, questionUUID string, answer types.Answer) error
}

type FileInterface interface {
	Init() error

	SaveFile(file *types.File) (string, error)
	IsFileExist(fileName string) (bool, string, error)
}
