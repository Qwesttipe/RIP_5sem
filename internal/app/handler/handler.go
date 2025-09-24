package handler

import (
	"RIP/internal/app/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetLoads(ctx *gin.Context) {
	var loads []repository.Load
	var err error

	searchQuery := ctx.Query("findloads")
	if searchQuery == "" {
		loads, err = h.Repository.GetLoads()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		loads, err = h.Repository.GetLoadsByPattern(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":  time.Now().Format("15:04:05"),
		"loads": loads,
		"query": searchQuery,
	})
}

func (h *Handler) GetLoad(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	load, err := h.Repository.GetLoad(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "load.html", gin.H{
		"load": load,
	})
}

func (h *Handler) GetReq(ctx *gin.Context) { // Изменено с GetRequests на GetReq
	idStr := ctx.Param("id") // Добавлен параметр id
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		id = 1 // Значение по умолчанию
	}

	request, requestItems, err := h.Repository.GetRequests()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "req.html", gin.H{ // Изменено на req.html
		"time":         time.Now().Format("15:04:05"),
		"request":      request,
		"requestItems": requestItems,
		"reqID":        id, // Добавлен ID заявки
	})
}
