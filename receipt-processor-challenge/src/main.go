package main

import (
	"fmt"
	"net/http"
	"strings"
	"strconv"
	"math"
	"regexp"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"receipt-processor-challenge/src/models"
)

var db = make(map[uuid.UUID]int)
var count = 0

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/receipts/process", func(c *gin.Context) {
		var receipt models.Receipt

		// Validate input
		if err := c.ShouldBindJSON(&receipt); err != nil {
		  c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		  return
		}

		nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

		points := len(nonAlphanumericRegex.ReplaceAllString(strings.ReplaceAll(receipt.Retailer, " ", ""), ""))

		if(strings.HasSuffix(receipt.Total, ".00")) {
			points = points + 50 + 25
		}

		if(strings.HasSuffix(receipt.Total, ".25") || strings.HasSuffix(receipt.Total, ".50") || strings.HasSuffix(receipt.Total, ".75")) {
			points = points + 25
		}

		points = points + (len(receipt.Items) / 2 * 5)

		for _, item := range receipt.Items {
			if len(strings.Trim(item.ShortDescription, " ")) % 3 == 0 {
				price, err := strconv.ParseFloat(item.Price, 64)
				
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		  			return
				}

				points = points + int(math.Ceil(price * 0.2))
			}
		}

		date, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", receipt.PurchaseDate, receipt.PurchaseTime))
				
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if date.Day() % 2 == 1 {
			points = points + 6
		}

		after := time.Date(date.Year(), date.Month(), date.Day(), 14, 0, 0, 0, time.UTC)
		before := time.Date(date.Year(), date.Month(), date.Day(), 16, 0, 0, 0, time.UTC)

		if date.After(after) && date.Before(before) {
			points = points + 10
		}

		id := uuid.New()
		db[id] = points
		c.JSON(http.StatusOK, gin.H{"id": id.String()})
	})

	// Get user value
	r.GET("/receipts/:id/points", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err == nil {
			points, ok := db[id]

			if ok {
				c.JSON(http.StatusOK, gin.H{"points": points})
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "no points for id " + c.Param("id")})
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in localhost:8080
	r.Run(":8080")
}
