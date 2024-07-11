package middleware

import (
	"errors"
	"os"
	"strings"

	"paper-management-backend/database"
	"paper-management-backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func Authorization(c *fiber.Ctx) error {
	errResponse := errors.New("unauthorized")

	// log.Println("Authorization middleware invoked")

	// Parse JWT token from Authorization header
	headerToken := c.Get("Authorization")
	if headerToken == "" || !strings.HasPrefix(headerToken, "Bearer ") {
		// log.Println("Authorization header missing or invalid")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// log.Println("Authorization header found:", headerToken)

	// Extract JWT token from "Bearer <token>"
	stringToken := strings.Split(headerToken, " ")[1]
	// log.Println("Extracted JWT token:", stringToken)

	// Parse JWT token
	token, err := jwt.ParseWithClaims(stringToken, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			// log.Println("Unexpected signing method:", t.Header["alg"])
			return nil, errResponse
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		// log.Println("Failed to parse JWT token:", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if !token.Valid {
		// log.Println("Invalid JWT token")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Extract user ID from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		// log.Println("Failed to extract claims from JWT token")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Ensure userId is a valid UUID
	if userID, ok := claims["idle"].(string); ok {
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			// log.Println("Invalid user ID in JWT claims:", userID)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		var user models.User
		if err := database.DB.First(&user, "uuid = ?", parsedUserID).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		// Store user information in locals
		c.Locals("userId", parsedUserID)
		c.Locals("userName", user.Name)
		c.Locals("userEmail", user.Email)
		c.Locals("userPhone", user.PhoneNumber)

		// c.Locals("userId", parsedUserID)
		// log.Println("User ID extracted and set in locals:", parsedUserID)
		return c.Next()
	}

	// log.Println("User ID not found in JWT claims")
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
}
