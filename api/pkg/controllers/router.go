package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/plutov/formulosity/api/pkg/http/response"
)

// NewRouter returns new router
func NewRouter(h *Handler) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/", h.healthCheckHandler)
	e.GET("/app/surveys", h.getSurveys)
	e.PATCH("/app/surveys/:survey_uuid", h.surveyUUIDMiddleware(h.updateSurvey))
	e.GET("/app/surveys/:survey_uuid/sessions", h.surveyUUIDMiddleware(h.getSurveySessions))
	e.GET("/app/surveys/:survey_uuid/download/:file_name", h.surveyUUIDMiddleware(h.downloadFile))

	surveys := e.Group("/surveys")
	surveys.GET("/:url_slug", h.getSurvey)
	surveys.GET("/:url_slug/css", h.getSurveyCSS)
	surveys.PUT("/:url_slug/sessions", h.createSurveySession)
	surveys.GET("/:url_slug/sessions/:session_uuid", h.getSurveySessionHandler)
	surveys.POST("/:url_slug/sessions/:session_uuid/questions/:question_uuid/answers", h.submitSurveyAnswer)

	return e
}

func (h *Handler) healthCheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func (h *Handler) surveyUUIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		surveyUUID := c.Param("survey_uuid")
		if surveyUUID == "" {
			return response.BadRequest(c, "survey_uuid is required")
		}

		survey, err := h.Services.Storage.GetSurveyByField("uuid", surveyUUID)
		if err != nil {
			return response.NotFound(c, "survey not found")
		}

		c.Set("survey", *survey)

		return next(c)
	}
}
