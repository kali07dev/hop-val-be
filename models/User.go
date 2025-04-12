package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                      uint           `gorm:"primaryKey" json:"id"`
	Name                    string         `json:"name"`
	Email                   string         `gorm:"uniqueIndex" json:"email"`
	Status                  string         `json:"status"`
	EmailVerifiedAt         *time.Time     `json:"email_verified_at"` 
	Phone                   string         `json:"phone"`
	PasswordLastUpdatedAt   *time.Time     `json:"password_last_updated_at"`
	FinancialInstitutionID  *uint          `json:"financial_institution_id"` 
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`           
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	Role                    string         `json:"role"`
	SignatureStorageURL     *string        `json:"signature_storage_url"`
	ProfileImageStorageURL  *string        `json:"profile_image_storage_url"`


}
