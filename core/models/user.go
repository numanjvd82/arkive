package models

import "time"

type User struct {
	ID                          string
	BrandName                   string
	Email                       string
	VaultSalt                   []byte
	EncryptedMasterKey          []byte
	EncryptedMasterKeyRecovery  []byte
	UsedBytes                   int64
	ReservedBytes               int64
	LastLoginAt                 *time.Time
	RecoverySetupToken          *string
	RecoverySetupTokenExpiresAt *time.Time
	PasswordResetTokenHash      *string
	PasswordResetTokenExpiresAt *time.Time
	PasswordResetConsumedAt     *time.Time
	UpdatedAt                   time.Time
	CreatedAt                   time.Time
}
