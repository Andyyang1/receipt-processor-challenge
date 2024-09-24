package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type Items struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Receipt struct {
	ID            string  `json:"id"`
	Retailer      string  `json:"retailer"`
	PurchasedDate string  `json:"purchaseDate"`
	PurchasedTime string  `json:"purchaseTime"`
	Total         string  `json:"total"`
	Items         []Items `json:"items"`
}

type Response struct {
	ID     string `json:"id"`
	Points int    `json:"points"`
}

var receiptStore = make(map[string]Receipt)

func addRetailerPoints(retailerName string) int {
	points := 0
	for _, r := range retailerName {
		if unicode.IsDigit(r) || unicode.IsLetter(r) {
			points++
		}
	}
	return points
}

func isMultiple(total float64) bool {
	return int(total*100)%25 == 0
}
func isRoundDollar(total float64) bool {
	return total == float64(int(total))
}

func calculateItemDescription(items []Items) int {
	currPoint := 0
	for _, item := range items {
		trimmedDescription := strings.TrimSpace(item.ShortDescription)
		descriptionLen := len(trimmedDescription)
		priceFloat, _ := strconv.ParseFloat(item.Price, 64)

		if descriptionLen%3 == 0 {
			currPoint += int(math.Ceil(priceFloat * 0.02))

		}
	}
	return currPoint
}

func isOddDate(dateString string) bool {
	date, _ := time.Parse("2006-01-02", dateString)
	day := date.Day()
	return day%2 != 0

}

func isBetween2And4PM(purchasedTime string) bool {
	parsedTime, err := time.Parse("15:04", purchasedTime) // "15:04" is the Go time layout for "HH:MM"
	if err != nil {
		return false
	}
	hour := parsedTime.Hour()
	return hour >= 14 && hour < 16
}

func calculatePoints(receipt Receipt) int {
	var point int = 0

	retailerName := receipt.Retailer

	point += addRetailerPoints(retailerName)

	totalFloat, _ := strconv.ParseFloat(receipt.Total, 64)

	if isRoundDollar(totalFloat) {
		point += 50
	}

	if isMultiple(totalFloat) {
		point += 25
	}

	//every 2 items add 5 points
	numberOfItems := len(receipt.Items)
	point += (numberOfItems / 2) * 5

	//check item description
	point += calculateItemDescription(receipt.Items)

	//check if date is odd
	if isOddDate(receipt.PurchasedDate) {
		point += 6
	}

	//check time is between 2pm and 4pm
	if isBetween2And4PM(receipt.PurchasedTime) {
		point += 10
	}

	return point
}

func processReceiptPointsHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/receipts/")
	id = strings.TrimSuffix(id, "/points")

	receipt, exists := receiptStore[id]
	if !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}
	points := calculatePoints(receipt)

	response := Response{Points: points}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	fmt.Printf("Total Points %+v \n", response.Points)
}

func processReceiptHandler(w http.ResponseWriter, r *http.Request) {
	//check the request is a POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	var receipt Receipt

	err = json.Unmarshal(body, &receipt)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	receipt.ID = id
	receiptStore[id] = receipt

	// Create a response object
	response := Response{ID: id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to my web service \n")
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/receipts/process", processReceiptHandler)
	http.HandleFunc("/receipts/{id}/points", processReceiptPointsHandler)

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
