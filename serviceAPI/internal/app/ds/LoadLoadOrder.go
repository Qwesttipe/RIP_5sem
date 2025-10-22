package ds

import "database/sql"

// LoadLoadOrder представляет связь "нагрузка ↔ заказ" (многие ко многим)
type LoadLoadOrder struct {
	ID          int             `gorm:"primaryKey"`
	LoadID      int             `gorm:"not null"`
	LoadOrderID int             `gorm:"not null"`
	CPU         sql.NullFloat64 `json:"cpu"`
	RAM         sql.NullFloat64 `json:"ram"`

	Load      Load      `gorm:"foreignKey:LoadID;references:ID"`
	LoadOrder LoadOrder `gorm:"foreignKey:LoadOrderID;references:ID"`
}
