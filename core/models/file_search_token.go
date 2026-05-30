package models

type FileSearchToken struct {
	TokenHash []byte
	Field     string
	Weight    int
}
