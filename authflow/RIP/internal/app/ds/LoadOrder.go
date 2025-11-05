package ds

import (
	"database/sql"
	"time"
)

type LoadOrder struct {
	ID            int             `gorm:"primaryKey;autoIncrement"`             // первичный ключ
	IOPS          sql.NullFloat64 `gorm:"type:NUMERIC(50,2);default:null"`      // количество i/o запросов в сек
	DBcache       sql.NullFloat64 `gorm:"type:NUMERIC(5,2);default:null"`       // кэш бд
	RAM_OS        sql.NullFloat64 `gorm:"type:NUMERIC(5,2);default:null"`       // оперативная память сервера
	RPS           sql.NullFloat64 `gorm:"type:NUMERIC(50,2);default:null"`      // rps
	CPUload       sql.NullFloat64 `gorm:"type:NUMERIC(5,2);default:null"`       // целевая загрузка цпу
	ARPT          sql.NullFloat64 `gorm:"type:NUMERIC(50,2);default:null"`      // среднее время обработки запроса
	CreatorID     int             `gorm:"not null"`                             // ID создателя заказа
	ModeratorID   *int            `gorm:""`                                     // ID модератора заказа
	RequestStatus string          `gorm:"type:varchar(50)"`                     // статус заказа
	DateCreate    time.Time       `gorm:"not null;autoCreateTime"`              // дата создания
	DateForm      time.Time       `gorm:"default:null"`                         // дата последнего обновления
	DateFinish    sql.NullTime    `gorm:"default:null"`                         // дата завершения (может быть null)
	Creator       User            `gorm:"foreignKey:CreatorID;references:ID"`   // связь с пользователем-автором
	Moderator     User            `gorm:"foreignKey:ModeratorID;references:ID"` // связь с пользователем-модератором
}

type OrderResponse struct {
	ID         int        `json:"id"`
	Status     string     `json:"status"`
	DateCreate time.Time  `json:"date_create"`
	DateForm   *time.Time `json:"date_form,omitempty"`
	DateFinish *time.Time `json:"date_finish,omitempty"`
}

type OrdersListResponse struct {
	Status string          `json:"status" example:"success"`
	Orders []OrderResponse `json:"orders"`
}

type UpdateOrderRequest struct {
	IOPS    *float64 `json:"iops,omitempty"`
	DBcache *float64 `json:"dbcache,omitempty"`
	RAM_OS  *float64 `json:"ramos,omitempty"`
	RPS     *float64 `json:"rps,omitempty"`
	CPUload *float64 `json:"cpuload,omitempty"`
	ARPT    *float64 `json:"arpt,omitempty"`
}

type OrderWithLoads struct {
	ID         int           `json:"id"`
	CreatorID  int           `json:"creator_id"`
	Status     string        `json:"status"`
	IOPS       *float64      `json:"iops,omitempty"`
	DBcache    *float64      `json:"dbcache,omitempty"`
	RAM_OS     *float64      `json:"ramos,omitempty"`
	RPS        *float64      `json:"rps,omitempty"`
	CPUload    *float64      `json:"cpuload,omitempty"`
	ARPT       *float64      `json:"arpt,omitempty"`
	DateCreate time.Time     `json:"date_create"`
	DateForm   *time.Time    `json:"date_form,omitempty"`
	DateFinish *time.Time    `json:"date_finish,omitempty"`
	Loads      []LoadInOrder `json:"loads"`
}

type CompleteOrderRequest struct {
	Status      string `json:"status" binding:"required,oneof=завершен отклонен"`
	ModeratorID int    `json:"moderator_id" binding:"required"`
}
