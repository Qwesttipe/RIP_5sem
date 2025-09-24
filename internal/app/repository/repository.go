package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Load struct {
	ID          int
	CPU         int
	Pattern     string
	Description string
	RAMNumbers  string
	Image       string
}

// Request - общая информация о заявке
type Request struct {
	ID              int
	FindDescription string
	TotalRAM        int
	FinalCPU        int
}

// RequestItem - отдельная нагрузка в заявке
type RequestItem struct {
	ID            int
	RequestID     int
	LoadName      string
	AnomalousRAM  string
	CalculatedCPU int
	Image         string
}

func (r *Repository) GetLoads() ([]Load, error) {
	loads := []Load{
		{
			ID:          1,
			CPU:         20,
			Pattern:     "Равномерная нагрузка на сервер",
			Description: "Расчёт при использовании балансировщиков нагрузки и их штатной работе.",
			Image:       "http://127.0.0.1:3031/browser/img/ravnom-load.jpg",
		},
		{
			ID:          2,
			CPU:         40,
			Pattern:     "Пиковая нагрузка",
			Description: "Расчёт при кратковременном и резком увеличении трафика.",
			Image:       "http://127.0.0.1:3031/browser/img/pik-load.jpg",
		},
		{
			ID:          3,
			CPU:         30,
			Pattern:     "Растущая нагрузка",
			Description: "Расчёт при постепенном и спокойном увеличении затрачиваемых ресурсов сервера.",
			Image:       "http://127.0.0.1:3031/browser/img/rast-load.jpg",
		},
		{
			ID:          4,
			CPU:         100,
			Pattern:     "Нагрузка типа DDoS",
			Description: "Расчёт при DDoS-атаке и перегрузке сервера.",
			Image:       "http://127.0.0.1:3031/browser/img/ddos-load.jpg",
		},
	}

	if len(loads) == 0 {
		return nil, fmt.Errorf("нет данных о нагрузках")
	}

	return loads, nil
}

func (r *Repository) GetLoad(id int) (Load, error) {
	loads, err := r.GetLoads()
	if err != nil {
		return Load{}, err
	}

	for _, load := range loads {
		if load.ID == id {
			return load, nil
		}
	}
	return Load{}, fmt.Errorf("нагрузка не найдена")
}

func (r *Repository) GetLoadsByPattern(pattern string) ([]Load, error) {
	loads, err := r.GetLoads()
	if err != nil {
		return []Load{}, err
	}

	var result []Load
	for _, load := range loads {
		if strings.Contains(strings.ToLower(load.Pattern), strings.ToLower(pattern)) ||
			strings.Contains(strings.ToLower(load.Description), strings.ToLower(pattern)) {
			result = append(result, load)
		}
	}

	return result, nil
}

func (r *Repository) GetRequests() (Request, []RequestItem, error) {
	// Общая информация о заявке
	request := Request{
		ID:              1,
		FindDescription: "Рассчет необходимых мощностей при нагрузке на сервер",
		TotalRAM:        96,
		FinalCPU:        60,
	}

	// Отдельные нагрузки в заявке
	requestItems := []RequestItem{
		{
			ID:            1,
			RequestID:     1,
			LoadName:      "Равномерная нагрузка",
			AnomalousRAM:  "32",
			CalculatedCPU: 20,
			Image:         "http://127.0.0.1:3031/browser/img/ravnom-load.jpg",
		},
		{
			ID:            2,
			RequestID:     1,
			LoadName:      "Пиковая нагрузка",
			AnomalousRAM:  "64",
			CalculatedCPU: 40,
			Image:         "http://127.0.0.1:3031/browser/img/pik-load.jpg",
		},
	}

	if len(requestItems) == 0 {
		return Request{}, nil, fmt.Errorf("нет данных о заявках")
	}

	return request, requestItems, nil
}
