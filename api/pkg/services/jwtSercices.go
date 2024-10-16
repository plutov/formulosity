package services

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/plutov/formulosity/api/pkg/types"
)

type JwtService interface {
	GenerateToken(user *types.User) (string, error)
	ValidateToken(tokenString string) (*types.User, error)
}

type jwtService struct {
	secretKey []byte
	Services
}

func NewJWTService(secretKey string, svc Services) JwtService {
	return &jwtService{
		secretKey: []byte(secretKey),
		Services:  svc,
	}
}

func (j *jwtService) GenerateToken(user *types.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.Id
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	return token.SignedString(j.secretKey)
}

func (j *jwtService) ValidateToken(tokenString string) (*types.User, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		Id := int64(claims["user_id"].(float64))
		return j.Services.Storage.GetUserById(Id)
	}
	return nil, jwt.ErrSignatureInvalid
}
