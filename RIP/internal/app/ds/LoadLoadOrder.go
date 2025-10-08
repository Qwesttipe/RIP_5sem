package ds

import "database/sql"

// LoadLoadOrder представляет связь "материал ↔ заказ" (многие ко многим)
type LoadLoadOrder struct {
	ID          int `gorm:"primaryKey"`
	LoadID      int `gorm:"not null"`
	LoadOrderID int `gorm:"not null"`
	RPS         sql.NullFloat64

	Load      Load      `gorm:"foreignKey:LoadID;references:ID"`
	LoadOrder LoadOrder `gorm:"foreignKey:LoadOrderID;references:ID"`
}
