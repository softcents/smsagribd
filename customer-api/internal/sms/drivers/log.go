package drivers

import (
	"context"
	"fmt"
	"log"
	"time"

	"smswhiteinfo/customer-api/internal/sms"
)

type LogDriver struct {
	credentials map[string]any
}

func NewLogDriver(credentials map[string]any) *LogDriver {
	return &LogDriver{credentials: credentials}
}

func (d *LogDriver) Send(ctx context.Context, req sms.SmsRequest) (sms.SmsResponse, error) {
	msgID := fmt.Sprintf("mock-log-%d", time.Now().UnixNano())
	log.Printf("[LOG DRIVER] SMS To: %s | From: %s | Body: %s | ID: %s", req.To, req.From, req.Body, msgID)

	return sms.SmsResponse{
		Successful:       true,
		GatewayMessageId: &msgID,
		Raw:              map[string]any{"id": msgID, "status": "sent"},
	}, nil
}

func (d *LogDriver) SendBulk(ctx context.Context, req sms.BulkSmsRequest) (sms.BulkSmsResponse, error) {
	batchID := fmt.Sprintf("mock-log-batch-%d", time.Now().UnixNano())
	perRecipient := make([]string, len(req.Recipients))
	for i := range req.Recipients {
		perRecipient[i] = fmt.Sprintf("%s-%d", batchID, i)
	}

	log.Printf("[LOG DRIVER] Bulk SMS to %d recipients | From: %s | BatchID: %s", len(req.Recipients), req.From, batchID)

	return sms.BulkSmsResponse{
		Successful:      true,
		GatewayBatchId:  &batchID,
		PerRecipientIds: perRecipient,
		Raw:             map[string]any{"batch_id": batchID, "count": len(req.Recipients)},
	}, nil
}

func (d *LogDriver) FetchDeliveryStatus(ctx context.Context, gatewayMessageId string, recipient string) (sms.DeliveryStatusResponse, error) {
	now := time.Now()
	return sms.DeliveryStatusResponse{
		Outcome:      sms.OutcomeDelivered,
		RawStatus:    "DELIVRD",
		DeliveredAt:  &now,
		ErrorMessage: "",
		Raw:          map[string]any{"status": "DELIVRD"},
	}, nil
}
