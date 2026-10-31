package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"smswhiteinfo/customer-api/internal/sms"
)

type PaymentRequest struct {
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	Amount        float64
	Currency      string
	Metadata      map[string]any
	ReturnURL     string
	WebhookURL    string
}

type PaymentResponse struct {
	Successful       bool
	GatewayDriver    string
	GatewayStatus    string
	GatewayPaymentID string
	CheckoutURL      string
	Amount           float64
	Currency         string
	Raw              map[string]any
}

type PipraPayDriver struct {
	apiKey  string
	baseURL string
	client  *sms.SafeHTTPClient
}

func NewPipraPayDriver(apiKey, baseURL string) (*PipraPayDriver, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("PipraPay API key is required")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("PipraPay base URL is required")
	}

	return &PipraPayDriver{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  sms.NewSafeHTTPClient(baseURL, 15*time.Second),
	}, nil
}

func (d *PipraPayDriver) Create(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	metaJSON, _ := json.Marshal(req.Metadata)
	payload := map[string]any{
		"full_name":     req.CustomerName,
		"email_address": req.CustomerEmail,
		"mobile_number": req.CustomerPhone,
		"amount":        fmt.Sprintf("%.2f", req.Amount),
		"currency":      req.Currency,
		"metadata":      string(metaJSON),
		"return_url":    req.ReturnURL,
		"webhook_url":   req.WebhookURL,
	}

	headers := map[string]string{
		"MHS-PIPRAPAY-API-KEY": d.apiKey,
	}

	respBytes, statusCode, err := d.client.PostJSON(ctx, "checkout/redirect", payload, headers)
	if err != nil {
		return nil, err
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("PipraPay checkout HTTP %d: %s", statusCode, string(respBytes))
	}

	var data map[string]any
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return nil, err
	}

	ppID, _ := data["pp_id"].(string)
	ppURL, _ := data["pp_url"].(string)

	if ppID == "" {
		return nil, fmt.Errorf("PipraPay checkout did not return a pp_id: %s", string(respBytes))
	}

	return &PaymentResponse{
		Successful:       true,
		GatewayDriver:    "piprapay",
		GatewayStatus:    "pending",
		GatewayPaymentID: ppID,
		CheckoutURL:      ppURL,
		Amount:           req.Amount,
		Currency:         req.Currency,
		Raw:              data,
	}, nil
}

func (d *PipraPayDriver) Verify(ctx context.Context, ppID string) (*PaymentResponse, error) {
	payload := map[string]any{
		"pp_id": ppID,
	}

	headers := map[string]string{
		"MHS-PIPRAPAY-API-KEY": d.apiKey,
	}

	respBytes, statusCode, err := d.client.PostJSON(ctx, "verify-payment", payload, headers)
	if err != nil {
		return nil, err
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("PipraPay verify HTTP %d: %s", statusCode, string(respBytes))
	}

	var data map[string]any
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return nil, err
	}

	status, _ := data["status"].(string)
	currency, _ := data["currency"].(string)

	var amount float64
	switch v := data["amount"].(type) {
	case float64:
		amount = v
	case string:
		var parsed float64
		_, _ = fmt.Sscanf(v, "%f", &parsed)
		amount = parsed
	}

	return &PaymentResponse{
		Successful:       strings.ToLower(status) == "completed",
		GatewayDriver:    "piprapay",
		GatewayStatus:    status,
		GatewayPaymentID: ppID,
		Amount:           amount,
		Currency:         currency,
		Raw:              data,
	}, nil
}

func (d *PipraPayDriver) HandleWebhook(payload map[string]any) *PaymentResponse {
	status, _ := payload["status"].(string)
	ppID, _ := payload["pp_id"].(string)
	currency, _ := payload["currency"].(string)

	var amount float64
	switch v := payload["amount"].(type) {
	case float64:
		amount = v
	case string:
		var parsed float64
		_, _ = fmt.Sscanf(v, "%f", &parsed)
		amount = parsed
	}

	return &PaymentResponse{
		Successful:       strings.ToLower(status) == "completed",
		GatewayDriver:    "piprapay",
		GatewayStatus:    status,
		GatewayPaymentID: ppID,
		Amount:           amount,
		Currency:         currency,
		Raw:              payload,
	}
}
