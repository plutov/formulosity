package controllers

import (
	"errors"
	"io"

	"github.com/plutov/formulosity/pkg/http/response"
	"github.com/plutov/formulosity/pkg/surveys"
	surveyspkg "github.com/plutov/formulosity/pkg/surveys"
	"github.com/plutov/formulosity/pkg/types"
	"github.com/labstack/echo/v4"
)

func (h *Handler) createSurveySession(c echo.Context) error {
	survey, err := h.getLaunchedSurvey(c)
	if err != nil {
		return response.NotFound(c, err.Error())
	}

	ipAddr := c.RealIP()
	session, err := surveys.CreateSurveySession(h.Services, survey, ipAddr)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}

	return response.Ok(c, *session)
}

func (h *Handler) getSurveySessionHandler(c echo.Context) error {
	session, _, err := h.getSurveySession(c)
	if err != nil {
		return response.NotFound(c, err.Error())
	}

	return response.Ok(c, *session)
}

func (h *Handler) getSurveySession(c echo.Context) (*types.SurveySession, *types.Survey, error) {
	sessionUUID := c.Param("session_uuid")
	if sessionUUID == "" {
		return nil, nil, errors.New("session_uuid is required")
	}

	survey, err := h.getLaunchedSurvey(c)
	if err != nil {
		return nil, nil, err
	}

	session, err := surveys.GetSurveySession(h.Services, *survey, sessionUUID)
	if err != nil {
		return nil, nil, errors.New("session not found")
	}

	return session, survey, nil
}

func (h *Handler) submitSurveyAnswer(c echo.Context) error {
	questionUUID := c.Param("question_uuid")
	if questionUUID == "" {
		return response.BadRequest(c, "question_uuid is required")
	}

	session, survey, err := h.getSurveySession(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	if session.Status != types.SurveySessionStatus_InProgress {
		return response.BadRequest(c, "session is not in progress")
	}

	question, err := survey.Config.FindQuestionByUUID(questionUUID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	req, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	mainErr, detailsErr := surveys.SubmitAnswer(h.Services, session, survey, question, req)
	if mainErr != nil {
		if detailsErr != nil {
			return response.BadRequestWithDetails(c, mainErr.Error(), detailsErr.Error())
		}

		return response.BadRequest(c, mainErr.Error())
	}

	session, _, err = h.getSurveySession(c)
	if err != nil {
		return response.NotFound(c, err.Error())
	}

	return response.Ok(c, *session)
}

func (h *Handler) getSurveySessions(c echo.Context) error {
	surveyCtx := c.Get("survey").(types.Survey)

	req := new(types.SurveySessionsFilter)
	if err := c.Bind(req); err != nil {
		return response.BadRequestDefaultMessage(c)
	}
	if err := req.Validate(); err != nil {
		return response.BadRequest(c, err.Error())
	}

	survey, err := surveyspkg.GetSurveyByUUID(h.Services, surveyCtx.UUID)
	if err != nil || survey == nil {
		return response.BadRequest(c, "survey not found")
	}

	sessions, pagesCount, err := surveyspkg.GetSurveySessions(h.Services, *survey, req)
	if err != nil {
		return response.InternalErrorDefaultMsg(c)
	}

	return response.Ok(c, echo.Map{
		"survey":      *survey,
		"sessions":    sessions,
		"pages_count": pagesCount,
	})
}
