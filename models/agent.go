package models

import "time"

type Agent struct {
	ID                uint    `gorm:"primaryKey" json:"id"`
	UserID            uint    `json:"user_id"` // Foreign Key to User
	Phone1            string  `json:"phone_1"`
	Phone2            *string `json:"phone_2"` // Nullable
	Headline1         *string `json:"headline1"`
	Headline2         *string `json:"headline2"`
	About             *string `json:"about"`
	IsAgreementSigned string  `json:"is_agreement_signed"` 
	AgentType         string  `json:"agent_type"`
	BankName          *string `json:"bank_name"`
	AccountName       *string `json:"account_name"`
	AccountNumber     *string `json:"account_number"`
	AccountType       *string `json:"account_type"`
	AccountBranch     *string `json:"account_branch"`
	Linkedin     *string   `json:"linkedin"`
	Address      *string   `json:"address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CoverageArea *string   `json:"coverage_area"`

	// Relationships
	User User `gorm:"foreignKey:UserID"` // Belongs To User

	// One Agent can have many Properties
	Properties []Property `gorm:"foreignKey:AgentID"`

}
