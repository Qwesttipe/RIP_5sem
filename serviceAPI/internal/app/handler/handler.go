package handler

import (
	"RIP/internal/app/repository"
	"html/template"

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

// RegisterHandler регистрируем маршруты
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.GetLoads)
	router.GET("/detailed_load/:id", h.GetLoad)
	router.GET("/loads_order/:id", h.GetLoadsOrder)
	router.GET("/api/load/:id", h.GetLoadAPI)
	router.GET("/api/loads", h.GetLoadsAPI)
	router.GET("/api/load_orders/draft/cart", h.GetDraftCartAPI)
	router.GET("/api/load_orders", h.GetOrdersAPI)
	router.GET("/api/load_orders/:id", h.GetOrderWithLoadsAPI)
	router.GET("/api/users/:id", h.GetUserAPI)

	router.POST("/load_orders/draft/add/:id", h.AddLoadToDraftOrder)
	router.POST("/load_orders/delete/:id", h.DeleteLoadsOrder)
	router.POST("/api/load", h.CreateLoadAPI)
	router.POST("/api/load_orders/draft/add/:id", h.AddLoadToDraftOrderAPI)
	router.POST("/api/load/:id/image", h.UploadLoadImage)
	router.POST("/api/load/:id/delete", h.DeleteLoadLogicalAPI)
	router.POST("/api/load_orders/delete/:id", h.DeleteLoadsOrderAPI)
	router.POST("/api/users/register", h.RegisterUserAPI)
	router.POST("/api/users/login", h.LoginUserAPI)
	router.POST("/api/users/logout", h.LogoutUserAPI)

	router.PUT("/api/load/:id", h.UpdateLoadAPI)
	router.PUT("/api/load_orders/:id", h.UpdateLoadOrderAPI)
	router.PUT("/api/load_orders/:id/form", h.FormLoadOrderAPI)
	router.PUT("/api/load_orders/:id/complete", h.CompleteOrRejectOrderAPI)
	router.PUT("/api/load_orders/loads/:order_id/:load_id/RPS", h.UpdateRPSAPI)
	router.PUT("/api/users/:id", h.UpdateUserAPI)

	router.DELETE("/api/load_orders/:order_id/load/:load_id", h.DeleteLoadFromOrderAPI)

}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	// Функция для склонения русских слов (1 — одна, 2-4 — несколько, 5+ — many)
	plural := func(n int, one, few, many string) string {
		nn := n % 100
		if nn >= 11 && nn <= 19 {
			return many
		}
		i := nn % 10
		if i == 1 {
			return one
		}
		if i >= 2 && i <= 4 {
			return few
		}
		return many
	}

	tmpl := template.Must(template.New("templates").Funcs(template.FuncMap{
		"plural": plural,
	}).ParseGlob("templates/*"))

	router.SetHTMLTemplate(tmpl)
	router.Static("/resources", "./resources")
}

// errorHandler для удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
