package ds

import (
	"database/sql"
	"time"
)

type LoadOrder struct {
	ID            int             `gorm:"primaryKey;autoIncrement"`             // первичный ключ
	IOPS          sql.NullFloat64 `gorm:"type:NUMERIC(5,2);default:null"`       // высота потолка
	DBcache       sql.NullFloat64 `gorm:"type:NUMERIC(5,2);default:null"`       // толщина стены
	CreatorID     int             `gorm:"not null"`                             // ID создателя заказа
	ModeratorID   int             `gorm:"not null"`                             // ID модератора заказа
	RequestStatus string          `gorm:"type:varchar(50);not null"`            // статус заказа
	DateCreate    time.Time       `gorm:"not null;autoCreateTime"`              // дата создания
	DateForm      time.Time       `gorm:"nullable;autoUpdateTime"`              // дата последнего обновления
	DateFinish    sql.NullTime    `gorm:"default:null"`                         // дата завершения (может быть null)
	Creator       User            `gorm:"foreignKey:CreatorID;references:ID"`   // связь с пользователем-автором
	Moderator     User            `gorm:"foreignKey:ModeratorID;references:ID"` // связь с пользователем-модератором
}
