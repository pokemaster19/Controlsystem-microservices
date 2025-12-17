package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(user models.User, cfg *config.Config) (string, error) {

	tokenLifespan, err := strconv.Atoi(strings.TrimSpace(cfg.TokenMinuteLifespan))

	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["id"] = user.ID
	claims["roles"] = user.Roles
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(tokenLifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(cfg.TokenSecret))

}

func GenerateRefreshToken(user models.User, cfg *config.Config) (string, error) {
	tokenLifespan, err := strconv.Atoi(strings.TrimSpace(cfg.RefreshTokenHourLifespan))
	if err != nil {
		tokenLifespan = 72
		fmt.Println("Warning: REFRESH_TOKEN_HOUR_LIFESPAN not set or invalid. Using default 72 hours.")
	}

	claims := jwt.MapClaims{}
	claims["id"] = user.ID
	claims["roles"] = user.Roles
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(tokenLifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.RefreshTokenSecret))
}

func ValidateToken(c *gin.Context, cfg *config.Config) error {
	token, err := GetToken(c, cfg)

	if err != nil {
		return err
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return nil
	}

	return errors.New("invalid token provided")
}

func GetToken(c *gin.Context, cfg *config.Config) (*jwt.Token, error) {
	tokenString := getTokenFromRequest(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(cfg.TokenSecret), nil
	})
	return token, err
}

func getTokenFromRequest(c *gin.Context) string {
	bearerToken := c.Request.Header.Get("Authorization")

	splitToken := strings.Split(bearerToken, " ")
	if len(splitToken) == 2 {
		return splitToken[1]
	}
	return ""
}

func ParseToken(tokenString string, cfg *config.Config) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.TokenSecret), nil
	})
}
