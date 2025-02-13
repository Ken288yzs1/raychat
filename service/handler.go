package service

import (
	"fmt"
	"raychat/chat"
	"raychat/middlewares"
	"raychat/service/models"
	"raychat/settings"

	"github.com/gin-gonic/gin"
)

func Run() {
	r := gin.Default()
	v1 := r.Group("/hf/v1")
	{
		v1.GET("/models", models.GetModelsEndpoint)
		v1.POST("/chat/completions", middlewares.Auth, chat.ChatEndpoint)
		v1.OPTIONS("/chat/completions", OptionsHandler)
	}
	r.Run(fmt.Sprintf(":%d", settings.Get().Port))
}

func OptionsHandler(c *gin.Context) {
	// Set headers for CORS
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "*")
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
