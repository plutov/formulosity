package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type DataResponse struct {
	Code         int         `json:"code"`
	Message      string      `json:"message"`
	Data         interface{} `json:"data"`
	ErrorDetails string      `json:"error_details"`
}

// Ok returns status 200 with data.
func Ok(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, DataResponse{
		Code:    http.StatusOK,
		Message: "",
		Data:    data,
	})
}

// OkWithMsg returns status 200 with a message and data.
func OkWithMsg(c echo.Context, msg string, data interface{}) error {
	return c.JSON(http.StatusOK, DataResponse{
		Code:    http.StatusOK,
		Message: msg,
		Data:    data,
	})
}

// Created returns status 201 with a message and data.
func Created(c echo.Context, msg string, data interface{}) error {
	return c.JSON(http.StatusCreated, DataResponse{
		Code:    http.StatusCreated,
		Message: msg,
		Data:    data,
	})
}

func BadRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, DataResponse{
		Code:    http.StatusBadRequest,
		Message: msg,
	})
}

func BadRequestWithDetails(c echo.Context, msg string, details string) error {
	return c.JSON(http.StatusBadRequest, DataResponse{
		Code:         http.StatusBadRequest,
		Message:      msg,
		ErrorDetails: details,
	})
}

func BadRequestDefaultMessage(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, DataResponse{
		Code:    http.StatusBadRequest,
		Message: "bad request",
	})
}

// Unauthorized returns status 401 with a message.
func Unauthorized(c echo.Context, msg string) error {
	return c.JSON(http.StatusUnauthorized, DataResponse{
		Code:    http.StatusUnauthorized,
		Message: msg,
	})
}

// Forbidden returns status 403 with a message.
func Forbidden(c echo.Context, msg string) error {
	return c.JSON(http.StatusForbidden, DataResponse{
		Code:    http.StatusForbidden,
		Message: msg,
	})
}

// NotFound returns status 404 with a message.
func NotFound(c echo.Context, msg string) error {
	return c.JSON(http.StatusNotFound, DataResponse{
		Code:    http.StatusNotFound,
		Message: msg,
	})
}

// Conflict returns status 409 with a message.
func Conflict(c echo.Context, msg string) error {
	return c.JSON(http.StatusConflict, DataResponse{
		Code:    http.StatusConflict,
		Message: msg,
	})
}

// InternalError returns status 500 with a message.
func InternalError(c echo.Context, msg string) error {
	return c.JSON(http.StatusInternalServerError, DataResponse{
		Code:    http.StatusInternalServerError,
		Message: msg,
	})
}

func InternalErrorDefaultMsg(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, DataResponse{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	})
}
