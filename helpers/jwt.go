package helpers

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret []byte

func GenerateJWT(id string) (string, error) {
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	// Buat klaim token
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"idle": id,
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // Token berlaku selama 24 jam
	})

	// Tanda tangani token dengan kunci rahasia yang hanya diketahui oleh server
	tokenString, err := claims.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func HandleJwtToken(c *fiber.Ctx) (uuid.UUID, error) {
	errResponse := errors.New("unauthorized")

	// Parse JWT token from Authorization header
	headerToken := c.Get("Authorization")
	if headerToken == "" || !strings.HasPrefix(headerToken, "Bearer ") {
		// log.Println("Authorization header is missing or does not start with 'Bearer '")
		return uuid.Nil, errResponse
	}
	// log.Println("Authorization header found:", headerToken)

	// Extract JWT token from "Bearer <token>"
	stringToken := strings.Split(headerToken, " ")[1]
	// log.Println("Extracted JWT token:", stringToken)

	// Parse JWT token
	token, err := jwt.ParseWithClaims(stringToken, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Println("Invalid signing method")
			return nil, errResponse
		}
		// log.Println("Signing method is valid")
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		// log.Println("Error parsing JWT token:", err)
		return uuid.Nil, errResponse
	}
	if !token.Valid {
		// log.Println("JWT token is not valid")
		return uuid.Nil, errResponse
	}
	// log.Println("JWT token is valid")

	// Extract user ID from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		// log.Println("Failed to extract claims from JWT token")
		return uuid.Nil, errResponse
	}
	// log.Println("Extracted claims:", claims)

	if userID, ok := claims["idle"].(string); ok {
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			// log.Println("Failed to parse userID:", err)
			return uuid.Nil, errResponse
		}
		// log.Println("Parsed userID:", parsedUserID)
		return parsedUserID, nil
	}

	// log.Println("userID not found in claims")
	return uuid.Nil, errResponse
}
