package surveys

import (
	"errors"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/plutov/formulosity/pkg/log"
	"github.com/plutov/formulosity/pkg/services"
	"github.com/plutov/formulosity/pkg/types"
)

const URL_SLUG_LENGTH = 12

func CreateSurvey(svc services.Services, survey *types.Survey) error {
	logCtx := log.With("survey", *survey)
	logCtx.Info("creating survey")

	var err error
	survey.URLSlug, err = gonanoid.Generate("abcdefghijklmnopqrstuvwxyz1234567890", URL_SLUG_LENGTH)
	if err != nil {
		msg := "unable to generate url_slug"
		logCtx.WithError(err).Error(msg)
		return errors.New(msg)
	}

	survey.CustomThemeURL, err = UploadCustomTheme(svc, survey.URLSlug, survey.Config.ThemeContents)
	if err != nil {
		msg := "unable to upload custom theme"
		logCtx.WithError(err).Error(msg)
		return errors.New(msg)
	}

	if err := svc.Storage.CreateSurvey(survey); err != nil {
		msg := "unable to create survey"
		logCtx.WithError(err).Error(msg)
		return errors.New(msg)
	}

	if err := svc.Storage.UpsertSurveyQuestions(survey); err != nil {
		msg := "unable to upsert survey questions"
		logCtx.WithError(err).Error(msg)
		return errors.New(msg)
	}

	logCtx.Info("survey created")

	return nil
}

func UpdateSurvey(svc services.Services, survey *types.Survey) error {
	logCtx := log.With("survey_uuid", survey.UUID)
	logCtx.Info("updating survey")

	var err error
	survey.CustomThemeURL, err = UploadCustomTheme(svc, survey.URLSlug, survey.Config.ThemeContents)
	if err != nil {
		msg := "unable to upload custom theme"
		logCtx.WithError(err).Error(msg)
		return errors.New(msg)
	}

	if err := svc.Storage.UpdateSurvey(survey); err != nil {
		msg := "unable to update survey"
		logCtx.WithError(err).Error(msg)
		return errors.New(msg)
	}

	if err := svc.Storage.UpsertSurveyQuestions(survey); err != nil {
		msg := "unable to upsert survey questions"
		logCtx.WithError(err).Error(msg)
		return errors.New(msg)
	}

	logCtx.Info("survey updated")

	return nil
}

func GetSurvey(svc services.Services, urlSlug string) (*types.Survey, error) {
	if len(urlSlug) != URL_SLUG_LENGTH {
		return nil, errors.New("invalid url_slug")
	}

	survey, err := getSurveyByField(svc, "url_slug", urlSlug)
	if err != nil {
		return nil, err
	}

	return survey, nil
}

func GetSurveyByUUID(svc services.Services, uuid string) (*types.Survey, error) {
	return getSurveyByField(svc, "uuid", uuid)
}

func getSurveyByField(svc services.Services, field string, value string) (*types.Survey, error) {
	logCtx := log.With(field, value)
	logCtx.Info("getting survey")

	survey, err := svc.Storage.GetSurveyByField(field, value)
	if err != nil {
		logCtx.WithError(err).Error("unable to get survey")
		return nil, errors.New("survey not found")
	}

	// survey not found
	if survey == nil {
		return nil, errors.New("survey not found")
	}

	questionsDB, err := svc.Storage.GetSurveyQuestions(survey.ID)
	if err != nil {
		msg := "survey questions not found"
		logCtx.WithError(err).Error(msg)
		return nil, errors.New(msg)
	}

	questionsDBMap := map[string]types.Question{}
	for _, q := range questionsDB {
		questionsDBMap[q.ID] = q
	}

	// only keep questions in Config found in the DB
	filteredQuestions := []types.Question{}
	for _, q := range survey.Config.Questions.Questions {
		if questionDB, ok := questionsDBMap[q.ID]; ok {
			q.UUID = questionDB.UUID
			filteredQuestions = append(filteredQuestions, q)
		}
	}
	survey.Config.Questions.Questions = filteredQuestions

	return survey, nil
}
