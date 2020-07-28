package auth

import (
	"contact-monitoring-ingest-api/internal/device"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type Claims struct {
	Venue string `json:"venue"`
	jwt.StandardClaims
}

func GetDeviceTokenHandler(signingKey string, deviceRepo device.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.Param("id")
		device, err := deviceRepo.Get(deviceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "That device is not verified"})
			return
		}

		expiresAt := time.Now().Add(time.Hour)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
			device.Venue,
			jwt.StandardClaims{
				ExpiresAt: expiresAt.UTC().Unix(),
			},
		})

		ss, err := token.SignedString([]byte(signingKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get signed token"})
			return
		}

		c.JSON(200, gin.H{
			"token":     ss,
			"expiresAt": expiresAt,
		})
	}
}

func DeviceTokenMiddleware(signingKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(c.Request.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(
			authHeader[1],
			&Claims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(signingKey), nil
			},
		)

		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)

		if ok && token.Valid {
			c.Set("claims", claims)
			c.Next()
		} else {
			c.Status(http.StatusUnauthorized)
			c.Abort()
		}
	}
}
