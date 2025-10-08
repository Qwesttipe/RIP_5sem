package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DELETE /api/orders/:order_id/Load/:Load_id
func (h *Handler) DeleteLoadFromOrderAPI(ctx *gin.Context) {
	orderIDStr := ctx.Param("order_id")
	loadIDStr := ctx.Param("load_id")

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID заказа"))
		return
	}

	loadID, err := strconv.Atoi(loadIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID материала"))
		return
	}

	if err := h.Repository.DeleteLoadFromOrder(loadID, orderID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"load_id":  loadID,
		"order_id": orderID,
		"message":  "услуга удалена из заявки",
	})
}

// PUT /api/orders/:order_id/load/:load_id/RPS
func (h *Handler) UpdateRPSAPI(ctx *gin.Context) {
	orderIDStr := ctx.Param("order_id")
	loadIDStr := ctx.Param("load_id")

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID заказа"))
		return
	}

	loadID, err := strconv.Atoi(loadIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID материала"))
		return
	}

	var req struct {
		RPS float64 `json:"RPS"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректное тело запроса"))
		return
	}

	if err := h.Repository.UpdateRPS(loadID, orderID, req.RPS); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"load_id":  loadID,
		"order_id": orderID,
		"RPS":      req.RPS,
	})
}
