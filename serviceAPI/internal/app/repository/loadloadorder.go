package repository

import (
	"RIP/internal/app/ds"
)

func (r *Repository) DeleteLoadFromOrder(loadID, orderID int) error {
	return r.db.
		Where("load_id = ? AND load_order_id = ?", loadID, orderID).
		Delete(&ds.LoadLoadOrder{}).Error
}

func (r *Repository) UpdateRPS(loadID, orderID int, RPS float64) error {
	return r.db.
		Model(&ds.LoadLoadOrder{}).
		Where("load_id = ? AND load_order_id = ?", loadID, orderID).
		Update("RPS", RPS).Error
}
