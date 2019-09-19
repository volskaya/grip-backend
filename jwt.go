package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

type JwtProvider struct {
	secret string
}

func (self *JwtProvider) createToken(auth *DiscordAuth) (*string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		JwtClaims{
			auth.user.Avatar,
			auth.user.Discriminator,
			auth.user.Username,
			jwt.StandardClaims{
				Id:        auth.user.ID,
				ExpiresAt: auth.token.Expiry.Unix(),
			},
		},
	)

	signed, err := token.SignedString([]byte(self.secret))
	return &signed, err
}

func (self *JwtProvider) validateToken(token *string) (*jwt.Token, error) {
	tokenParsed, err := jwt.Parse(
		*token,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected sign method: %s\n", token.Header["alg"])
			}

			return []byte(self.secret), nil
		},
	)

	return tokenParsed, err
}

func (self *JwtProvider) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			authorized bool
			id         string
			ctx        = r.Context()
		)

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		// TODO: Handle expiration
		if len(auth) == 2 || auth[0] == "Bearer" {
			if token, err := self.validateToken(&auth[1]); err == nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					if userID, ok := claims["jti"].(string); ok {
						authorized = true
						id = userID
					}
				} else {
					log.Printf("Could not parse claims: %s\n", claims)
				}
			} else {
				log.Println(err)
			}
		}

		ctx = context.WithValue(ctx, "authorized", authorized)
		ctx = context.WithValue(ctx, "user_id", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
