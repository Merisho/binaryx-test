package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/merisho/binaryx-test/activerecord"
	log "github.com/sirupsen/logrus"
)

func (s *Server) authMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, invalidAuthHeader)
		return
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, invalidAuthHeader)
		return
	}

	tokenStr := authHeader[len(prefix):]
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, invalidToken)
		return
	}

	id, err := uuid.Parse(claims.Id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, invalidToken)
		return
	}

	user, err := s.activeRecords.User().FindByID(c, id)
	if err != nil {
		switch err.(type) {
		case activerecord.NotFoundError:
			c.AbortWithStatusJSON(http.StatusUnauthorized, invalidToken)
		default:
			log.WithError(err).Error("could not find user by id")
			c.Abort()
		}

		return
	}

	c.Set("user", user)

	log.Println("=========================================== SETTING USER", user.Email())

	c.Next()
}
