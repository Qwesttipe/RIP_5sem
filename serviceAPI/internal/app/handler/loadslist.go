package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

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
	userID := 1

	// Ищем черновой заказ пользователя
	order, err := h.Repository.GetDraftOrder(userID)
	var orderCount int64 = 0
	if err != nil {
		logrus.Warn("Не удалось получить черновой заказ: ", err)
	} else if order != nil {
		// Получаем количество нагрузочный расчётов в черновике
		orderCount, err = h.Repository.GetOrderLoadsCount(order.ID)
		if err != nil {
			logrus.Warn("Не удалось получить количество нагрузочный расчётов в черновике: ", err)
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
	userID := 1

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

	// Добавляем нагрузочный расчёт
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
		"loads": mmos, // вот сюда прокидываем список нагрузочный расчётов
	})
}

// GET /api/loads/:id
func (h *Handler) GetLoadAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	load, err := h.Repository.GetLoadByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if load == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"description": "нагрузочный расчёт не найден",
		})
		return
	}

	// Если нагрузочный расчёт найден, но он невидим — считаем, что он отсутствует
	if !load.Visability {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"description": "нагрузочный расчёт не найден",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"load":   load,
	})
}

// GET /api/loads?title=<название>
func (h *Handler) GetLoadsAPI(ctx *gin.Context) {
	title := ctx.Query("title")

	loads, err := h.Repository.GetLoadsFiltered(title)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"loads":  loads,
	})
}

// POST /api/load
func (h *Handler) CreateLoadAPI(ctx *gin.Context) {
	var input ds.Load

	// Привязываем JSON из запроса
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Создаём нагрузочный расчёт через репозиторий
	if err := h.Repository.CreateLoad(&input); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"load":   input,
	})
}

// PUT /api/load/:id
func (h *Handler) UpdateLoadAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var input ds.Load
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.UpdateLoad(id, &input)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"load":   input,
	})
}

// POST /api/load/:id/delete
func (h *Handler) DeleteLoadLogicalAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteLoadLogical(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "нагрузочный расчёт успешно скрыт",
	})
}

// POST /api/orders/draft/add/:id
func (h *Handler) AddLoadToDraftOrderAPI(ctx *gin.Context) {
	// Получаем ID нагрузочный расчёта из URL
	loadIDStr := ctx.Param("id")
	loadID, err := strconv.Atoi(loadIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Для примера используем userID = 1
	userID := 1

	// Получаем черновой заказ пользователя
	order, err := h.Repository.GetDraftOrder(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Если чернового заказа нет — создаём новый
	if order == nil {
		order, err = h.Repository.CreateDraftOrder(userID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	// Добавляем нагрузочный расчёт в заказ
	if err := h.Repository.AddLoadToOrder(order.ID, loadID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Получаем новое количество нагрузочный расчётов в заказе
	count, _ := h.Repository.GetOrderLoadsCount(order.ID)

	ctx.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "нагрузочный расчёт добавлен в черновой заказ",
		"orderID":   order.ID,
		"itemCount": count,
	})
}

// Загрузить/заменить изображение нагрузочный расчёта
func (h *Handler) UploadLoadImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid load id"})
		return
	}

	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no image file"})
		return
	}

	if err := h.Repository.UploadLoadImage(id, fileHeader); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "image uploaded"})
}
