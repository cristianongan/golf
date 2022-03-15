package auth

import (
	"start/models"

	"github.com/dgrijalva/jwt-go"
)

func CreateToken(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)
	return token.SignedString([]byte(secret))
}

func VerifyCmsJwtToken(tokenString, secretKey string) (models.CmsUserProfile, error) {
	keyFn := func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	}

	profile := models.CmsUserProfile{}

	token, err := jwt.ParseWithClaims(tokenString, &profile, keyFn)

	if err == nil && token.Valid {
		return profile, nil
	}

	return profile, err
}
