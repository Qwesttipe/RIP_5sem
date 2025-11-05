package repository

import (
	"RIP/internal/app/ds"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func (r *Repository) GetLoads() ([]ds.Load, error) {
	var loads []ds.Load
	result := r.db.Where("visability = ?", true).Order("id ASC").Find(&loads)
	if result.Error != nil {
		return nil, result.Error
	}
	return loads, nil
}

func (r *Repository) GetLoadsByTitle(title string) ([]ds.Load, error) {
	var loads []ds.Load
	result := r.db.Where("visability = ? AND LOWER(title) LIKE ?", true, "%"+strings.ToLower(title)+"%").Find(&loads)
	if result.Error != nil {
		return nil, result.Error
	}
	return loads, nil
}

// Получаем черновой заказ пользователя
func (r *Repository) GetDraftOrder(ctx context.Context, userID int) (*ds.LoadOrder, error) {
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
func (r *Repository) CreateDraftOrder(ctx context.Context, userID int) (*ds.LoadOrder, error) {
	order := ds.LoadOrder{
		CreatorID:     userID,
		RequestStatus: "черновик",
		ModeratorID:   nil, // черновик создаётся без модератора
	}

	if err := r.db.Create(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

// Добавляем нагрузку в заказ, если его там нет
func (r *Repository) AddLoadToOrder(orderID int, loadID int) error {
	var count int64
	err := r.db.Model(&ds.LoadLoadOrder{}).Where("load_order_id = ? AND load_id = ?", orderID, loadID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // нагрузка уже добавлена
	}

	item := ds.LoadLoadOrder{
		LoadID:      loadID,
		LoadOrderID: orderID,
	}
	return r.db.Create(&item).Error
}

// Получаем количество нагрузок в заказе
func (r *Repository) GetOrderLoadsCount(orderID int) (int64, error) {
	var count int64
	err := r.db.Model(&ds.LoadLoadOrder{}).Where("load_order_id = ?", orderID).Count(&count).Error
	return count, err
}

// SetOrderStatus обновляет статус заказа по его ID
func (r *Repository) SetOrderStatus(orderID int, status string) error {
	return r.db.Model(&ds.LoadOrder{}).Where("id = ?", orderID).Update("request_status", status).Error
}

// Получаем одну нагрузку по ID
func (r *Repository) GetLoadByID(id int) (*ds.Load, error) {
	var load ds.Load
	err := r.db.First(&load, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &load, nil
}

// Получаем список нагрузок с опциональной фильтрацией по названию
func (r *Repository) GetLoadsFiltered(title string) ([]ds.Load, error) {
	var loads []ds.Load
	query := r.db.Model(&ds.Load{}).Where("visability = ?", true)
	if title != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%")
	}
	err := query.Find(&loads).Error
	if err != nil {
		return nil, err
	}
	return loads, nil
}

// Создаём новую нагрузку
func (r *Repository) CreateLoad(load *ds.Load) error {
	return r.db.Create(load).Error
}

// Обновляем нагрузку по ID
func (r *Repository) UpdateLoad(id int, updated *ds.Load) error {
	var load ds.Load
	if err := r.db.First(&load, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // нагрузка не найдена
		}
		return err
	}

	// Обновляем все поля
	return r.db.Model(&load).Updates(updated).Error
}

// Логическое удаление нагрузки
func (r *Repository) DeleteLoadLogical(id int) error {
	result := r.db.Model(&ds.Load{}).
		Where("id = ?", id).
		Update("visability", false)
	return result.Error
}

// UploadLoadImage загружает новое изображение нагрузки в MinIO, удаляет старое, обновляет image_url
func (r *Repository) UploadLoadImage(id int, fileHeader *multipart.FileHeader) error {
	// Получаем нагрузку из БД
	var load ds.Load
	if err := r.db.First(&load, id).Error; err != nil {
		return err
	}

	// Удаляем старое изображение из MinIO, если есть
	if load.Image != "" {
		parts := strings.Split(load.Image, "/")
		objectName := parts[len(parts)-1]
		_ = r.minioClient.RemoveObject(context.Background(), r.bucketName, objectName, minio.RemoveObjectOptions{})
	}
	// Открываем новый файл
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// Расширение файла
	ext := filepath.Ext(fileHeader.Filename)
	base := strings.TrimSuffix(fileHeader.Filename, ext)

	// Переводим в латиницу
	latinBase := toLatin(base)

	// Генерация имени файла
	objectName := fmt.Sprintf("load-%s%s", latinBase, ext)

	// Загружаем в MinIO
	_, err = r.minioClient.PutObject(
		context.Background(),
		r.bucketName,
		objectName,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{ContentType: fileHeader.Header.Get("Content-Type")},
	)
	if err != nil {
		return err
	}

	// Формируем URL
	imageURL := fmt.Sprintf("http://%s/%s/%s", r.minioClient.EndpointURL().Host, r.bucketName, objectName)

	// Обновляем поле ImageURL в БД
	return r.db.Model(&ds.Load{}).Where("id = ?", id).Update("image", imageURL).Error
}

// toLatin переводит строку в латиницу, оставляет только ASCII буквы и цифры
func toLatin(s string) string {
	var out strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) && r <= unicode.MaxASCII {
			out.WriteRune(unicode.ToLower(r))
		} else if unicode.IsDigit(r) {
			out.WriteRune(r)
		}
	}
	return out.String()
}
