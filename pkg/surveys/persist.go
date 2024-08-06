package surveys

import (
	"fmt"

	"github.com/plutov/formulosity/pkg/log"
	"github.com/plutov/formulosity/pkg/services"
	"github.com/plutov/formulosity/pkg/types"
)

// Use cases
// 1. When it's a new survey - create it
// 2. When it's an existing survey - update it
func PersistSurveysSyncResult(svc services.Services, syncResult *types.SurveysSyncResult) error {
	logCtx := log.With("func", "PersistSurveysSyncResult")
	logCtx.Info("persisting surveys")

	if syncResult == nil {
		return fmt.Errorf("syncResult is nil")
	}

	currSurveys, err := svc.Storage.GetSurveys()
	if err != nil {
		logCtx.WithError(err).Error("unable to get current surveys")
		return fmt.Errorf("unable to get current surveys: %w", err)
	}

	surveysToUpdate := []*types.Survey{}
	surveysToCreate := []*types.Survey{}

	// find surveys to update
	for _, currSurvey := range currSurveys {
		currSurveyCopy := *currSurvey
		isDeleted := true
		for _, survey := range syncResult.Surveys {
			surveyCopy := *survey
			if currSurvey.Name == survey.Name {
				currSurveyCopy.ParseStatus = types.SurveyParseStatus_Success
				currSurveyCopy.ErrorLog = ""
				currSurveyCopy.Config = surveyCopy.Config

				surveysToUpdate = append(surveysToUpdate, &currSurveyCopy)
				isDeleted = false
			}
		}
		for _, errorSurvey := range syncResult.Errors {
			if currSurvey.Name == errorSurvey.Name {
				currSurveyCopy.ParseStatus = types.SurveyParseStatus_Error
				currSurveyCopy.ErrorLog = errorSurvey.ErrString

				surveysToUpdate = append(surveysToUpdate, &currSurveyCopy)
				isDeleted = false
			}
		}

		if isDeleted {
			currSurveyCopy.ParseStatus = types.SurveyParseStatus_Deleted
			currSurveyCopy.ErrorLog = ""
			surveysToUpdate = append(surveysToUpdate, &currSurveyCopy)
		}
	}

	// find surveys to create (parse result: success)
	for _, survey := range syncResult.Surveys {
		surveyCopy := *survey

		isNew := true
		for _, currSurvey := range currSurveys {
			if currSurvey.Name == survey.Name {
				isNew = false
				break
			}
		}
		if !isNew {
			continue
		}

		surveyCopy.ParseStatus = types.SurveyParseStatus_Success
		surveyCopy.DeliveryStatus = types.SurveyDeliveryStatus_Launched
		surveyCopy.ErrorLog = ""
		surveysToCreate = append(surveysToCreate, &surveyCopy)
	}

	// find surveys to create (parse result: error)
	for _, survey := range syncResult.Errors {
		isNew := true
		for _, currSurvey := range currSurveys {
			if currSurvey.Name == survey.Name {
				isNew = false
				break
			}
		}
		if !isNew {
			continue
		}

		surveysToCreate = append(surveysToCreate, &types.Survey{
			ParseStatus:    types.SurveyParseStatus_Error,
			DeliveryStatus: types.SurveyDeliveryStatus_Stopped,
			ErrorLog:       survey.ErrString,
			Config:         nil,
			Name:           survey.Name,
		})
	}

	// create surveys
	for _, survey := range surveysToCreate {
		surveyToCreate := *survey
		if err := CreateSurvey(svc, &surveyToCreate); err != nil {
			return err
		}
	}

	// update surveys
	for _, survey := range surveysToUpdate {
		surveyToUpdate := *survey
		if err := UpdateSurvey(svc, &surveyToUpdate); err != nil {
			return err
		}
	}

	logCtx.Info("surveys persisted")

	return nil
}
