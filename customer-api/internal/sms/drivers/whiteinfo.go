package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"smswhiteinfo/customer-api/internal/sms"
)

type WhiteInfoDriver struct {
	credentials map[string]any
	client      *sms.SafeHTTPClient
}

func NewWhiteInfoDriver(credentials map[string]any) *WhiteInfoDriver {
	baseURL := getString(credentials, "base_url", "https://sms.whiteinfobd.com")
	return &WhiteInfoDriver{
		credentials: credentials,
		client:      sms.NewSafeHTTPClient(baseURL, 15*time.Second),
	}
}

func (d *WhiteInfoDriver) authHeaders() map[string]string {
	token := getString(d.credentials, "api_token", getString(d.credentials, "api_key", ""))
	headers := make(map[string]string)
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	return headers
}

func (d *WhiteInfoDriver) Send(ctx context.Context, req sms.SmsRequest) (sms.SmsResponse, error) {
	senderId := req.From
	if senderId == "" {
		senderId = getString(d.credentials, "sender_id", "")
	}
	if senderId == "" {
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: "No sender ID provided for White Info BD gateway.",
		}, nil
	}

	payload := map[string]any{
		"sender_id": senderId,
		"numbers":   d.formatRecipient(req.To),
		"body":      req.Body,
	}

	respBytes, statusCode, err := d.client.PostJSON(ctx, "api/v1/messages", payload, d.authHeaders())
	if err != nil {
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	if statusCode < 300 {
		var batchID *string
		if b := getString(data, "batch_id", ""); b != "" {
			batchID = &b
		}
		return sms.SmsResponse{
			Successful:       true,
			GatewayMessageId: batchID,
			Raw:              data,
		}, nil
	}

	errMsg := getString(data, "message", fmt.Sprintf("White Info BD API request failed (HTTP %d).", statusCode))
	return sms.SmsResponse{
		Successful:   false,
		ErrorMessage: errMsg,
		Raw:          data,
	}, nil
}

func (d *WhiteInfoDriver) SendBulk(ctx context.Context, req sms.BulkSmsRequest) (sms.BulkSmsResponse, error) {
	senderId := req.From
	if senderId == "" {
		senderId = getString(d.credentials, "sender_id", "")
	}
	if senderId == "" {
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: "No sender ID provided for White Info BD gateway.",
		}, nil
	}

	formatted := make([]string, len(req.Recipients))
	for i, r := range req.Recipients {
		formatted[i] = d.formatRecipient(r)
	}

	payload := map[string]any{
		"sender_id": senderId,
		"numbers":   strings.Join(formatted, ","),
		"body":      req.Body,
	}

	respBytes, statusCode, err := d.client.PostJSON(ctx, "api/v1/messages", payload, d.authHeaders())
	if err != nil {
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	if statusCode < 300 {
		var batchID *string
		if b := getString(data, "batch_id", ""); b != "" {
			batchID = &b
		}

		perRecipient := make([]string, len(req.Recipients))
		for i := range req.Recipients {
			if batchID != nil {
				perRecipient[i] = *batchID
			}
		}

		return sms.BulkSmsResponse{
			Successful:      true,
			GatewayBatchId:  batchID,
			PerRecipientIds: perRecipient,
			Raw:             data,
		}, nil
	}

	errMsg := getString(data, "message", fmt.Sprintf("White Info BD bulk API request failed (HTTP %d).", statusCode))
	return sms.BulkSmsResponse{
		Successful:   false,
		ErrorMessage: errMsg,
		Raw:          data,
	}, nil
}

func (d *WhiteInfoDriver) FetchDeliveryStatus(ctx context.Context, gatewayMessageId string, recipient string) (sms.DeliveryStatusResponse, error) {
	endpoint := fmt.Sprintf("api/v1/batches/%s", gatewayMessageId)
	respBytes, statusCode, err := d.client.Get(ctx, endpoint, nil, d.authHeaders())
	if err != nil {
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	if statusCode == 404 {
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeFailed,
			RawStatus:    "NOT_FOUND",
			ErrorMessage: getString(data, "message", "Batch not found on White Info BD."),
			Raw:          data,
		}, nil
	}

	if statusCode >= 400 {
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			RawStatus:    "HTTP_ERROR",
			ErrorMessage: getString(data, "message", "Failed to check batch status."),
			Raw:          data,
		}, nil
	}

	status := strings.ToLower(strings.TrimSpace(getString(data, "status", "")))
	switch status {
	case "completed", "delivered", "success":
		now := time.Now()
		return sms.DeliveryStatusResponse{
			Outcome:     sms.OutcomeDelivered,
			RawStatus:   status,
			DeliveredAt: &now,
			Raw:         data,
		}, nil
	case "failed", "rejected", "cancelled":
		errMsg := getString(data, "message", fmt.Sprintf("Batch %s", status))
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeFailed,
			RawStatus:    status,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	default:
		return sms.DeliveryStatusResponse{
			Outcome:   sms.OutcomeInTransit,
			RawStatus: status,
			Raw:       data,
		}, nil
	}
}

func (d *WhiteInfoDriver) GetDashboardData(ctx context.Context) (sms.DashboardData, error) {
	respBytes, statusCode, err := d.client.Get(ctx, "api/v1/balance", nil, d.authHeaders())
	if err != nil || statusCode >= 400 {
		return sms.DashboardData{
			Cards: []sms.StatCard{{
				Label:       "Wallet Balance",
				Value:       "Offline",
				Description: "Unable to connect",
				Color:       "danger",
				Icon:        "heroicon-o-exclamation-triangle",
			}},
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	balance := getString(data, "balance", "")
	currency := getString(data, "currency", "BDT")

	val := "N/A"
	if balance != "" {
		val = fmt.Sprintf("%s %s", balance, currency)
	}

	return sms.DashboardData{
		Cards: []sms.StatCard{{
			Label:       "Wallet Balance",
			Value:       val,
			Description: "Available balance on White Info BD gateway",
			Color:       "success",
			Icon:        "heroicon-o-banknotes",
		}},
	}, nil
}

func (d *WhiteInfoDriver) FetchSenders(ctx context.Context) ([]sms.SenderItem, error) {
	respBytes, statusCode, err := d.client.Get(ctx, "api/v1/senders", nil, d.authHeaders())
	if err != nil || statusCode >= 400 {
		return nil, err
	}

	var data struct {
		Data []struct {
			SenderID string `json:"sender_id"`
			Type     string `json:"type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return nil, err
	}

	items := make([]sms.SenderItem, len(data.Data))
	for i, item := range data.Data {
		items[i] = sms.SenderItem{
			SenderID: item.SenderID,
			Type:     item.Type,
		}
	}
	return items, nil
}

func (d *WhiteInfoDriver) formatRecipient(phone string) string {
	digits := nonDigitRegex.ReplaceAllString(phone, "")
	if strings.HasPrefix(digits, "8801") && len(digits) == 13 {
		return digits[2:]
	}
	return digits
}
