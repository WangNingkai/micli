package internal

import (
	"github.com/gin-gonic/gin"
	"micli/internal/middleware"
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
	api := a.Group("api")
	api.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

}
