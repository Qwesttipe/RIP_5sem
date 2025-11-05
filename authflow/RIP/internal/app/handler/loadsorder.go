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
	// Берём ID авторизованного пользователя из контекста
	userID, _ := h.getUserFromContext(ctx)

	// Ищем черновик
	order, err := h.Repository.GetDraftOrder(ctx.Request.Context(), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if order == nil {
		// Если черновика нет — возвращаем пустую корзину
		ctx.JSON(http.StatusOK, gin.H{
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
		"orderID":   order.ID,
		"itemCount": count,
	})
}

// GetOrdersAPI godoc
// @Summary      Получить список заказов
// @Description  Возвращает список заказов с возможностью фильтрации по статусу и диапазону дат. Разрешённые статусы: "сформирован", "завершен", "отклонен".
// @Tags		 Заявки с нагрузками
// @Accept       json
// @Produce      json
// @Param        status  query     string  false  "Статус заказа (сформирован, завершен, отклонен), можно указать несколько через запятую"
// @Param        start   query     string  false  "Дата начала фильтрации (формат YYYY-MM-DD)"
// @Param        end     query     string  false  "Дата окончания фильтрации (формат YYYY-MM-DD)"
// @Success      200     {object}  ds.OrdersListResponse  "Успешный ответ со списком заказов"
// @Failure      500     {object}  ds.ErrorResponse       "Ошибка на сервере"
// @Security BearerAuth
// @Router       /api/load_orders [get]
func (h *Handler) GetOrdersAPI(ctx *gin.Context) {
	status := ctx.Query("status")
	start := ctx.Query("start") // формат YYYY-MM-DD
	end := ctx.Query("end")

	// Получаем ID авторизованного пользователя
	userID, _ := h.getUserFromContext(ctx)

	orders, err := h.Repository.GetOrdersFilteredForUser(ctx.Request.Context(), status, start, end, userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
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

	// Проверяем, что запрашиваемый заказ принадлежит текущему пользователю или пользователь — Admin
	userID, roleID := h.getUserFromContext(ctx)
	isAdmin := false
	if roleID == 2 { // role.Admin == 2
		isAdmin = true
	}
	if !isAdmin && order.CreatorID != userID {
		// не владелец и не админ — запрещено
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Формируем DTO нагрузки
	var loadsDTO []ds.LoadInOrder
	for _, mmo := range loads {
		loadsDTO = append(loadsDTO, ds.LoadInOrder{
			ID:          mmo.LoadID,
			Title:       mmo.Load.Title,
			Consumption: &mmo.Load.Consumption,
			Image:       mmo.Load.Image,
		})
	}

	// DTO заказа
	var iops, dbcache, arpt, ramos, cpuload, rps *float64
	if order.IOPS.Valid {
		iops = &order.IOPS.Float64
	}
	if order.DBcache.Valid {
		dbcache = &order.DBcache.Float64
	}
	if order.ARPT.Valid {
		arpt = &order.ARPT.Float64
	}
	if order.RAM_OS.Valid {
		ramos = &order.RAM_OS.Float64
	}
	if order.CPUload.Valid {
		cpuload = &order.CPUload.Float64
	}
	if order.RPS.Valid {
		rps = &order.RPS.Float64
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
		ARPT:       arpt,
		RAM_OS:     ramos,
		CPUload:    cpuload,
		RPS:        rps,
		DateCreate: order.DateCreate,
		DateForm:   &order.DateForm,
		DateFinish: dateFinish,
		Loads:      loadsDTO,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"order": resp,
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
		"order": order,
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
		"order": order,
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
		"order": order,
		"loads": loads,
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
		"orderID": orderID,
	})
}
