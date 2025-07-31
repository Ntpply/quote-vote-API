package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"golang/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var QuoteCollection *mongo.Collection

func SetQuoteCollection(c *mongo.Collection) {
	QuoteCollection = c
}

type AddQuoteRequest struct {
	Text     string `json:"text" binding:"required"`
	Author   string `json:"author"`
	Category string `json:"category"`
}

func AddQuote(c *gin.Context) {
	var req AddQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text is required"})
		return
	}

	author := req.Author
	if author == "" {
		author = "anonymous"
	}

	quote := models.Quote{
		Text:      req.Text,
		Author:    author,
		CreatedAt: time.Now(),
		Votes:     0,
		VotedBy:   []string{},
		Category:  req.Category,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := QuoteCollection.InsertOne(ctx, quote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add quote"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Quote added"})
}

// GET /quotes?limit=10&page=1&search=xxx&sort=createdAt
func GetQuotes(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")
	search := c.Query("search")
	category := c.Query("category")
	sortBy := c.DefaultQuery("sortBy", "createdAt")   // default: sort by createdAt
	sortOrder := c.DefaultQuery("sortOrder", "desc")  // desc / asc

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 10
	}
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page <= 0 {
		page = 1
	}

	// === Filter ===
	filter := bson.M{}
	if search != "" {
		filter["text"] = bson.M{"$regex": search, "$options": "i"}
	}
	if category != "" {
		filter["category"] = category
	}

	// === Sort ===
	sort := bson.D{}
	order := 1
	if sortOrder == "desc" {
		order = -1
	}
	sort = append(sort, bson.E{Key: sortBy, Value: order})

	opts := options.Find().
		SetLimit(limit).
		SetSkip((page - 1) * limit).
		SetSort(sort)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := QuoteCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quotes"})
		return
	}
	defer cursor.Close(ctx)

	var quotes []models.Quote
	if err := cursor.All(ctx, &quotes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse quotes"})
		return
	}

	c.JSON(http.StatusOK, quotes)
}


type VoteRequest struct {
	Username string `json:"username" binding:"required"`
	Action   string `json:"action" binding:"required"` // "vote" or "unvote"
}

func VoteQuote(c *gin.Context) {
	quoteID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(quoteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quote ID"})
		return
	}

	var req VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and action are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var quote models.Quote
	err = QuoteCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&quote)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quote"})
		return
	}

	userAlreadyVoted := false
	for _, voter := range quote.VotedBy {
		if voter == req.Username {
			userAlreadyVoted = true
			break
		}
	}

	var update bson.M

	switch req.Action {
	case "vote":
		if userAlreadyVoted {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User has already voted for this quote"})
			return
		}

		update = bson.M{
			"$inc":  bson.M{"votes": 1},
			"$push": bson.M{"votedBy": req.Username},
		}
	case "unvote":
		if !userAlreadyVoted {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User has not voted for this quote"})
			return
		}

		update = bson.M{
			"$inc":  bson.M{"votes": -1},
			"$pull": bson.M{"votedBy": req.Username},
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	_, err = QuoteCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote updated successfully"})
}



type UpdateQuoteRequest struct {
	Text string `json:"text" binding:"required"`
}

func UpdateQuote(c *gin.Context) {
	quoteID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(quoteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quote ID"})
		return
	}

	var req UpdateQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// หา quote ที่จะ update
	var quote models.Quote
	err = QuoteCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&quote)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quote"})
		return
	}

	// เช็คจำนวนโหวต ถ้าไม่เป็น 0 ห้ามแก้ไข
	if quote.Votes > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot edit quote that already has votes"})
		return
	}

	// อัพเดตข้อความคำคม
	update := bson.M{
		"$set": bson.M{"text": req.Text},
	}

	_, err = QuoteCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quote"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quote updated successfully"})
}
