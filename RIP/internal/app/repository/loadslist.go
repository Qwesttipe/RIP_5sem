package repository

import (
	"RIP/internal/app/ds"
	"errors"
	"strings"

	"gorm.io/gorm"
)

func (r *Repository) GetLoads() ([]ds.Load, error) {
	var loads []ds.Load
	result := r.db.Find(&loads)
	if result.Error != nil {
		return nil, result.Error
	}
	return loads, nil
}

func (r *Repository) GetLoadsByTitle(title string) ([]ds.Load, error) {
	var loads []ds.Load
	result := r.db.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%").Find(&loads)
	if result.Error != nil {
		return nil, result.Error
	}
	return loads, nil
}

// Получаем черновой заказ пользователя
func (r *Repository) GetDraftOrder(userID int) (*ds.LoadOrder, error) {
	var order ds.LoadOrder
	err := r.db.Where("creator_id = ? AND request_status = ?", userID, "черновик").First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // черновик отсутствует
		}
		return nil, err
	}
	return &order, nil
}

// Создаём новый черновой заказ
func (r *Repository) CreateDraftOrder(userID int) (*ds.LoadOrder, error) {
	order := ds.LoadOrder{
		CreatorID:     userID,
		ModeratorID:   userID, // можно назначить себя модератором, либо 0/NULL
		RequestStatus: "черновик",
	}
	if err := r.db.Create(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// Добавляем материал в заказ, если его там нет
func (r *Repository) AddLoadToOrder(orderID int, loadID int) error {
	var count int64
	err := r.db.Model(&ds.LoadLoadOrder{}).Where("load_order_id = ? AND load_id = ?", orderID, loadID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // материал уже добавлен
	}

	item := ds.LoadLoadOrder{
		LoadID:      loadID,
		LoadOrderID: orderID,
	}
	return r.db.Create(&item).Error
}

// Получаем количество материалов в заказе
func (r *Repository) GetOrderLoadsCount(orderID int) (int64, error) {
	var count int64
	err := r.db.Model(&ds.LoadLoadOrder{}).Where("load_order_id = ?", orderID).Count(&count).Error
	return count, err
}

// SetOrderStatus обновляет статус заказа по его ID
func (r *Repository) SetOrderStatus(orderID int, status string) error {
	return r.db.Model(&ds.LoadOrder{}).Where("id = ?", orderID).Update("request_status", status).Error
}
