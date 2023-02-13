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

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/receipts/process", func(c *gin.Context) {
		var receipt models.Receipt

		// Validate input
		if err := c.ShouldBindJSON(&receipt); err != nil {
		  c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		  return
		}

		// Alphanumeric Regex
		alphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9]+`)

		// One point for every alphanumeric character in the retailer name
		points := len(alphanumericRegex.ReplaceAllString(receipt.Retailer, ""))

		switch receipt.Total[len(receipt.Total)-3:] {
		case ".00":
			// 50 points if the total is a round dollar amount with no cents
			// 25 points if the total is a multiple of 0.25
			points = points + 50 + 25
		case ".25":
			// 25 points if the total is a multiple of 0.25
			points = points + 25
		case ".50":
			// 25 points if the total is a multiple of 0.25
			points = points + 25
		case ".75":
			// 25 points if the total is a multiple of 0.25
			points = points + 25
		}

		// 5 points for every two items on the receip
		points = points + (len(receipt.Items) / 2 * 5)


		for _, item := range receipt.Items {
			//If the trimmed length of the item description is a multiple of 3
			if len(strings.Trim(item.ShortDescription, " ")) % 3 == 0 {
				price, err := strconv.ParseFloat(item.Price, 64)

				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		  			return
				}

				// Multiply the price by 0.2 and round up to the nearest integer
				// The result is the number of points earned
				points = points + int(math.Ceil(price * 0.2))
			}
		}

		date, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", receipt.PurchaseDate, receipt.PurchaseTime))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if date.Day() % 2 == 1 {
			// 6 points if the day in the purchase date is odd
			points = points + 6
		}

		after := time.Date(date.Year(), date.Month(), date.Day(), 14, 0, 0, 0, time.UTC)
		before := time.Date(date.Year(), date.Month(), date.Day(), 16, 0, 0, 0, time.UTC)

		if date.After(after) && date.Before(before) {
			// 10 points if the time of purchase is after 2:00pm and before 4:00pm
			points = points + 10
		}

		// Set points for generated uuid of receipt
		id := uuid.New()
		db[id] = points
		c.JSON(http.StatusOK, gin.H{"id": id.String()})
	})

	// Get user value
	r.GET("/receipts/:id/points", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))

		if err == nil {
			// Returns points of uuid
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
