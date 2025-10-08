package ds

// Load представляет строительный материал
type Load struct {
	ID          uint    `gorm:"primaryKey;autoIncrement"`    // первичный ключ
	Title       string  `gorm:"type:varchar(255)"`           // название
	Description string  `gorm:"type:varchar(255)"`           // описание
	Image       string  `gorm:"type:varchar(255)"`           // ссылка на изображение
	Consumption float64 `gorm:"type:decimal(10,3);not null"` // расход (м³ на 1 м³ кладки)
	Count       int     `gorm:"type:int;not null"`           // количество (шт./м³)
	//MainLoad     string  `gorm:"type:varchar(255)"`                  // основной материал
	//CountPerM2   int     `gorm:"type:int"`                           // шт. на м²
	//CountPerM3   int     `gorm:"type:int"`                           // шт. на м³
	//NetWeight    float64 `gorm:"type:decimal(10,2)"`                 // масса (кг)
	//LengthMM     int     `gorm:"type:int"`                           // длина (мм)
	//HeightMM     int     `gorm:"type:int"`                           // высота (мм)
	//WidthMM      int     `gorm:"type:int"`                           // ширина (мм)
	//Country      string  `gorm:"type:varchar(100)"`                  // страна
	Visability bool `gorm:"type:boolean;not null;default:true"` // видимость

}

type LoadInOrder struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Consumption float64  `json:"consumption"`
	Count       int      `json:"count"`
	Image       string   `json:"image"`
	RPS         *float64 `json:"RPS,omitempty"`
}
