package handler

import (
	"net/http"
	"strconv"

	"RIP/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetLoads(ctx *gin.Context) {
	searchQuery := ctx.Query("load_search")
	var loads []ds.Load
	var err error

	if searchQuery == "" {
		loads, err = h.Repository.GetLoads()
	} else {
		loads, err = h.Repository.GetLoadsByTitle(searchQuery)
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Для примера используем userID = 1
	userID := 2

	// Ищем черновой заказ пользователя
	order, err := h.Repository.GetDraftOrder(userID)
	var orderCount int64 = 0
	if err != nil {
		logrus.Warn("Не удалось получить черновой заказ: ", err)
	} else if order != nil {
		// Получаем количество материалов в черновике
		orderCount, err = h.Repository.GetOrderLoadsCount(order.ID)
		if err != nil {
			logrus.Warn("Не удалось получить количество материалов в черновике: ", err)
			orderCount = 0
		}
	}

	ctx.HTML(http.StatusOK, "loads_list.html", gin.H{
		"loads":      loads,
		"query":      searchQuery,
		"orderCount": orderCount,
		// передаём ID чернового заказа в шаблон (0 если заказа нет)
		"orderID": func() int {
			if order != nil {
				return order.ID
			}
			return 0
		}(),
	})
}

// POST /orders/draft/add/:id
func (h *Handler) AddLoadToDraftOrder(ctx *gin.Context) {
	loadIDStr := ctx.Param("id")
	loadID, err := strconv.Atoi(loadIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Для примера используем userID = 1
	userID := 2

	// Получаем черновой заказ
	order, err := h.Repository.GetDraftOrder(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if order == nil {
		order, err = h.Repository.CreateDraftOrder(userID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	// Добавляем материал
	if err := h.Repository.AddLoadToOrder(order.ID, loadID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Сохраняем count в сессии или просто редиректим на текущую страницу
	ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
}

// Получение заказа по ID
func (h *Handler) GetLoadsOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Если id == 0 — отвечаем унифицированным сообщением, не перенаправляя/не показывая внутреннюю ошибку
	if id == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"description": "заказ не найден или удален",
		})
		return
	}

	order, mmos, err := h.Repository.GetOrderByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Если статус заказа не черновик, считаем, что заказа нет/он удалён
	if order.RequestStatus != "черновик" {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"description": "заказ не найден или удален",
		})
		return
	}

	ctx.HTML(http.StatusOK, "loads_order.html", gin.H{
		"order": order,
		"loads": mmos, // вот сюда прокидываем список материалов
	})
}
