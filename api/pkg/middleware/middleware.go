package middleware

import (
	"github.com/google/s2a-go/example/echo"
	"github.com/plutov/formulosity/api/pkg/services"
)

func AuthMiddleware(svc services.JwtService) echo.MiddlewareFunc {}
return func(next echo.HandlerFunc)echo.HandlerFunc{
	return func(c echo.Context)error{
		token:=c.Request().Header.Get("Authorization")
		if token == ""{
			return response.Unauthorized(c, "Missing authorization token")
		}
		user, err:=svc.JWTService.ValidateToken(token)

		if err != nil{
			return response.Unauthorized(c, "Invalid authorization token")
		}
		c.Set("user", user)
		return next(c)
	}
}
