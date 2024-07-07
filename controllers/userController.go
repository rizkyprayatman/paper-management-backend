package controllers

import (
	"paper-management-backend/database"
	"paper-management-backend/helpers"
	"paper-management-backend/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	var existingUser models.User
	if err := database.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email already registered"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot hash password"})
	}
	user.Password = string(hashedPassword)

	var role models.Role
	if err := database.DB.Where("name = ?", "user").First(&role).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Role not found"})
	}
	user.RoleID = role.ID

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot create user"})
	}

	response := struct {
		Message string `json:"message"`
		Email   string `json:"email"`
	}{
		Message: "Registration successful",
		Email:   user.Email,
	}

	return c.JSON(response)
}

func Login(c *fiber.Ctx) error {
	// Mendapatkan kredensial pengguna dari permintaan
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&credentials); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Mencari pengguna berdasarkan alamat email
	var user models.User
	if err := database.DB.Where("email = ?", credentials.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Memverifikasi kata sandi
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Generate JWT token
	token, err := helpers.GenerateJWT(user.UUID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot generate token"})
	}

	// Kirim respons JSON dengan name, email, phone number dan token JWT;
	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"name":        user.Name,
			"email":       user.Email,
			"phoneNumber": user.PhoneNumber,
		},
	})
}
