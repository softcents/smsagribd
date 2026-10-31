package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"smswhiteinfo/customer-api/internal/sms"
)

type QSmsDriver struct {
	credentials map[string]any
	client      *sms.SafeHTTPClient
}

func NewQSmsDriver(credentials map[string]any) *QSmsDriver {
	baseURL := getString(credentials, "base_url", "")
	return &QSmsDriver{
		credentials: credentials,
		client:      sms.NewSafeHTTPClient(baseURL, 15*time.Second),
	}
}

func (d *QSmsDriver) Send(ctx context.Context, req sms.SmsRequest) (sms.SmsResponse, error) {
	isUnicode := "false"
	if req.IsUnicode {
		isUnicode = "true"
	}

	payload := map[string]any{
		"ApiKey":        getString(d.credentials, "api_key", ""),
		"ClientId":      getString(d.credentials, "client_id", ""),
		"SenderId":      req.From,
		"MobileNumbers": strings.TrimPrefix(strings.TrimSpace(req.To), "+"),
		"Message":       req.Body,
		"Is_Unicode":    isUnicode,
	}

	respBytes, statusCode, err := d.client.PostJSON(ctx, "api/v2/SendSMS", payload, nil)
	if err != nil {
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	errorCode := fmt.Sprintf("%v", data["ErrorCode"])
	if statusCode >= 400 || errorCode != "0" {
		errMsg := getString(data, "ErrorDescription", "Unknown error")
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}

	var msgID *string
	if list, ok := data["Data"].([]any); ok && len(list) > 0 {
		if first, ok := list[0].(map[string]any); ok {
			idStr := fmt.Sprintf("%v", first["MessageId"])
			if idStr != "" && idStr != "<nil>" {
				msgID = &idStr
			}
		}
	}

	return sms.SmsResponse{
		Successful:       true,
		GatewayMessageId: msgID,
		Raw:              data,
	}, nil
}

func (d *QSmsDriver) SendBulk(ctx context.Context, req sms.BulkSmsRequest) (sms.BulkSmsResponse, error) {
	cleanRecipients := make([]string, len(req.Recipients))
	for i, r := range req.Recipients {
		cleanRecipients[i] = strings.TrimPrefix(strings.TrimSpace(r), "+")
	}

	isUnicode := "false"
	if req.IsUnicode {
		isUnicode = "true"
	}

	payload := map[string]any{
		"ApiKey":        getString(d.credentials, "api_key", ""),
		"ClientId":      getString(d.credentials, "client_id", ""),
		"SenderId":      req.From,
		"MobileNumbers": strings.Join(cleanRecipients, ","),
		"Message":       req.Body,
		"Is_Unicode":    isUnicode,
	}

	respBytes, statusCode, err := d.client.PostJSON(ctx, "api/v2/SendSMS", payload, nil)
	if err != nil {
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	errorCode := fmt.Sprintf("%v", data["ErrorCode"])
	if statusCode >= 400 || errorCode != "0" {
		errMsg := getString(data, "ErrorDescription", "Unknown error")
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}

	byPhone := make(map[string]string)
	if list, ok := data["Data"].([]any); ok {
		for _, item := range list {
			if m, ok := item.(map[string]any); ok {
				phone := strings.TrimPrefix(fmt.Sprintf("%v", m["MobileNumber"]), "+")
				id := fmt.Sprintf("%v", m["MessageId"])
				byPhone[phone] = id
			}
		}
	}

	perRecipient := make([]string, len(req.Recipients))
	for i, r := range req.Recipients {
		clean := strings.TrimPrefix(strings.TrimSpace(r), "+")
		perRecipient[i] = byPhone[clean]
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

func (d *QSmsDriver) FetchDeliveryStatus(ctx context.Context, gatewayMessageId string, recipient string) (sms.DeliveryStatusResponse, error) {
	params := map[string]string{
		"ApiKey":    getString(d.credentials, "api_key", ""),
		"ClientId":  getString(d.credentials, "client_id", ""),
		"MessageId": gatewayMessageId,
	}

	respBytes, statusCode, err := d.client.Get(ctx, "api/v2/MessageStatus", params, nil)
	if err != nil {
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	errorCode := fmt.Sprintf("%v", data["ErrorCode"])
	if statusCode >= 400 || errorCode != "0" {
		errMsg := getString(data, "ErrorDescription", "Unknown error")
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}

	subData, _ := data["Data"].(map[string]any)
	status := strings.ToUpper(getString(subData, "Status", ""))
	rawDoneDate := getString(subData, "DoneDate", "")

	isDelivered := status == "DELIVRD" || status == "REQUEST SUCCESS"
	if isDelivered {
		doneAt := parseQSmsDoneDate(rawDoneDate)
		return sms.DeliveryStatusResponse{
			Outcome:     sms.OutcomeDelivered,
			RawStatus:   status,
			DeliveredAt: doneAt,
			Raw:         data,
		}, nil
	}

	if status == "SENT" || status == "ACCEPTD" || status == "" {
		return sms.DeliveryStatusResponse{
			Outcome:   sms.OutcomeInTransit,
			RawStatus: status,
			Raw:       data,
		}, nil
	}

	return sms.DeliveryStatusResponse{
		Outcome:      sms.OutcomeFailed,
		RawStatus:    status,
		ErrorMessage: status,
		Raw:          data,
	}, nil
}

func parseQSmsDoneDate(doneDate string) *time.Time {
	formats := []string{
		"02 Jan 2006 15:04:05",
		"02-Jan-2006 03:04:05 PM",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, doneDate); err == nil {
			return &t
		}
	}
	return nil
}
