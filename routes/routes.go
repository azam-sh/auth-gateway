package routes

import (
	"authgateway/controllers"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StartRoutes() {
	r := gin.Default()

	r.GET("/ping", ping)
	r.GET("/", controllers.Home)
	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)

	r.GET("/google/login", controllers.GoogleLogin)
	r.GET("/callback", controllers.Callback)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "page not found"})
	})

	err := r.Run()
	if err != nil {
		log.Panic("failed to start router")
	}
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, "Connection established!")
}
