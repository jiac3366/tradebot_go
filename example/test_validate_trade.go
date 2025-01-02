package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"tradebot_go/tradebot/binance"
)

// ... existing code ...
func main() {
	msg := map[string]interface{}{
		"result": nil,
		"id":     1735795563523823814,
	}
	var trade binance.Trade

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("JSON marshal error: %v\n", err)
		return
	}

	if err := json.Unmarshal(jsonBytes, &trade); err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		return
	}

	fmt.Printf("trade: %+v\n", trade)

	validate := validator.New()

	err = validate.Struct(trade)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, e := range validationErrors {
			fmt.Println("Field:", e.Field(), "Error:", e.Tag())
		}
	}
}
