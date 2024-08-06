package controllers

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/plutov/formulosity/pkg/http/response"
)

// NewRouter returns new router
func NewRouter(h *Handler) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			internalHosts := map[string]bool{
				os.Getenv("UI_ADDR"): true,
			}
			_, ok := internalHosts[origin]
			return ok, nil
		},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/", h.healthCheckHandler)
	e.GET("/app/surveys", h.getSurveys)
	e.PATCH("/app/surveys/:survey_uuid", h.surveyUUIDMiddleware(h.updateSurvey))
	e.GET("/app/surveys/:survey_uuid/sessions", h.surveyUUIDMiddleware(h.getSurveySessions))

	surveysGroup := e.Group("/surveys")
	surveysGroup.GET("/:url_slug", h.getSurvey)
	surveysGroup.PUT("/:url_slug/sessions", h.createSurveySession)
	surveysGroup.GET("/:url_slug/sessions/:session_uuid", h.getSurveySessionHandler)
	surveysGroup.POST("/:url_slug/sessions/:session_uuid/questions/:question_uuid/answers", h.submitSurveyAnswer)

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
