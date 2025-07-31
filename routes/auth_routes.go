package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"golang/controllers"
)

func RegisterAuthRoutes(r *gin.Engine, db *mongo.Database) {
	controllers.SetUserCollection(db.Collection("users"))

	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}
}
