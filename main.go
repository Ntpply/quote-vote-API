package main

import (
	"log"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang/config"
	"golang/routes"
	"time"
	
)

func main() {
	client, db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer client.Disconnect(nil)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{
			"http://localhost:3000",
			"http://192.168.1.55:3000",   
			
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	routes.RegisterAuthRoutes(r, db)
	routes.RegisterQuoteRoutes(r, db)
	r.Run(":8080")
}