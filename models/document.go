package models

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID              uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID          uuid.UUID     `json:"user_id"`
	User            User          `gorm:"foreignKey:UserID" json:"user"`
	JenisDocumentID uint          `json:"jenis_document_id"`
	JenisDocument   JenisDocument `gorm:"foreignKey:JenisDocumentID" json:"jenis_document"`
	CountPage       int           `json:"count_page"`
	FileName        string        `json:"file_name"`
	URLFile         string        `json:"url_file"`
	Key             string        `json:"key"`
	IsUsed          bool          `json:"is_used"`
	Location        string        `json:"location"`
	QrInformation   string        `json:"qr_location"`
	ThumbnailURL    string        `json:"thumbnail_url"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	DeletedAt       *time.Time    `json:"deleted_at"`
}

type JenisDocument struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"unique" json:"name"`
	Description string `json:"description"`
}
