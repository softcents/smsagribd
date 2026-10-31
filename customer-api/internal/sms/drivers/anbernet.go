package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"smswhiteinfo/customer-api/internal/sms"
)

type AnbernetDriver struct {
	credentials map[string]any
	client      *sms.SafeHTTPClient
}

func NewAnbernetDriver(credentials map[string]any) *AnbernetDriver {
	baseURL := getString(credentials, "base_url", "")
	return &AnbernetDriver{
		credentials: credentials,
		client:      sms.NewSafeHTTPClient(baseURL, 15*time.Second),
	}
}

func (d *AnbernetDriver) Send(ctx context.Context, req sms.SmsRequest) (sms.SmsResponse, error) {
	payload := d.buildPayload([]string{req.To}, req.Body, req.From)

	respBytes, statusCode, err := d.client.PostJSON(ctx, "api/v1/sendsms", payload, nil)
	if err != nil {
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	data := unwrapResponseData(respBytes)
	status := strings.ToLower(getString(data, "status", getString(data, "Status", "")))
	hasMsgIDs := hasMessageIds(data)
	isSuccessful := statusCode < 400 && (status == "success" || (status == "" && hasMsgIDs))

	if !isSuccessful {
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: extractAnbernetError(data),
			Raw:          data,
		}, nil
	}

	msgID := extractAnbernetFirstId(data)
	var pID *string
	if msgID != "" {
		pID = &msgID
	}

	return sms.SmsResponse{
		Successful:       true,
		GatewayMessageId: pID,
		Raw:              data,
	}, nil
}

func (d *AnbernetDriver) SendBulk(ctx context.Context, req sms.BulkSmsRequest) (sms.BulkSmsResponse, error) {
	payload := d.buildPayload(req.Recipients, req.Body, req.From)

	respBytes, statusCode, err := d.client.PostJSON(ctx, "api/v1/sendsms", payload, nil)
	if err != nil {
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	data := unwrapResponseData(respBytes)
	status := strings.ToLower(getString(data, "status", getString(data, "Status", "")))
	hasMsgIDs := hasMessageIds(data)
	isSuccessful := statusCode < 400 && (status == "success" || (status == "" && hasMsgIDs))

	if !isSuccessful {
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: extractAnbernetError(data),
			Raw:          data,
		}, nil
	}

	byReceiver := make(map[string]string)
	var rawList []any
	if list, ok := data["messageids"].([]any); ok {
		rawList = list
	} else if list, ok := data["message_ids"].([]any); ok {
		rawList = list
	}

	for _, item := range rawList {
		if m, ok := item.(map[string]any); ok {
			receiver := strings.TrimPrefix(getString(m, "receiver", getString(m, "to", getString(m, "mobile", ""))), "+")
			id := getString(m, "messageid", getString(m, "messageId", getString(m, "id", "")))
			if receiver != "" && id != "" {
				byReceiver[receiver] = id
			}
		}
	}

	perRecipient := make([]string, len(req.Recipients))
	for i, r := range req.Recipients {
		clean := strings.TrimPrefix(strings.TrimSpace(r), "+")
		perRecipient[i] = byReceiver[clean]
	}

	var batchID *string
	if len(perRecipient) > 0 && perRecipient[0] != "" {
		batchID = &perRecipient[0]
	}

	return sms.BulkSmsResponse{
		Successful:      true,
		GatewayBatchId:  batchID,
		PerRecipientIds: perRecipient,
		Raw:             data,
	}, nil
}

func (d *AnbernetDriver) FetchDeliveryStatus(ctx context.Context, gatewayMessageId string, recipient string) (sms.DeliveryStatusResponse, error) {
	payload := map[string]any{
		"account":   getString(d.credentials, "account", ""),
		"password":  getString(d.credentials, "password", ""),
		"token":     getString(d.credentials, "token", ""),
		"messageid": gatewayMessageId,
	}

	respBytes, statusCode, err := d.client.PostJSON(ctx, "api/v1/dlrcheck", payload, nil)
	if err != nil {
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	data := unwrapResponseData(respBytes)
	if statusCode >= 400 {
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			ErrorMessage: "HTTP error",
			Raw:          data,
		}, nil
	}

	status := strings.ToUpper(getString(data, "status", ""))
	switch status {
	case "DELIVERED":
		var deliveredAt *time.Time
		if dtStr := getString(data, "datetime", ""); dtStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", dtStr); err == nil {
				deliveredAt = &t
			}
		}
		return sms.DeliveryStatusResponse{
			Outcome:     sms.OutcomeDelivered,
			RawStatus:   status,
			DeliveredAt: deliveredAt,
			Raw:         data,
		}, nil
	case "ACCEPTD", "PENDING":
		return sms.DeliveryStatusResponse{
			Outcome:   sms.OutcomeInTransit,
			RawStatus: status,
			Raw:       data,
		}, nil
	default:
		errMsg := status
		if errMsg == "" {
			errMsg = "Unknown status"
		}
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeFailed,
			RawStatus:    status,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}
}

func (d *AnbernetDriver) buildPayload(recipients []string, body string, senderId string) map[string]any {
	normalized := make([]string, len(recipients))
	for i, r := range recipients {
		normalized[i] = strings.TrimPrefix(strings.TrimSpace(r), "+")
	}

	return map[string]any{
		"account":    getString(d.credentials, "account", ""),
		"password":   getString(d.credentials, "password", ""),
		"token":      getString(d.credentials, "token", ""),
		"senderid":   senderId,
		"receivers":  normalized,
		"msgdata":    body,
		"flashon":    0,
		"transtype":  "T",
		"campaignId": "",
	}
}

func unwrapResponseData(b []byte) map[string]any {
	trimmed := strings.TrimSpace(string(b))
	if strings.HasPrefix(trimmed, "[") {
		var arr []any
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil && len(arr) > 0 {
			if m, ok := arr[0].(map[string]any); ok {
				return m
			}
		}
	}
	var res map[string]any
	_ = json.Unmarshal([]byte(trimmed), &res)
	if res == nil {
		res = make(map[string]any)
	}
	return res
}

func hasMessageIds(data map[string]any) bool {
	if list, ok := data["messageids"].([]any); ok && len(list) > 0 {
		return true
	}
	if list, ok := data["message_ids"].([]any); ok && len(list) > 0 {
		return true
	}
	if v, ok := data["messageid"]; ok && v != nil && fmt.Sprintf("%v", v) != "" {
		return true
	}
	return false
}

func extractAnbernetFirstId(data map[string]any) string {
	if list, ok := data["messageids"].([]any); ok && len(list) > 0 {
		if m, ok := list[0].(map[string]any); ok {
			return getString(m, "messageid", getString(m, "messageId", getString(m, "id", "")))
		}
	}
	if v := getString(data, "messageid", getString(data, "messageId", "")); v != "" {
		return v
	}
	return ""
}

func extractAnbernetError(data map[string]any) string {
	if v := getString(data, "message", getString(data, "Message", "")); v != "" {
		return v
	}
	if v := getString(data, "error", getString(data, "Error", "")); v != "" {
		return v
	}
	if v := getString(data, "description", ""); v != "" {
		return v
	}
	if _, ok := data["Balance Status"]; ok {
		return "Balance check response instead of SMS send"
	}
	return "Unknown error"
}
