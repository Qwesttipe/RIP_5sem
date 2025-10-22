package ds

// Load представляет нагрузку
type Load struct {
	ID          uint    `gorm:"primaryKey;autoIncrement"`           // первичный ключ
	Title       string  `gorm:"type:varchar(255)"`                  // название
	Description string  `gorm:"type:varchar(255)"`                  // описание
	Image       string  `gorm:"type:varchar(255)"`                  // ссылка на изображение
	Consumption float64 `gorm:"type:decimal(10,3);not null"`        // расход
	Visability  bool    `gorm:"type:boolean;not null;default:true"` // видимость

}

type LoadInOrder struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Image       string   `json:"image"`
	Consumption *float64 `json:"RPS,omitempty"`
}
