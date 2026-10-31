package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"smswhiteinfo/customer-api/internal/sms"
)

type ReveSmsDriver struct {
	credentials map[string]any
	client      *sms.SafeHTTPClient
}

func NewReveSmsDriver(credentials map[string]any) *ReveSmsDriver {
	baseURL := getString(credentials, "base_url", "")
	return &ReveSmsDriver{
		credentials: credentials,
		client:      sms.NewSafeHTTPClient(baseURL, 15*time.Second),
	}
}

func (d *ReveSmsDriver) Send(ctx context.Context, req sms.SmsRequest) (sms.SmsResponse, error) {
	payload := map[string]any{
		"apikey":         getString(d.credentials, "api_key", ""),
		"secretkey":      getString(d.credentials, "secret_key", ""),
		"callerID":       req.From,
		"toUser":         strings.TrimPrefix(strings.TrimSpace(req.To), "+"),
		"messageContent": req.Body,
	}

	respBytes, statusCode, err := d.client.PostJSON(ctx, "sendtext", payload, nil)
	if err != nil {
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	status := fmt.Sprintf("%v", data["Status"])
	if statusCode >= 400 || status != "0" {
		errMsg := getString(data, "Text", "Unknown error")
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}

	var msgID *string
	if idStr := getString(data, "Message_ID", ""); idStr != "" {
		msgID = &idStr
	}

	return sms.SmsResponse{
		Successful:       true,
		GatewayMessageId: msgID,
		Raw:              data,
	}, nil
}

func (d *ReveSmsDriver) SendBulk(ctx context.Context, req sms.BulkSmsRequest) (sms.BulkSmsResponse, error) {
	cleanRecipients := make([]string, len(req.Recipients))
	for i, r := range req.Recipients {
		cleanRecipients[i] = strings.TrimPrefix(strings.TrimSpace(r), "+")
	}

	payload := map[string]any{
		"apikey":         getString(d.credentials, "api_key", ""),
		"secretkey":      getString(d.credentials, "secret_key", ""),
		"callerID":       req.From,
		"toUser":         strings.Join(cleanRecipients, ","),
		"messageContent": req.Body,
	}

	respBytes, statusCode, err := d.client.PostJSON(ctx, "sendtext", payload, nil)
	if err != nil {
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	status := fmt.Sprintf("%v", data["Status"])
	if statusCode >= 400 || status != "0" {
		errMsg := getString(data, "Text", "Unknown error")
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}

	rawIds := getString(data, "Message_ID", "")
	var perRecipientIds []string
	if rawIds != "" {
		parts := strings.Split(rawIds, ",")
		for _, p := range parts {
			perRecipientIds = append(perRecipientIds, strings.TrimSpace(p))
		}
	}

	var batchID *string
	if len(perRecipientIds) > 0 {
		batchID = &perRecipientIds[0]
	}

	return sms.BulkSmsResponse{
		Successful:      true,
		GatewayBatchId:  batchID,
		PerRecipientIds: perRecipientIds,
		Raw:             data,
	}, nil
}

func (d *ReveSmsDriver) FetchDeliveryStatus(ctx context.Context, gatewayMessageId string, recipient string) (sms.DeliveryStatusResponse, error) {
	params := map[string]string{
		"apikey":    getString(d.credentials, "api_key", ""),
		"secretkey": getString(d.credentials, "secret_key", ""),
		"messageid": gatewayMessageId,
	}

	respBytes, statusCode, err := d.client.Get(ctx, "getstatus", params, nil)
	if err != nil {
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	status := fmt.Sprintf("%v", data["Status"])
	text := strings.ToUpper(getString(data, "Text", ""))

	if statusCode >= 400 {
		errMsg := text
		if errMsg == "" {
			errMsg = "HTTP error"
		}
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}

	// Status "0" = DELIVRD, "4" = SENT (accepted, in transit), anything else = failed.
	switch status {
	case "0":
		deliveredAt := parseEpochMs(getString(data, "Delivery Time", ""))
		return sms.DeliveryStatusResponse{
			Outcome:     sms.OutcomeDelivered,
			RawStatus:   text,
			DeliveredAt: deliveredAt,
			Raw:         data,
		}, nil
	case "4":
		return sms.DeliveryStatusResponse{
			Outcome:   sms.OutcomeInTransit,
			RawStatus: text,
			Raw:       data,
		}, nil
	default:
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeFailed,
			RawStatus:    text,
			ErrorMessage: text,
			Raw:          data,
		}, nil
	}
}

func parseEpochMs(epochMillis string) *time.Time {
	if epochMillis == "" || epochMillis == "0" {
		return nil
	}
	ms, err := strconv.ParseInt(epochMillis, 10, 64)
	if err != nil {
		return nil
	}
	t := time.UnixMilli(ms)
	return &t
}
