package tocken

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretSize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {

	if len(secretKey) < minSecretSize {
		return nil, fmt.Errorf("invalid length")
	}

	return &JWTMaker{secretKey}, nil
}

func (maker *JWTMaker) CreateTocken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

func (maker *JWTMaker) VeifyTocken(token string) (*Payload, error) {
	Keyfunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("Invalid Tocken")
		}
		return []byte(maker.secretKey), nil
	}

	jwtTocken, err := jwt.ParseWithClaims(token, &Payload{}, Keyfunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	payload, ok := jwtTocken.Claims.(*Payload)
	if !ok {
		return nil, err
	}

	return payload, nil
}
