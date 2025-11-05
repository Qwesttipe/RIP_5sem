package repository

import (
	"RIP/internal/app/ds"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *Repository) GetOrderByID(id int) (ds.LoadOrder, []ds.LoadLoadOrder, error) {
	var order ds.LoadOrder
	if err := r.db.First(&order, id).Error; err != nil {
		return ds.LoadOrder{}, nil, fmt.Errorf("заказ с ID=%d не найден", id)
	}

	var mmos []ds.LoadLoadOrder
	if err := r.db.Preload("Load").Where("load_order_id = ?", id).Find(&mmos).Error; err != nil {
		return ds.LoadOrder{}, nil, err
	}

	return order, mmos, nil
}

func (r *Repository) GetOrdersFiltered(status, start, end string) ([]ds.OrderResponse, error) {
	var orders []ds.OrderResponse

	// Разрешённые статусы для выдачи
	allowedStatuses := map[string]bool{
		"сформирован": true,
		"завершен":    true,
		"отклонен":    true,
	}

	query := r.db.
		Table("load_orders mo").
		Select(`mo.id, 
		        mo.request_status as status, 
		        mo.date_create, 
		        mo.date_form, 
		        mo.date_finish, 
		        u1.login as moderator, 
		        u2.login as creator`).
		Joins("LEFT JOIN users u1 ON u1.id = mo.moderator_id").
		Joins("LEFT JOIN users u2 ON u2.id = mo.creator_id")

	// фильтр по статусу
	if status != "" {
		statuses := []string{}
		for _, s := range strings.Split(status, ",") {
			s = strings.TrimSpace(s)
			if allowedStatuses[s] { // оставляем только разрешённые
				statuses = append(statuses, s)
			}
		}
		if len(statuses) == 0 {
			// Если после фильтра ничего не осталось — не выдаём ни один заказ
			return []ds.OrderResponse{}, nil
		}
		query = query.Where("mo.request_status IN ?", statuses)
	} else {
		// Если статус не указан — выдаём все разрешённые статусы
		query = query.Where("mo.request_status IN ?", []string{"сформирован", "завершен", "отклонен"})
	}

	// фильтр по диапазону дат
	if start != "" && end != "" {
		query = query.Where("mo.date_create BETWEEN ? AND ?", start, end)
	}

	if err := query.Scan(&orders).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении заказов: %w", err)
	}

	return orders, nil
}

// GetOrdersFilteredForUser возвращает заказы указанного пользователя (creator_id = userID)
func (r *Repository) GetOrdersFilteredForUser(ctx context.Context, status, start, end string, userID int) ([]ds.OrderResponse, error) {
	var orders []ds.OrderResponse

	// Разрешённые статусы для выдачи
	allowedStatuses := map[string]bool{
		"сформирован": true,
		"завершен":    true,
		"отклонен":    true,
	}

	query := r.db.WithContext(ctx).
		Table("load_orders mo").
		Select(`mo.id, 
				mo.request_status as status, 
				mo.date_create, 
				mo.date_form, 
				mo.date_finish, 
				u1.login as moderator, 
				u2.login as creator`).
		Joins("LEFT JOIN users u1 ON u1.id = mo.moderator_id").
		Joins("LEFT JOIN users u2 ON u2.id = mo.creator_id").
		Where("mo.creator_id = ? OR ? = 2", userID, userID)

	// фильтр по статусу
	if status != "" {
		statuses := []string{}
		for _, s := range strings.Split(status, ",") {
			s = strings.TrimSpace(s)
			if allowedStatuses[s] { // оставляем только разрешённые
				statuses = append(statuses, s)
			}
		}
		if len(statuses) == 0 {
			return []ds.OrderResponse{}, nil
		}
		query = query.Where("mo.request_status IN ?", statuses)
	} else {
		// Если статус не указан — выдаём все разрешённые статусы
		query = query.Where("mo.request_status IN ?", []string{"сформирован", "завершен", "отклонен"})
	}

	// фильтр по диапазону дат
	if start != "" && end != "" {
		query = query.Where("mo.date_create BETWEEN ? AND ?", start, end)
	}

	if err := query.Scan(&orders).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении заказов: %w", err)
	}

	return orders, nil
}

func (r *Repository) UpdateLoadOrder(orderID int, req ds.UpdateOrderRequest) error {
	updates := make(map[string]interface{})

	if req.IOPS != nil {
		updates["iops"] = *req.IOPS
	}
	if req.DBcache != nil {
		updates["dbcache"] = *req.DBcache
	}

	if len(updates) == 0 {
		return nil // ничего менять не нужно
	}

	return r.db.Model(&ds.LoadOrder{}).Where("id = ?", orderID).Updates(updates).Error
}

func (r *Repository) FormLoadOrder(orderID int) error {
	var count int64
	if err := r.db.Model(&ds.LoadLoadOrder{}).
		Where("load_order_id = ? AND rps IS NULL", orderID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("ошибка проверки rps: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("нельзя сформировать заказ: не все поля заполнены")
	}

	// Обновляем заказ: статус и date_form
	updates := map[string]interface{}{
		"request_status": "сформирован",
		"date_form":      time.Now(),
	}

	// Обновляем только если текущий статус черновик
	if err := r.db.Model(&ds.LoadOrder{}).
		Where("id = ? AND request_status = ?", orderID, "черновик").
		Updates(updates).Error; err != nil {
		return fmt.Errorf("ошибка обновления заказа: %w", err)
	}

	return nil
}

func (r *Repository) CompleteOrRejectOrder(orderID int, req ds.CompleteOrderRequest) error {
	var order ds.LoadOrder
	if err := r.db.First(&order, orderID).Error; err != nil {
		return fmt.Errorf("заказ с ID=%d не найден", orderID)
	}

	// Обновляем статус, модератора и дату завершения
	updates := map[string]interface{}{
		"request_status": req.Status,
		"moderator_id":   req.ModeratorID,
		"date_finish":    time.Now(),
	}

	if err := r.db.Model(&ds.LoadOrder{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
		return fmt.Errorf("не удалось обновить заказ: %w", err)
	}

	// Если заказ отклонён — прекращаем выполнение, не считаем расход
	if req.Status == "отклонен" {
		return nil
	}

	// Рассчитываем расход (только если заказ завершён)
	var mmos []ds.LoadLoadOrder
	if err := r.db.Preload("Load").Where("load_order_id = ?", orderID).Find(&mmos).Error; err != nil {
		return err
	}

	for _, mmo := range mmos {
		if mmo.CPU.Valid && order.IOPS.Valid && order.DBcache.Valid {
			mmo.RAM = sql.NullFloat64{
				Float64: order.RAM_OS.Float64 + (order.IOPS.Float64 * order.IOPS.Float64) + order.DBcache.Float64*1.25*mmo.Load.Consumption,
			}

			// нагрузочный расход
			mmo.CPU = sql.NullFloat64{
				Float64: (order.RPS.Float64 / order.IOPS.Float64 * order.ARPT.Float64 / (1 - order.CPUload.Float64)) * mmo.Load.Consumption,
				Valid:   true,
			}

			if err := r.db.Save(&mmo).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Repository) SoftDeleteOrder(orderID int) error {
	updates := map[string]interface{}{
		"request_status": "удален",
		"date_form":      time.Now(), // дата завершения
	}

	return r.db.Model(&ds.LoadOrder{}).Where("id = ?", orderID).Updates(updates).Error
}
