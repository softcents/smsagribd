package drivers

import (
	"fmt"
	"strings"

	"smswhiteinfo/customer-api/internal/sms"
)

// CreateDriver instantiates an SMS gateway driver from its driver name and decrypted credentials.
func CreateDriver(driver string, credentials map[string]any) (sms.SmsGateway, error) {
	d := strings.ToLower(strings.TrimSpace(driver))

	switch d {
	case "infozillion":
		return NewInfozillionDriver(credentials), nil
	case "revesms":
		return NewReveSmsDriver(credentials), nil
	case "qsms":
		return NewQSmsDriver(credentials), nil
	case "anbernet":
		return NewAnbernetDriver(credentials), nil
	case "whiteinfo":
		return NewWhiteInfoDriver(credentials), nil
	case "log", "mock":
		return NewLogDriver(credentials), nil
	default:
		return nil, fmt.Errorf("unsupported SMS gateway driver: %s", driver)
	}
}
