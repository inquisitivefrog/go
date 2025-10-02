package models

import (
    "time"
)

type User struct {
    ID        uint       `gorm:"primaryKey"`
    CreatedAt time.Time  `gorm:"type:timestamptz"`
    UpdatedAt time.Time  `gorm:"type:timestamptz"`
    DeletedAt *time.Time `gorm:"type:timestamptz;index"`
    Username  string     `gorm:"unique;not null"`
    Password  string     `gorm:"not null"`
    Email     string     `gorm:"unique;not null"`
    Role      string     `gorm:"not null;default:'user'"`
}
