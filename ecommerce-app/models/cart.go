package models

import (
    "time"

    "gorm.io/gorm"
)

type Cart struct {
    ID        uint           `gorm:"primaryKey" json:"ID"`
    CreatedAt time.Time      `json:"CreatedAt"`
    UpdatedAt time.Time      `json:"UpdatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"DeletedAt"`
    UserID    uint           `json:"user_id"`
    ProductID uint           `json:"product_id"`
    Quantity  int            `json:"quantity"`
    Product   Product        `gorm:"foreignKey:ProductID" json:"product"`
}
