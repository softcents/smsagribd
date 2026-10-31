package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"smswhiteinfo/customer-api/internal/sms"
)

var (
	nonDigitRegex = regexp.MustCompile(`\D+`)
	hasAlphaRegex = regexp.MustCompile(`[a-zA-Z]`)
)

type InfozillionDriver struct {
	credentials map[string]any
	client      *sms.SafeHTTPClient
}

func NewInfozillionDriver(credentials map[string]any) *InfozillionDriver {
	baseURL := getString(credentials, "base_url", "https://api.mnpspbd.com")
	return &InfozillionDriver{
		credentials: credentials,
		client:      sms.NewSafeHTTPClient(baseURL, 15*time.Second),
	}
}

func (d *InfozillionDriver) Send(ctx context.Context, req sms.SmsRequest) (sms.SmsResponse, error) {
	payload := d.BuildSendPayload([]string{req.To}, req.Body, req.From, req.IsUnicode)
	endpoint := d.resolveEndpoint("send-sms")

	respBytes, statusCode, err := d.client.PostJSON(ctx, endpoint, payload, nil)
	if err != nil {
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	serverCode := fmt.Sprintf("%v", data["serverResponseCode"])
	if statusCode >= 400 || serverCode != "9000" {
		errMsg := getString(data, "serverResponseMessage", fmt.Sprintf("HTTP %d error", statusCode))
		return sms.SmsResponse{
			Successful:   false,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}

	var txnID *string
	if val, ok := data["serverTxnId"].(string); ok && val != "" {
		txnID = &val
	}

	return sms.SmsResponse{
		Successful:       true,
		GatewayMessageId: txnID,
		Raw:              data,
	}, nil
}

func (d *InfozillionDriver) SendBulk(ctx context.Context, req sms.BulkSmsRequest) (sms.BulkSmsResponse, error) {
	payload := d.BuildSendPayload(req.Recipients, req.Body, req.From, req.IsUnicode)
	endpoint := d.resolveEndpoint("send-sms")

	respBytes, statusCode, err := d.client.PostJSON(ctx, endpoint, payload, nil)
	if err != nil {
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	serverCode := fmt.Sprintf("%v", data["serverResponseCode"])
	if statusCode >= 400 || serverCode != "9000" {
		errMsg := getString(data, "serverResponseMessage", fmt.Sprintf("HTTP %d error", statusCode))
		return sms.BulkSmsResponse{
			Successful:   false,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}

	var txnID *string
	if val, ok := data["serverTxnId"].(string); ok && val != "" {
		txnID = &val
	}

	failedRecipients := d.extractFailedRecipients(data, req.Recipients)

	perRecipient := make([]string, len(req.Recipients))
	for i := range req.Recipients {
		if txnID != nil {
			perRecipient[i] = *txnID
		}
	}

	return sms.BulkSmsResponse{
		Successful:       true,
		GatewayBatchId:   txnID,
		PerRecipientIds:  perRecipient,
		FailedRecipients: failedRecipients,
		Raw:              data,
	}, nil
}

func (d *InfozillionDriver) FetchDeliveryStatus(ctx context.Context, gatewayMessageId string, recipient string) (sms.DeliveryStatusResponse, error) {
	dlrPayload := map[string]any{
		"username":            getString(d.credentials, "username", ""),
		"password":            getString(d.credentials, "password", ""),
		"billMsisdn":          getString(d.credentials, "bill_msisdn", ""),
		"usernameSecondary":   getString(d.credentials, "username_secondary", ""),
		"passwordSecondary":   getString(d.credentials, "password_secondary", ""),
		"billMsisdnSecondary": getString(d.credentials, "bill_msisdn_secondary", ""),
		"apiKey":              getString(d.credentials, "api_key", ""),
		"serverReference":     gatewayMessageId,
	}

	if recipient != "" {
		dlrPayload["msisdnList"] = []string{strings.TrimPrefix(recipient, "+")}
	} else {
		dlrPayload["msisdnList"] = []string{}
	}

	endpoint := d.resolveEndpoint("check-delivery-report")
	respBytes, statusCode, err := d.client.PostJSON(ctx, endpoint, dlrPayload, nil)
	if err != nil {
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			ErrorMessage: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	var data map[string]any
	_ = json.Unmarshal(respBytes, &data)

	serverCode := fmt.Sprintf("%v", data["serverResponseCode"])
	if statusCode >= 400 || serverCode != "9000" {
		errMsg := getString(data, "serverResponseMessage", "Delivery check failed")
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeInTransit,
			ErrorMessage: errMsg,
			Raw:          data,
		}, nil
	}

	return d.mapDeliveryStatus(data), nil
}

func (d *InfozillionDriver) GetDashboardData(ctx context.Context) (sms.DashboardData, error) {
	cards := make([]sms.StatCard, 0)

	// Fetch current wallet balance
	walletVal := d.fetchBalance(ctx, "check-current-balance", map[string]any{
		"apiKey": getString(d.credentials, "api_key", ""),
	})

	walletCard := sms.StatCard{
		Label:       "Wallet Balance",
		Value:       "N/A",
		Description: "Available balance at CP end",
		Color:       "danger",
		Icon:        "heroicon-o-banknotes",
	}
	if walletVal != "" {
		walletCard.Value = walletVal
		walletCard.Color = "success"
	}
	cards = append(cards, walletCard)

	// Carrier balances
	carrierCodes := d.dashboardCarrierCodes()
	for code, name := range carrierCodes {
		balance := d.fetchBalance(ctx, "check-credit-balance", map[string]any{
			"username": getString(d.credentials, "username", ""),
			"password": getString(d.credentials, "password", ""),
			"mno":      code,
			"apiKey":   getString(d.credentials, "api_key", ""),
		})

		val := "N/A"
		if balance != "" {
			val = balance
		}
		cards = append(cards, sms.StatCard{
			Label:       fmt.Sprintf("%s Balance", name),
			Value:       val,
			Description: fmt.Sprintf("%s carrier credit", code),
			Icon:        "heroicon-o-device-phone-mobile",
		})
	}

	return sms.DashboardData{Cards: cards}, nil
}

func (d *InfozillionDriver) fetchBalance(ctx context.Context, endpoint string, payload map[string]any) string {
	respBytes, statusCode, err := d.client.PostJSON(ctx, d.resolveEndpoint(endpoint), payload, nil)
	if err != nil || statusCode >= 400 {
		return ""
	}
	var data map[string]any
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return ""
	}
	if fmt.Sprintf("%v", data["serverResponseCode"]) != "9000" {
		return ""
	}
	return getString(data, "availableBalance", "")
}

// BuildSendPayload builds the exact upstream Infozillion JSON request payload,
// enforcing the live hotfix rule for numeric CLIs.
func (d *InfozillionDriver) BuildSendPayload(recipients []string, body string, senderId string, isUnicode bool) map[string]any {
	configuredType := getString(d.credentials, "transaction_type", "")
	transactionType := configuredType
	if transactionType == "" {
		if len(recipients) > 1 {
			transactionType = "P"
		} else {
			transactionType = "T"
		}
	}

	messageType := getString(d.credentials, "message_type", "1")
	if isUnicode {
		messageType = "3"
	}

	isLongSMS := false
	if v, ok := d.credentials["is_long_sms"]; ok {
		isLongSMS = fmt.Sprintf("%v", v) == "true" || fmt.Sprintf("%v", v) == "1"
	} else if len([]rune(body)) > 160 {
		isLongSMS = true
	}

	cli := strings.TrimSpace(senderId)
	if cli == "" {
		cli = strings.TrimSpace(getString(d.credentials, "cli", ""))
	}
	if cli == "" {
		cli = strings.TrimSpace(getString(d.credentials, "bill_msisdn", ""))
	}

	// Hotfix: Numeric CLI (digits only, no letters):
	// CLI and billMsisdn are ALWAYS the exact same normalized number.
	cliDigits := nonDigitRegex.ReplaceAllString(cli, "")
	isNumericCli := cli != "" && cliDigits != "" && !hasAlphaRegex.MatchString(cli)

	var billMsisdn string
	if isNumericCli {
		if strings.HasPrefix(cliDigits, "0") {
			cliDigits = "880" + cliDigits[1:]
		} else if !strings.HasPrefix(cliDigits, "880") {
			cliDigits = "880" + cliDigits
		}
		cli = cliDigits
		billMsisdn = cliDigits
	} else {
		// Masking / alphanumeric CLI: keep the gateway's configured bill_msisdn.
		billMsisdn = strings.TrimSpace(getString(d.credentials, "bill_msisdn", ""))
	}

	msisdnList := make([]string, len(recipients))
	for i, r := range recipients {
		msisdnList[i] = strings.TrimPrefix(strings.TrimSpace(r), "+")
	}

	payload := map[string]any{
		"username":        getString(d.credentials, "username", ""),
		"password":        getString(d.credentials, "password", ""),
		"billMsisdn":      billMsisdn,
		"apiKey":          getString(d.credentials, "api_key", ""),
		"cli":             cli,
		"msisdnList":      msisdnList,
		"transactionType": transactionType,
		"messageType":     messageType,
		"isLongSMS":       isLongSMS,
		"message":         body,
	}

	campaignId := getString(d.credentials, "campaign_id", "")
	if transactionType == "P" {
		if campaignId != "" {
			payload["campaignId"] = campaignId
		} else {
			payload["campaignId"] = "camp-" + time.Now().Format("20060102150405")
		}
	} else if campaignId != "" {
		payload["campaignId"] = campaignId
	}

	if secUser := getString(d.credentials, "username_secondary", ""); secUser != "" {
		payload["usernameSecondary"] = secUser
		payload["passwordSecondary"] = getString(d.credentials, "password_secondary", "")
		payload["billMsisdnSecondary"] = getString(d.credentials, "bill_msisdn_secondary", "")
	}

	return payload
}

func (d *InfozillionDriver) mapDeliveryStatus(data map[string]any) sms.DeliveryStatusResponse {
	// 1. DND filtering
	if dnd, ok := data["dndMsisdn"].([]any); ok && len(dnd) > 0 {
		var numbers []string
		for _, n := range dnd {
			numbers = append(numbers, fmt.Sprintf("%v", n))
		}
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeFailed,
			ErrorMessage: "Recipient filtered due to DND enlistment: " + strings.Join(numbers, ", "),
			Raw:          data,
		}
	}

	// 2. Invalid MSISDN
	if inv, ok := data["invalidMsisdn"].([]any); ok && len(inv) > 0 {
		var numbers []string
		for _, n := range inv {
			numbers = append(numbers, fmt.Sprintf("%v", n))
		}
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeFailed,
			ErrorMessage: "Invalid MSISDN prefix or format: " + strings.Join(numbers, ", "),
			Raw:          data,
		}
	}

	// 3. deliveryStatus array
	if delStatuses, ok := data["deliveryStatus"].([]any); ok && len(delStatuses) > 0 {
		firstStatus := strings.TrimSpace(fmt.Sprintf("%v", delStatuses[0]))
		lower := strings.ToLower(firstStatus)

		if strings.Contains(lower, "undelivered") || strings.Contains(lower, "failed") || strings.Contains(lower, "rejected") || strings.Contains(lower, "expired") {
			return sms.DeliveryStatusResponse{
				Outcome:      sms.OutcomeFailed,
				RawStatus:    firstStatus,
				ErrorMessage: "SMS delivery status: " + firstStatus,
				Raw:          data,
			}
		}

		if strings.Contains(lower, "delivered") {
			now := time.Now()
			return sms.DeliveryStatusResponse{
				Outcome:     sms.OutcomeDelivered,
				RawStatus:   firstStatus,
				DeliveredAt: &now,
				Raw:         data,
			}
		}

		return sms.DeliveryStatusResponse{
			Outcome:   sms.OutcomeInTransit,
			RawStatus: firstStatus,
			Raw:       data,
		}
	}

	// 4. a2pDeliveryStatus fallback
	a2pStatus := getString(data, "a2pDeliveryStatus", "")
	a2pLower := strings.ToLower(a2pStatus)
	httpStatus := strings.ToLower(getString(data, "ansSendSmsHttpStatus", ""))

	if strings.Contains(a2pLower, "failure") || strings.Contains(a2pLower, "failed") || strings.Contains(httpStatus, "bad_gateway") || strings.HasPrefix(httpStatus, "4") || strings.HasPrefix(httpStatus, "5") {
		errMsg := getString(data, "serverResponseMessage", a2pStatus)
		return sms.DeliveryStatusResponse{
			Outcome:      sms.OutcomeFailed,
			RawStatus:    a2pStatus,
			ErrorMessage: errMsg,
			Raw:          data,
		}
	}

	return sms.DeliveryStatusResponse{
		Outcome:   sms.OutcomeInTransit,
		RawStatus: a2pStatus,
		Raw:       data,
	}
}

func (d *InfozillionDriver) extractFailedRecipients(data map[string]any, recipients []string) map[string]string {
	failed := make(map[string]string)
	known := make(map[string]bool)
	for _, r := range recipients {
		norm := normalizeMsisdn(r)
		if norm != "" {
			known[norm] = true
		}
	}

	fields := map[string]string{
		"dndMsisdn":     "Recipient is on the DND list.",
		"invalidMsisdn": "Recipient MSISDN is invalid or unsupported.",
	}

	for f, reason := range fields {
		if list, ok := data[f].([]any); ok {
			for _, item := range list {
				norm := normalizeMsisdn(fmt.Sprintf("%v", item))
				if norm != "" && known[norm] {
					failed[norm] = reason
				}
			}
		}
	}

	return failed
}

func (d *InfozillionDriver) resolveEndpoint(action string) string {
	apiType := strings.ToLower(strings.TrimSpace(getString(d.credentials, "api_type", "mno")))

	switch action {
	case "send-sms":
		if apiType == "iptsp" {
			return "a2p-sms-iptsp/api/v1/send-sms"
		}
		return "a2p-sms/api/v1/send-sms"
	case "check-delivery-report":
		if apiType == "iptsp" {
			return "a2p-proxy-api-iptsp/api/v1/check-delivery-report"
		}
		return "a2p-proxy-api/api/v1/check-delivery-report"
	case "check-credit-balance":
		if apiType == "iptsp" {
			return "a2p-proxy-api-iptsp/api/v1/check-credit-balance"
		}
		return "a2p-proxy-api/api/v1/check-credit-balance"
	case "check-current-balance":
		return "a2p-wallet/api/v1/check-current-balance"
	default:
		return action
	}
}

func (d *InfozillionDriver) dashboardCarrierCodes() map[string]string {
	apiType := strings.ToLower(strings.TrimSpace(getString(d.credentials, "api_type", "mno")))
	if apiType == "iptsp" {
		return map[string]string{
			"ADN": "ADN", "MN": "Mirnet", "BN": "Bracnet", "AIT": "AmberIT",
			"RT": "RanksITT", "PRM": "Premium", "FN": "FusionNet", "BR": "Brilliant",
			"MTN": "Metronet", "RCO": "RaceOnline", "BPL": "BanglaPhone", "ICN": "Icon",
			"WEB": "Weblink", "BDC": "BDCOM", "RDT": "RedData", "L3": "Link3",
			"BTCL": "BTCL", "AGNI": "AGNI", "ICC": "ICC", "RG": "RoyalGreen",
		}
	}
	return map[string]string{
		"GP": "GrameenPhone",
		"RB": "Robi",
		"TT": "Teletalk",
		"BL": "Banglalink",
	}
}

func getString(m map[string]any, key string, def string) string {
	if v, ok := m[key]; ok && v != nil {
		s := fmt.Sprintf("%v", v)
		if s != "" {
			return s
		}
	}
	return def
}

func normalizeMsisdn(value string) string {
	digits := nonDigitRegex.ReplaceAllString(value, "")
	if strings.HasPrefix(digits, "0") {
		digits = "880" + digits[1:]
	} else if !strings.HasPrefix(digits, "880") {
		digits = "880" + digits
	}
	return digits
}
