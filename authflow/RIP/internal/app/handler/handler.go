package handler

import (
	"RIP/internal/app/config"
	redisclient "RIP/internal/app/redis"
	"RIP/internal/app/repository"
	"RIP/internal/app/role"
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
	Redis      *redisclient.Client
}

func NewHandler(r *repository.Repository, cfg *config.Config, redis *redisclient.Client) *Handler {
	return &Handler{
		Repository: r,
		Config:     cfg,
		Redis:      redis,
	}
}

// RegisterHandler регистрируем маршруты
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// ---------------------------
	// Публичные маршруты
	// ---------------------------
	router.GET("/api/loads", h.GetLoadsAPI)
	router.GET("/api/loads/:id", h.GetLoadAPI)
	router.POST("/sign_up", h.Register)
	router.POST("/api/users/login", h.LoginUserAPI)
	router.POST("/api/users/logout", h.LogoutUserAPI)

	// ---------------------------
	// Защищённые маршруты для всех авторизованных (User + Admin)
	// ---------------------------
	auth := router.Group("/api")
	auth.Use(h.WithAuthCheck(role.User, role.Admin))
	{
		// Пользователи
		auth.GET("/users/:id", h.GetUserAPI)
		auth.PUT("/users/:id", h.UpdateUserAPI)

		// Нагрузки и заказы (все методы кроме админских)
		auth.POST("/load_orders/draft/add/:id", h.AddLoadToDraftOrderAPI)
		auth.GET("/load_orders", h.GetOrdersAPI)
		auth.GET("/load_orders/:id", h.GetOrderWithLoadsAPI)
		auth.GET("/load_orders/draft/cart", h.GetDraftCartAPI)
		auth.PUT("/load_orders/:id", h.UpdateLoadOrderAPI)
		auth.PUT("/load_orders/:id/form", h.FormLoadOrderAPI)
		auth.PUT("/load_orders/loads/:order_id/:load_id/rps", h.UpdateRPSAPI)
		auth.POST("/load_orders/delete/:id", h.DeleteLoadsOrderAPI)
		auth.DELETE("/load_orders/:order_id/load/:load_id", h.DeleteLoadFromOrderAPI)
	}

	// ---------------------------
	// Только Admin
	// ---------------------------
	admin := router.Group("/api")
	admin.Use(h.WithAuthCheck(role.Admin))
	{
		admin.POST("/load", h.CreateLoadAPI)
		admin.PUT("/load/:id", h.UpdateLoadAPI)
		admin.POST("/load/:id/image", h.UploadLoadImage)
		admin.POST("/load/:id/delete", h.DeleteLoadLogicalAPI)
		admin.PUT("/load_orders/:id/complete", h.CompleteOrRejectOrderAPI)
	}
}

// RegisterStatic регистрирует статические файлы и шаблоны
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
	router.Static("/static", "./resources")
}

// errorHandler для удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"description": err.Error(),
	})
}

// getUserFromContext извлекает ID и роль пользователя из gin.Context
// Возвращает userID и role (int), при отсутствии — нули
func (h *Handler) getUserFromContext(ctx *gin.Context) (int, int) {
	uid, _ := ctx.Get("userID")
	rid, _ := ctx.Get("role")

	userID := 0
	if v, ok := uid.(int); ok {
		userID = v
	}

	roleID := 0
	if r, ok := rid.(int); ok {
		roleID = r
	}

	return userID, roleID
}
