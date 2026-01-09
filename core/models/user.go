package models

import "time"

type User struct {
	ID              string
	BrandName       string
	Email           string
	QuotaBytes      int64
	UsedBytes       int64
	ReservedBytes   int64
	IsPremium       bool
	IsEmailVerified bool
	IsBanned        bool
	BanReason       *string
	LastLoginAt     *time.Time
	LastActiveAt    time.Time
	LastIP          *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
