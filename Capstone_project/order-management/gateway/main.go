package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	r := gin.Default()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change_me"
	}

	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	r.POST("/login", func(c *gin.Context) {
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"role": "user",
		})
		v, err := t.SignedString([]byte(secret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": v})
	})

	r.Any("/api/v1/orders/*path", proxyOrder)
	r.Any("/graphql", proxyOrder)

	log.Println("gateway listening on :8088")
	if err := r.Run(":8088"); err != nil {
		log.Fatal(err)
	}
}

func proxyOrder(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change_me"
	}

	tokenString := strings.TrimPrefix(auth, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	path := c.Request.URL.Path
	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, "http://localhost:8080"+path, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not create upstream request"})
		return
	}
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream unavailable"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not read upstream response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
