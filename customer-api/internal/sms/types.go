package sms

import (
	"context"
	"time"
)

// DeliveryOutcome represents the normalized status of an SMS message.
type DeliveryOutcome string

const (
	OutcomeDelivered DeliveryOutcome = "delivered"
	OutcomeFailed    DeliveryOutcome = "failed"
	OutcomeInTransit DeliveryOutcome = "in_transit"
)

// SmsRequest encapsulates parameters for a single SMS dispatch.
type SmsRequest struct {
	To        string
	Body      string
	From      string // Sender ID / CLI
	IsUnicode bool
	Metadata  map[string]any
}

// SmsResponse encapsulates the result of a single SMS dispatch.
type SmsResponse struct {
	Successful       bool
	GatewayMessageId *string
	ErrorMessage     string
	Raw              any
}

// BulkSmsRequest encapsulates parameters for sending the same body to multiple recipients.
type BulkSmsRequest struct {
	Recipients []string
	Body       string
	From       string // Sender ID / CLI
	IsUnicode  bool
	Metadata   map[string]any
}

// BulkSmsResponse encapsulates the result of a bulk SMS dispatch.
type BulkSmsResponse struct {
	Successful       bool
	GatewayBatchId   *string
	PerRecipientIds  []string
	FailedRecipients map[string]string // phone -> failure reason
	ErrorMessage     string
	Raw              any
}

// DeliveryStatusResponse encapsulates the result of a delivery status check (DLR).
type DeliveryStatusResponse struct {
	Outcome      DeliveryOutcome
	RawStatus    string
	DeliveredAt  *time.Time
	ErrorMessage string
	Raw          any
}

// StatCard represents a dashboard stat metric.
type StatCard struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// DashboardData carries stat cards for admin dashboard metrics.
type DashboardData struct {
	Cards []StatCard `json:"cards"`
}

// SenderItem represents an allowed sender label from upstream.
type SenderItem struct {
	SenderID string `json:"sender_id"`
	Type     string `json:"type"`
}

// SmsGateway is the core contract for SMS delivery.
type SmsGateway interface {
	Send(ctx context.Context, req SmsRequest) (SmsResponse, error)
}

// SupportsBulkSms defines bulk transmission capabilities.
type SupportsBulkSms interface {
	SendBulk(ctx context.Context, req BulkSmsRequest) (BulkSmsResponse, error)
}

// SupportsDeliveryStatus defines delivery receipt polling capabilities.
type SupportsDeliveryStatus interface {
	FetchDeliveryStatus(ctx context.Context, gatewayMessageId string, recipient string) (DeliveryStatusResponse, error)
}

// SupportsDashboard defines balance/credit retrieval capabilities for admin views.
type SupportsDashboard interface {
	GetDashboardData(ctx context.Context) (DashboardData, error)
}

// SupportsSenders defines sender ID retrieval.
type SupportsSenders interface {
	FetchSenders(ctx context.Context) ([]SenderItem, error)
}
