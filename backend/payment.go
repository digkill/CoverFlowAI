package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type LavaTopCreateOrderRequest struct {
	Sum      float64 `json:"sum"`
	OrderID  string  `json:"orderId"`
	ShopID   string  `json:"shopId"`
	Currency string  `json:"currency"` // "RUB" or "USD"
}

type LavaTopCreateOrderResponse struct {
	Status string `json:"status"`
	Data   struct {
		URL     string `json:"url"`
		InvoiceID string `json:"invoiceId"`
		OrderID   string `json:"orderId"`
	} `json:"data"`
	Message string `json:"message"`
}

func createLavaTopOrder(transactionID string, amount float64, currency string, pkg *Package) (string, string, error) {
	shopID := os.Getenv("LAVA_SHOP_ID")
	secretKey := os.Getenv("LAVA_SECRET_KEY")
	apiURL := os.Getenv("LAVA_API_URL")

	if apiURL == "" {
		apiURL = "https://api.lava.top"
	}

	if shopID == "" || secretKey == "" {
		return "", "", fmt.Errorf("LAVA_SHOP_ID and LAVA_SECRET_KEY must be set")
	}

	// Convert currency code for Lava Top (RUB or USD)
	lavaCurrency := currency
	if currency == "USD" {
		lavaCurrency = "USD"
	} else if currency == "RUB" {
		lavaCurrency = "RUB"
	}

	reqBody := LavaTopCreateOrderRequest{
		Sum:      amount,
		OrderID:  transactionID,
		ShopID:   shopID,
		Currency: lavaCurrency,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/invoice/create", apiURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", secretKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("lava top API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var lavaResp LavaTopCreateOrderResponse
	if err := json.Unmarshal(body, &lavaResp); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	if lavaResp.Status != "success" {
		return "", "", fmt.Errorf("lava top error: %s", lavaResp.Message)
	}

	return lavaResp.Data.InvoiceID, lavaResp.Data.URL, nil
}

