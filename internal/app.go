package internal

import (
	"micli/internal/middleware"
	"micli/internal/static"

	"github.com/gin-gonic/gin"
)

type App struct {
	*gin.Engine
	Port string
}

func NewApp(port string) *App {
	return &App{
		Engine: gin.Default(),
		Port:   port,
	}
}

func (a *App) RegisterMiddlewares() {
	a.Use(middleware.CORSMiddleware())
}

func (a *App) RegisterRoutes() {
	web := a.Group("/")
	api := web.Group("api")
	api.Any("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})
	static.Static(web, func(handlers ...gin.HandlerFunc) {
		a.NoRoute(handlers...)
	})
}
