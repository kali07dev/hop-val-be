package models

import "time"

type Location struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PropertyID    uint      `gorm:"uniqueIndex" json:"property_id"` // Foreign Key & Unique ensures One-to-One
	Region        string    `json:"region"`
	District      string    `json:"district"`
	Area          string    `json:"area"`
	Postcode      *string   `json:"postcode"`
	SubArea       *string   `json:"sub_area"`
	GoogleMapLink *string   `json:"google_map_link"`
	Latitude      *string   `json:"latitude"`
	Longitude     *string   `json:"longitude"`
	ZoneCategory  string    `json:"zone_category"`
	Zoning        string    `json:"zoning"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
