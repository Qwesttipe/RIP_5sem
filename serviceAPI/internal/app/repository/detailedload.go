package repository

import "RIP/internal/app/ds"

func (r *Repository) GetLoad(id int) (ds.Load, error) {
	var load ds.Load
	result := r.db.First(&load, id)
	if result.Error != nil {
		return ds.Load{}, result.Error
	}
	return load, nil
}
