package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"golang/controllers"
	"golang/middleware"
)

func RegisterQuoteRoutes(r *gin.Engine, db *mongo.Database) {
	controllers.SetQuoteCollection(db.Collection("quotes"))

	quote := r.Group("/quotes")
	quote.Use(middleware.AuthMiddleware())
	{
		quote.POST("/addQuote", controllers.AddQuote)
		quote.GET("/getQuotes", controllers.GetQuotes)
		quote.POST("/vote/:id", controllers.VoteQuote)
		quote.PUT("/updateQuote/:id", controllers.UpdateQuote)

	}
}
