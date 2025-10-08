package repository

import (
	"RIP/internal/app/ds"
	"fmt"
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
