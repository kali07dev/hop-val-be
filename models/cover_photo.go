package models

type CoverPhoto struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Url         string `json:"url"`
	Description string `json:"description"`

	PropertyID uint `json:"-"` 
}