package surveys

import (
	"encoding/json"
	"errors"

	"github.com/plutov/formulosity/api/pkg/services"
	"github.com/plutov/formulosity/api/pkg/types"
)

// returns 2 errors: general and error details
func SubmitAnswer(svc services.Services, session *types.SurveySession, survey *types.Survey, question *types.Question, req []byte, file *types.File) (error, error) {
	logCtx := svc.Logger.With("session_uuid", session.UUID)
	logCtx.Info("submitting answer")

	answers, err := sessionAnswerMap(session, survey)
	if err != nil {
		return errors.New("unable to read existing answers"), nil
	}
	visible, err := survey.Config.IsQuestionVisible(*question, answers)
	if err != nil {
		return errors.New("invalid visibility condition"), err
	}
	if !visible {
		return errors.New("question is hidden"), nil
	}

	// Option labels may contain answer piping. Validate against rendered labels
	// without mutating the persisted survey template.
	validationQuestion := *question
	for i, option := range validationQuestion.Options {
		rendered, renderErr := types.Interpolate(option, answers)
		if renderErr != nil {
			return errors.New("invalid option template"), renderErr
		}
		validationQuestion.Options[i] = rendered
	}

	answer, err := question.GetAnswerType()
	if err != nil {
		return err, nil
	}

	switch a := answer.(type) {
	case *types.FileAnswer:
		if file != nil {
			a.FileSize = file.Size
			a.FileFormat = file.Format

			if err := answer.Validate(validationQuestion); err != nil {
				return errors.New("invalid answer"), err
			}

			filePath, err := svc.FileStorage.SaveFile(file)
			if err != nil {
				return errors.New("unable to save file"), nil
			}
			a.AnswerValue = filePath
		} else {
			return errors.New("file is required for this question type"), nil
		}
	default:
		if err := json.Unmarshal(req, &answer); err != nil {
			return errors.New("invalid request format"), nil
		}

		if err := answer.Validate(validationQuestion); err != nil {
			return errors.New("invalid answer"), err
		}
	}

	answerValue, err := types.AnswerValue(answer)
	if err != nil {
		return errors.New("invalid answer"), nil
	}
	answers[question.ID] = answerValue
	if err := svc.Storage.UpsertSurveyQuestionAnswer(session.UUID, question.UUID, answer); err != nil {
		msg := "unable to insert answer"
		logCtx.Error(msg, "err", err)
		return errors.New(msg), nil
	}

	logCtx.Info("answer submitted")

	// Re-evaluate in order and remove answers made stale by this submission.
	for _, q := range survey.Config.Questions.Questions {
		isVisible, evalErr := survey.Config.IsQuestionVisible(q, answers)
		if evalErr != nil {
			return errors.New("invalid visibility condition"), evalErr
		}
		if !isVisible {
			if err := svc.Storage.DeleteSurveyQuestionAnswer(session.UUID, q.UUID); err != nil {
				return errors.New("unable to clear hidden answer"), nil
			}
			delete(answers, q.ID)
		}
	}

	// mark session as completed if there are no more unanswered visible questions
	isCompleted := isSessionCompleted(survey, answers, session)

	if isCompleted {
		session.Status = types.SurveySessionStatus_Completed
		if err := svc.Storage.UpdateSurveySessionStatus(session.UUID, session.Status); err != nil {
			msg := "unable to update session status"
			logCtx.Error(msg, "err", err)
			return nil, errors.New(msg)
		}

		logCtx.Info("session completed")
	}

	return nil, nil
}

func sessionAnswerMap(session *types.SurveySession, survey *types.Survey) (map[string]any, error) {
	answers := make(map[string]any)
	for _, a := range session.QuestionAnswers {
		for _, q := range survey.Config.Questions.Questions {
			if q.UUID == a.QuestionUUID {
				if a.Answer != nil {
					value, err := types.AnswerValue(a.Answer)
					if err != nil {
						return nil, err
					}
					answers[q.ID] = value
				}
				break
			}
		}
	}
	return answers, nil
}

func isSessionCompleted(survey *types.Survey, answers map[string]any, session *types.SurveySession) bool {
	if session.Status == types.SurveySessionStatus_Completed {
		return true
	}
	for _, q := range survey.Config.Questions.Questions {
		visible, err := survey.Config.IsQuestionVisible(q, answers)
		if err != nil || !visible {
			continue
		}
		if _, ok := answers[q.ID]; !ok {
			return false
		}
	}
	return true
}
