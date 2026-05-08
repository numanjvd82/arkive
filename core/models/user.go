package models

import "time"

type User struct {
	ID                          string
	BrandName                   string
	Email                       string
	VaultSalt                   []byte
	EncryptedMasterKey          []byte
	QuotaBytes                  int64
	UsedBytes                   int64
	ReservedBytes               int64
	LastLoginAt                 *time.Time
	RecoverySetupToken          *string
	RecoverySetupTokenExpiresAt *time.Time
	UpdatedAt                   time.Time
	CreatedAt                   time.Time
}
