package drivers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"smswhiteinfo/customer-api/internal/sms"
)

func init() {
	sms.SetTestMode(true)
}

func TestSecurityGuardBlocksExternalOutbound(t *testing.T) {
	driver := NewInfozillionDriver(map[string]any{
		"base_url": "https://api.mnpspbd.com",
	})

	_, err := driver.Send(context.Background(), sms.SmsRequest{
		To:   "8801711111111",
		Body: "Test",
	})
	if err != nil {
		t.Fatalf("Unexpected Send error: %v", err)
	}

	// The driver internally returns an error message when the security guard blocks
	resp, _ := driver.Send(context.Background(), sms.SmsRequest{
		To:   "8801711111111",
		Body: "Test",
	})
	if resp.Successful {
		t.Fatalf("Expected security guard to block request, but got successful=true")
	}
	if !strings.Contains(resp.ErrorMessage, "SECURITY GUARD BLOCKED OUTBOUND REQUEST") {
		t.Errorf("Expected security guard error message, got: %s", resp.ErrorMessage)
	}
}

func TestInfozillionNumericCliHotfix(t *testing.T) {
	driver := NewInfozillionDriver(map[string]any{
		"bill_msisdn": "8809612000000",
		"cli":         "8809612000000",
	})

	// Case 1: Numeric CLI starting with 0
	p1 := driver.BuildSendPayload([]string{"8801700000001"}, "Hello", "01712345678", false)
	if p1["cli"] != "8801712345678" {
		t.Errorf("Expected cli '8801712345678', got '%v'", p1["cli"])
	}
	if p1["billMsisdn"] != "8801712345678" {
		t.Errorf("Expected billMsisdn '8801712345678', got '%v'", p1["billMsisdn"])
	}

	// Case 2: Numeric CLI starting with 880
	p2 := driver.BuildSendPayload([]string{"8801700000001"}, "Hello", "8801812345678", false)
	if p2["cli"] != "8801812345678" {
		t.Errorf("Expected cli '8801812345678', got '%v'", p2["cli"])
	}
	if p2["billMsisdn"] != "8801812345678" {
		t.Errorf("Expected billMsisdn '8801812345678', got '%v'", p2["billMsisdn"])
	}

	// Case 3: Masking / alphanumeric CLI
	p3 := driver.BuildSendPayload([]string{"8801700000001"}, "Hello", "WHITEINFO", false)
	if p3["cli"] != "WHITEINFO" {
		t.Errorf("Expected cli 'WHITEINFO', got '%v'", p3["cli"])
	}
	if p3["billMsisdn"] != "8809612000000" {
		t.Errorf("Expected billMsisdn '8809612000000', got '%v'", p3["billMsisdn"])
	}
}

func TestInfozillionMockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "send-sms") {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)

			_ = json.NewEncoder(w).Encode(map[string]any{
				"serverResponseCode": "9000",
				"serverTxnId":        "TXN-INFO-12345",
				"dndMsisdn":          []any{},
			})
			return
		}
		if strings.Contains(r.URL.Path, "check-delivery-report") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"serverResponseCode": "9000",
				"deliveryStatus":     []any{"DELIVERED"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	driver := NewInfozillionDriver(map[string]any{
		"base_url":    ts.URL,
		"bill_msisdn": "8809612000000",
	})

	// Test Send
	resp, err := driver.Send(context.Background(), sms.SmsRequest{
		To:   "8801700000001",
		Body: "Test SMS",
		From: "WHITEINFO",
	})
	if err != nil {
		t.Fatalf("Send error: %v", err)
	}
	if !resp.Successful || *resp.GatewayMessageId != "TXN-INFO-12345" {
		t.Errorf("Send response mismatch: %+v", resp)
	}

	// Test DLR
	dlr, err := driver.FetchDeliveryStatus(context.Background(), "TXN-INFO-12345", "8801700000001")
	if err != nil {
		t.Fatalf("DLR error: %v", err)
	}
	if dlr.Outcome != sms.OutcomeDelivered {
		t.Errorf("Expected OutcomeDelivered, got: %s", dlr.Outcome)
	}
}

func TestReveSmsMockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "sendtext") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Status":     "0",
				"Text":       "ACCEPTD",
				"Message_ID": "REVE-101,REVE-102",
			})
			return
		}
		if strings.Contains(r.URL.Path, "getstatus") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Status":        "0",
				"Text":          "DELIVRD",
				"Delivery Time": "1771111222000",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	driver := NewReveSmsDriver(map[string]any{
		"base_url":   ts.URL,
		"api_key":    "key",
		"secret_key": "sec",
	})

	bulkResp, err := driver.SendBulk(context.Background(), sms.BulkSmsRequest{
		Recipients: []string{"8801700000001", "8801700000002"},
		Body:       "Bulk Reve",
		From:       "WHITEINFO",
	})
	if err != nil {
		t.Fatalf("Reve sendBulk error: %v", err)
	}
	if !bulkResp.Successful || len(bulkResp.PerRecipientIds) != 2 {
		t.Fatalf("Unexpected bulk response: %+v", bulkResp)
	}

	dlr, err := driver.FetchDeliveryStatus(context.Background(), "REVE-101", "8801700000001")
	if err != nil {
		t.Fatalf("Reve DLR error: %v", err)
	}
	if dlr.Outcome != sms.OutcomeDelivered || dlr.DeliveredAt == nil {
		t.Errorf("Expected delivered with parsed date, got: %+v", dlr)
	}
}

func TestQSmsMockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "SendSMS") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ErrorCode": 0,
				"Data": []any{
					map[string]any{"MobileNumber": "8801700000001", "MessageId": "QSMS-999"},
				},
			})
			return
		}
		if strings.Contains(r.URL.Path, "MessageStatus") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ErrorCode": 0,
				"Data": map[string]any{
					"Status":   "REQUEST SUCCESS",
					"DoneDate": "20 Oct 2026 15:30:00",
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	driver := NewQSmsDriver(map[string]any{
		"base_url":  ts.URL,
		"api_key":   "k",
		"client_id": "c",
	})

	resp, err := driver.Send(context.Background(), sms.SmsRequest{
		To:   "8801700000001",
		Body: "QSms test",
		From: "WHITEINFO",
	})
	if err != nil {
		t.Fatalf("QSms Send error: %v", err)
	}
	if !resp.Successful || *resp.GatewayMessageId != "QSMS-999" {
		t.Errorf("QSms Send response mismatch: %+v", resp)
	}

	dlr, err := driver.FetchDeliveryStatus(context.Background(), "QSMS-999", "8801700000001")
	if err != nil {
		t.Fatalf("QSms DLR error: %v", err)
	}
	if dlr.Outcome != sms.OutcomeDelivered || dlr.DeliveredAt == nil {
		t.Errorf("QSms expected Delivered with date, got: %+v", dlr)
	}
}

func TestAnbernetMockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "sendsms") {
			// Anbernet returns array wrapped object
			_ = json.NewEncoder(w).Encode([]any{
				map[string]any{
					"status": "success",
					"messageids": []any{
						map[string]any{"receiver": "8801700000001", "messageid": "ANB-555"},
					},
				},
			})
			return
		}
		if strings.Contains(r.URL.Path, "dlrcheck") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":   "DELIVERED",
				"datetime": "2026-10-20 15:30:00",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	driver := NewAnbernetDriver(map[string]any{
		"base_url": ts.URL,
		"account":  "acc",
		"token":    "tok",
	})

	resp, err := driver.Send(context.Background(), sms.SmsRequest{
		To:   "8801700000001",
		Body: "Anbernet test",
		From: "WHITEINFO",
	})
	if err != nil {
		t.Fatalf("Anbernet Send error: %v", err)
	}
	if !resp.Successful || *resp.GatewayMessageId != "ANB-555" {
		t.Errorf("Anbernet Send mismatch: %+v", resp)
	}

	dlr, err := driver.FetchDeliveryStatus(context.Background(), "ANB-555", "8801700000001")
	if err != nil {
		t.Fatalf("Anbernet DLR error: %v", err)
	}
	if dlr.Outcome != sms.OutcomeDelivered || dlr.DeliveredAt == nil {
		t.Errorf("Anbernet expected Delivered, got: %+v", dlr)
	}
}

func TestWhiteInfoMockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer secret_token_xyz" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "Unauthenticated."})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "messages") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"batch_id": "BATCH-WHITE-777",
				"status":   "pending",
			})
			return
		}
		if strings.Contains(r.URL.Path, "batches/BATCH-WHITE-777") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "BATCH-WHITE-777",
				"status": "completed",
			})
			return
		}
		if strings.Contains(r.URL.Path, "balance") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"balance":  "1250.50",
				"currency": "BDT",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	driver := NewWhiteInfoDriver(map[string]any{
		"base_url":  ts.URL,
		"api_token": "secret_token_xyz",
		"sender_id": "FROLAX",
	})

	resp, err := driver.Send(context.Background(), sms.SmsRequest{
		To:   "01700000001",
		Body: "WhiteInfo test",
	})
	if err != nil {
		t.Fatalf("WhiteInfo Send error: %v", err)
	}
	if !resp.Successful || *resp.GatewayMessageId != "BATCH-WHITE-777" {
		t.Errorf("WhiteInfo Send mismatch: %+v", resp)
	}

	dlr, err := driver.FetchDeliveryStatus(context.Background(), "BATCH-WHITE-777", "01700000001")
	if err != nil {
		t.Fatalf("WhiteInfo DLR error: %v", err)
	}
	if dlr.Outcome != sms.OutcomeDelivered {
		t.Errorf("WhiteInfo expected Delivered, got: %s", dlr.Outcome)
	}

	dash, err := driver.GetDashboardData(context.Background())
	if err != nil {
		t.Fatalf("WhiteInfo Dashboard error: %v", err)
	}
	if len(dash.Cards) == 0 || dash.Cards[0].Value != "1250.50 BDT" {
		t.Errorf("WhiteInfo Dashboard mismatch: %+v", dash)
	}
}

func TestLogDriver(t *testing.T) {
	driver := NewLogDriver(nil)

	resp, err := driver.Send(context.Background(), sms.SmsRequest{
		To:   "8801700000001",
		Body: "Log test",
	})
	if err != nil || !resp.Successful || resp.GatewayMessageId == nil {
		t.Errorf("LogDriver Send failed: %+v, err: %v", resp, err)
	}

	bulkResp, err := driver.SendBulk(context.Background(), sms.BulkSmsRequest{
		Recipients: []string{"8801700000001", "8801700000002"},
		Body:       "Log bulk",
	})
	if err != nil || !bulkResp.Successful || len(bulkResp.PerRecipientIds) != 2 {
		t.Errorf("LogDriver SendBulk failed: %+v, err: %v", bulkResp, err)
	}
}
