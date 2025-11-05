package handler

import (
	"net/http"
	"strconv"

	"RIP/internal/app/ds"

	"github.com/gin-gonic/gin"
)

// Получение конкретного материала по ID
func (h *Handler) GetLoad(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var load ds.Load
	load, err = h.Repository.GetLoad(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.HTML(http.StatusOK, "detailed_load.html", gin.H{
		"load": load,
	})
}
