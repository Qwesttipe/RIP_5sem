package handler

import (
	"RIP/internal/app/ds"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// POST /orders/delete/:id - пометить заказ как удалённый
func (h *Handler) DeleteLoadsOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.SetOrderStatus(id, "удален"); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// После удаления можно редиректить на главную
	ctx.Redirect(http.StatusSeeOther, "/")
}

// GET /api/orders/draft/cart
func (h *Handler) GetDraftCartAPI(ctx *gin.Context) {
	// Пока без авторизации — используем userID = 1
	userID := 1

	// Ищем черновик
	order, err := h.Repository.GetDraftOrder(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if order == nil {
		// Если черновика нет — возвращаем пустую корзину
		ctx.JSON(http.StatusOK, gin.H{
			"status":    "success",
			"orderID":   0,
			"itemCount": 0,
		})
		return
	}

	// Считаем количество услуг
	count, err := h.Repository.GetOrderLoadsCount(order.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"orderID":   order.ID,
		"itemCount": count,
	})
}

func (h *Handler) GetOrdersAPI(ctx *gin.Context) {
	start := ctx.Query("start")
	end := ctx.Query("end")

	fixedStatuses := ctx.Query("status")
	if fixedStatuses == "" {
		fixedStatuses = "завершен,отклонен,отменен"
	}

	orders, err := h.Repository.GetOrdersFiltered(fixedStatuses, start, end)
	if err != nil {
		if err.Error() == "not_found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "заказы со статусом 'черновик' или 'удален' не доступны",
			})
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"orders": orders,
	})
}

func (h *Handler) GetOrderWithLoadsAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	order, loads, err := h.Repository.GetOrderByID(orderID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Формируем DTO материалов
	var loadsDTO []ds.LoadInOrder
	for _, mmo := range loads {
		var wl *float64
		if mmo.RPS.Valid {
			wl = &mmo.RPS.Float64
		}
		loadsDTO = append(loadsDTO, ds.LoadInOrder{
			ID:          mmo.LoadID,
			Title:       mmo.Load.Title,
			Consumption: mmo.Load.Consumption,
			Count:       mmo.Load.Count,
			Image:       mmo.Load.Image,
			RPS:         wl,
		})
	}

	// DTO заказа
	var iops, dbcache *float64
	if order.IOPS.Valid {
		iops = &order.IOPS.Float64
	}
	if order.DBcache.Valid {
		dbcache = &order.DBcache.Float64
	}

	var dateFinish *time.Time
	if order.DateFinish.Valid {
		dateFinish = &order.DateFinish.Time
	}

	resp := ds.OrderWithLoads{
		ID:         order.ID,
		CreatorID:  order.CreatorID,
		Status:     order.RequestStatus,
		IOPS:       iops,
		DBcache:    dbcache,
		DateCreate: order.DateCreate,
		DateForm:   &order.DateForm,
		DateFinish: dateFinish,
		Loads:      loadsDTO,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"order":  resp,
	})
}
func (h *Handler) UpdateLoadOrderAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Читаем тело запроса как map
	var raw map[string]interface{}
	if err := ctx.BindJSON(&raw); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Разрешённые поля
	allowed := map[string]bool{
		"iops":    true,
		"dbcache": true,
	}

	// Проверяем лишние поля
	for k := range raw {
		if !allowed[k] {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("недопустимое поле: %s", k))
			return
		}
	}

	// Преобразуем map в DTO
	var req ds.UpdateOrderRequest
	if v, ok := raw["iops"]; ok {
		if f, ok := v.(float64); ok {
			req.IOPS = &f
		}
	}
	if v, ok := raw["dbcache"]; ok {
		if f, ok := v.(float64); ok {
			req.DBcache = &f
		}
	}

	if err := h.Repository.UpdateLoadOrder(id, req); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	order, _, err := h.Repository.GetOrderByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"order":  order,
	})
}

func (h *Handler) FormLoadOrderAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.FormLoadOrder(orderID); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Возвращаем обновлённый заказ
	order, _, err := h.Repository.GetOrderByID(orderID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"order":  order,
	})
}

func (h *Handler) CompleteOrRejectOrderAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req ds.CompleteOrderRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.CompleteOrRejectOrder(orderID, req); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	order, loads, err := h.Repository.GetOrderByID(orderID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"order":  order,
		"loads":  loads,
	})
}

func (h *Handler) DeleteLoadsOrderAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID заказа"))
		return
	}

	// Проставляем статус "удален" и дату завершения
	if err := h.Repository.SoftDeleteOrder(orderID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"orderID": orderID,
	})
}
