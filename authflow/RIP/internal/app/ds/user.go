package ds

import (
	"RIP/internal/app/role"
)

type User struct {
	ID       int       `gorm:"primaryKey;autoIncrement"` // обязательный PK для GORM
	Login    string    `gorm:"unique;not null"`
	Role     role.Role `gorm:"type:int"` // хранить enum как int
	Password string
}

type RegisterReq struct {
	Login    string    `json:"login"`
	Password string    `json:"password"`
	Role     role.Role `json:"role,omitempty"` // опционально
}

// Структура ответа
type RegisterResp struct {
	Ok   bool      `json:"ok"`
	Role role.Role `json:"role,omitempty"` // добавляем поле роли
}
