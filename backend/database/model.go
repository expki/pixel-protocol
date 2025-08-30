package database

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `gorm:"primarykey"`
	UserName string    `gorm:"uniqueIndex:uq_user_username;not null"`
}
