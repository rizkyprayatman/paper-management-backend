package models

import (
	"github.com/google/uuid"
)

type User struct {
	UUID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"uuid"`
	Email         string    `gorm:"unique" json:"email"`
	Name          string    `json:"name"`
	PhoneNumber   string    `json:"phone_number"`
	Password      string    `json:"password,omitempty"`
	RoleID        uint      `json:"role_id"`
	Role          Role      `gorm:"foreignKey:RoleID" json:"role"`
	OAuthProvider string    `json:"oauth_provider,omitempty"`
	OAuthID       string    `json:"oauth_id,omitempty"`
}

type Role struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique" json:"name"`
}
