package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/merisho/binaryx-test/activerecord"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) getRequestUser(c *gin.Context) *activerecord.User {
	userRaw, ok := c.Get("user")
	if !ok {
		log.Error("CRITICAL: no user in the context. Call authMiddleware before getting the user from context")
		return nil
	}

	user, ok := userRaw.(*activerecord.User)
	if !ok {
		log.Error("CRITICAL: user in context is not of *activerecord.User")
		return nil
	}

	return user
}

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

	c.Next()
}

func (s *Server) token(ctx *gin.Context) {
	var req TokenRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	user, err := s.activeRecords.User().FindByEmail(ctx, req.Email)
	if err != nil {
		switch err.(type) {
		case activerecord.NotFoundError:
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		default:
			log.WithError(err).Error("could not find user by email")
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}

		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password()), []byte(req.Password))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid password"})
		return
	}

	expires := time.Now().Add(s.config.TokenTTLSeconds * time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		Id:        user.ID().String(),
		ExpiresAt: expires.Unix(),
	})

	signed, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		log.WithError(err).Error("error secret signing jwt token")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	res := TokenResponse{
		Token:     signed,
		ExpiresAt: expires.Unix(),
	}
	ctx.JSON(http.StatusOK, res)
}


func (s *Server) iam(ctx *gin.Context) {
	user := s.getRequestUser(ctx)
	if user == nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatusJSON(http.StatusOK, IAmResponse{
		ID:        user.ID().String(),
		Email:     user.Email(),
		FirstName: user.FirstName(),
		LastName:  user.LastName(),
	})
}
