package models

import "time"

type Folder struct {
	ID         string
	UserID     string
	Path       string
	Name       string
	ParentPath string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
