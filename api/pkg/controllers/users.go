package controllers

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/plutov/formulosity/api/pkg/http/response"
	"github.com/plutov/formulosity/api/pkg/types"
)

func (h *Handler) registerUser(c echo.Context) error {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`	
		Password string `json:"password"`
	}
	if err := c.Bind(&input); err != nil {
		return response.BadRequestDefaultMessage(c)
	}
	user := &types.User{
		Name: input.name,
		Email:input.email		
	}
	err=user.Password.ValidatePassword(input.password)

	if err != nil{
		return response.BadRequest(err.String())
	}
	err=h.Services.Storage.CreateUser(user)
	if err!= nil{
		switch{
		case errors.Is(err, types.ErrDuplicateEmail):
			return response.BadRequest(types.ErrDuplicateEmail)
		default:
			return response.BadRequest(err.String())
		}
	}
	return response.Created(c,"User Created Successfully",user)
}

func (h *Handler)loginUser(c echo.Context)error{
	var input struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	if err:=c.Bind(&input); err!=nil{
		return response.BadRequestDefaultMessage(c)
	}

	if input.Email == "" || input.Password == ""{
		return response.BadRequest(c,"Email and password are required")
	}

	user, err:= h.Services.Storage.GetUserByEmail(input.Email)

	if err!= nil{
		switch{
		case errors.Is(err, types.ErrRecordNotFound):
			return response.Unauthorized(c, types.ErrRecordNotFound.Error())
		default:
			return response.Unauthorized(c, "invalid Email")
		}
	}
	ok, err:=user.Password.Matches(input.Password)
	if err!= nil{
		response.InternalErrorDefaultMsg(c)
	}

	if !ok{
		response.Unauthorized(c, "Incorrect password")
	}
	token, err:=h.JWTService.GenerateToken(user)
	if err!= nil{
		return response.InternalErrorDefaultMsg(c)
	}

	return response.Ok(c, echo.Map{
		"token":token,
		"user":user,
	})
}