package surveys

import (
	"encoding/json"
	"errors"

	"github.com/plutov/formulosity/pkg/log"
	"github.com/plutov/formulosity/pkg/services"
	"github.com/plutov/formulosity/pkg/types"
)

// returns 2 errors: general and error details
func SubmitAnswer(svc services.Services, session *types.SurveySession, survey *types.Survey, question *types.Question, req []byte) (error, error) {
	logCtx := log.With("session_uuid", session.UUID)
	logCtx.Info("submitting answer")

	answer, err := question.GetAnswerType()
	if err != nil {
		return err, nil
	}

	if err := json.Unmarshal(req, &answer); err != nil {
		return errors.New("invalid request format"), nil
	}

	if err := answer.Validate(*question); err != nil {
		return errors.New("invalid answer"), err
	}

	if err := svc.Storage.UpsertSurveyQuestionAnswer(session.UUID, question.UUID, answer); err != nil {
		msg := "unable to insert answer"
		logCtx.WithError(err).Error(msg)
		return errors.New(msg), nil
	}

	logCtx.Info("answer submitted")

	// mark session as completed if there are no more unanswered questions
	isCompleted := isSessionCompleted(survey, session, question)

	if isCompleted {
		session.Status = types.SurveySessionStatus_Completed
		if err := svc.Storage.UpdateSurveySessionStatus(session.UUID, session.Status); err != nil {
			msg := "unable to update session status"
			logCtx.WithError(err).Error(msg)
			return nil, errors.New(msg)
		}

		logCtx.Info("session completed")
	}

	return nil, nil
}

func isSessionCompleted(survey *types.Survey, session *types.SurveySession, question *types.Question) bool {
	if session.Status == types.SurveySessionStatus_Completed {
		return true
	}

	for _, q := range survey.Config.Questions.Questions {
		hasAnswer := q.UUID == question.UUID
		for _, a := range session.QuestionAnswers {
			if q.UUID == a.QuestionUUID {
				hasAnswer = true
				break
			}
		}

		if !hasAnswer {
			return false
		}
	}

	return true
}
