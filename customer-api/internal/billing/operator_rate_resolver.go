package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

type OperatorRateNotConfiguredError struct {
	Phone        string
	OperatorName string
	IsMasking    bool
	Message      string
}

func (e *OperatorRateNotConfiguredError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	typeName := "non-masking"
	if e.IsMasking {
		typeName = "masking"
	}
	return fmt.Sprintf("No %s rate configured for operator [%s] on phone %s", typeName, e.OperatorName, e.Phone)
}

type Operator struct {
	ID              int64
	Name            string
	CountryID       int64
	CountryCode     string
	Codes           []string
	MaskingPrice    sql.NullFloat64
	NonMaskingPrice sql.NullFloat64
}

type RatePair struct {
	MaskingPrice    sql.NullFloat64
	NonMaskingPrice sql.NullFloat64
}

type OperatorRateResolver struct {
	operators     map[int64]Operator
	codeIndex     map[string]int64 // "countryCode:opCode" -> operatorID
	overrides     map[int64]RatePair
	hasActivePlan bool
	planName      string
}

// ForUser initializes an OperatorRateResolver for a specific user, loading active tariff plans or overrides.
func ForUser(ctx context.Context, db *sql.DB, userID int64) (*OperatorRateResolver, error) {
	// 1. Load active operators and their country codes
	rows, err := db.QueryContext(ctx, `
		SELECT o.id, o.name, o.country_id, COALESCE(c.code, '880'), o.codes, o.masking_price, o.non_masking_price
		FROM operators o
		LEFT JOIN countries c ON c.id = o.country_id AND c.is_active = 1
		WHERE o.is_active = 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query operators: %w", err)
	}
	defer rows.Close()

	operators := make(map[int64]Operator)
	codeIndex := make(map[string]int64)

	for rows.Next() {
		var op Operator
		var codesRaw []byte
		if err := rows.Scan(&op.ID, &op.Name, &op.CountryID, &op.CountryCode, &codesRaw, &op.MaskingPrice, &op.NonMaskingPrice); err != nil {
			return nil, err
		}

		if len(codesRaw) > 0 {
			_ = json.Unmarshal(codesRaw, &op.Codes)
		}

		operators[op.ID] = op

		for _, code := range op.Codes {
			key := fmt.Sprintf("%s:%s", op.CountryCode, strings.TrimSpace(code))
			codeIndex[key] = op.ID
		}
	}

	// 2. Check active tariff plan assignment
	var planID int64
	var planName string
	err = db.QueryRowContext(ctx, `
		SELECT tp.id, tp.name
		FROM tariff_plan_assignments tpa
		JOIN tariff_plans tp ON tp.id = tpa.tariff_plan_id
		WHERE tpa.user_id = ? AND tpa.is_active = 1
		LIMIT 1
	`, userID).Scan(&planID, &planName)

	overrides := make(map[int64]RatePair)

	if err == nil && planID > 0 {
		rateRows, rErr := db.QueryContext(ctx, `
			SELECT operator_id, masking_price, non_masking_price
			FROM tariff_rates
			WHERE tariff_plan_id = ?
		`, planID)
		if rErr == nil {
			defer rateRows.Close()
			for rateRows.Next() {
				var opID int64
				var rp RatePair
				if err := rateRows.Scan(&opID, &rp.MaskingPrice, &rp.NonMaskingPrice); err == nil {
					overrides[opID] = rp
				}
			}
		}

		return &OperatorRateResolver{
			operators:     operators,
			codeIndex:     codeIndex,
			overrides:     overrides,
			hasActivePlan: true,
			planName:      planName,
		}, nil
	}

	// 3. Fallback to legacy user_operator_rates table
	legacyRows, lErr := db.QueryContext(ctx, `
		SELECT operator_id, masking_price, non_masking_price
		FROM user_operator_rates
		WHERE user_id = ?
	`, userID)
	if lErr == nil {
		defer legacyRows.Close()
		for legacyRows.Next() {
			var opID int64
			var rp RatePair
			if err := legacyRows.Scan(&opID, &rp.MaskingPrice, &rp.NonMaskingPrice); err == nil {
				overrides[opID] = rp
			}
		}
	}

	return &OperatorRateResolver{
		operators:     operators,
		codeIndex:     codeIndex,
		overrides:     overrides,
		hasActivePlan: false,
	}, nil
}

func (r *OperatorRateResolver) MatchOperator(phone string) *Operator {
	digits := strings.TrimLeft(phone, "+")

	for key, opID := range r.codeIndex {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			continue
		}
		country := parts[0]
		code := parts[1]

		if strings.HasPrefix(digits, country) {
			rem := digits[len(country):]
			if strings.HasPrefix(rem, code) {
				if op, ok := r.operators[opID]; ok {
					return &op
				}
			}
		}
	}

	return nil
}

func (r *OperatorRateResolver) OperatorNameFor(phone string) string {
	op := r.MatchOperator(phone)
	if op != nil {
		return op.Name
	}
	return ""
}

func (r *OperatorRateResolver) PlanName() string {
	return r.planName
}

func (r *OperatorRateResolver) Resolve(phone string, isMasking bool) (float64, error) {
	op := r.MatchOperator(phone)
	if op == nil {
		return 0, &OperatorRateNotConfiguredError{
			Phone:     phone,
			IsMasking: isMasking,
			Message:   fmt.Sprintf("No operator matches phone %s", phone),
		}
	}

	typeName := "non-masking"
	if isMasking {
		typeName = "masking"
	}

	if r.hasActivePlan {
		if rp, ok := r.overrides[op.ID]; ok {
			if isMasking && rp.MaskingPrice.Valid {
				return rp.MaskingPrice.Float64, nil
			}
			if !isMasking && rp.NonMaskingPrice.Valid {
				return rp.NonMaskingPrice.Float64, nil
			}
		}

		planDesc := "assigned tariff plan"
		if r.planName != "" {
			planDesc = fmt.Sprintf("assigned tariff plan [%s]", r.planName)
		}

		return 0, &OperatorRateNotConfiguredError{
			Phone:        phone,
			OperatorName: op.Name,
			IsMasking:    isMasking,
			Message:      fmt.Sprintf("No %s rate configured in %s for operator [%s].", typeName, planDesc, op.Name),
		}
	}

	// Legacy override check
	if rp, ok := r.overrides[op.ID]; ok {
		if isMasking && rp.MaskingPrice.Valid {
			return rp.MaskingPrice.Float64, nil
		}
		if !isMasking && rp.NonMaskingPrice.Valid {
			return rp.NonMaskingPrice.Float64, nil
		}
	}

	// System default
	if isMasking && op.MaskingPrice.Valid {
		return op.MaskingPrice.Float64, nil
	}
	if !isMasking && op.NonMaskingPrice.Valid {
		return op.NonMaskingPrice.Float64, nil
	}

	return 0, &OperatorRateNotConfiguredError{
		Phone:        phone,
		OperatorName: op.Name,
		IsMasking:    isMasking,
	}
}

// RatesForUser returns all rates available across operators for affordability estimates.
func (r *OperatorRateResolver) RatesForUser(isMasking bool) []float64 {
	var rates []float64

	for _, op := range r.operators {
		if r.hasActivePlan {
			if rp, ok := r.overrides[op.ID]; ok {
				if isMasking && rp.MaskingPrice.Valid {
					rates = append(rates, rp.MaskingPrice.Float64)
				} else if !isMasking && rp.NonMaskingPrice.Valid {
					rates = append(rates, rp.NonMaskingPrice.Float64)
				}
			}
		} else {
			if rp, ok := r.overrides[op.ID]; ok {
				if isMasking && rp.MaskingPrice.Valid {
					rates = append(rates, rp.MaskingPrice.Float64)
					continue
				} else if !isMasking && rp.NonMaskingPrice.Valid {
					rates = append(rates, rp.NonMaskingPrice.Float64)
					continue
				}
			}

			if isMasking && op.MaskingPrice.Valid {
				rates = append(rates, op.MaskingPrice.Float64)
			} else if !isMasking && op.NonMaskingPrice.Valid {
				rates = append(rates, op.NonMaskingPrice.Float64)
			}
		}
	}

	return rates
}

// CalculateParts determines message part count and unicode encoding.
func CalculateParts(body string) (parts int, isUnicode bool) {
	runes := []rune(body)
	charCount := len(runes)

	// Check if any rune is non-GSM 03.38
	isUnicode = false
	for _, r := range runes {
		if !isGsm7Rune(r) {
			isUnicode = true
			break
		}
	}

	if charCount == 0 {
		return 1, false
	}

	if isUnicode {
		if charCount <= 70 {
			return 1, true
		}
		return int(math.Ceil(float64(charCount) / 67.0)), true
	}

	if charCount <= 160 {
		return 1, false
	}
	return int(math.Ceil(float64(charCount) / 153.0)), false
}

func isGsm7Rune(r rune) bool {
	// Standard GSM 03.38 Basic Character Set
	const gsm7Chars = "@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞ\x1bÆæßÉ !\"#¤%&'()*+,-./0123456789:;<=>?¡ABCDEFGHIJKLMNOPQRSTUVWXYZÄÖÑÜ§¿abcdefghijklmnopqrstuvwxyzäöñüà"
	if strings.ContainsRune(gsm7Chars, r) {
		return true
	}
	// Extended GSM 03.38 characters
	switch r {
	case '\f', '^', '{', '}', '\\', '[', '~', ']', '|', '€':
		return true
	}
	return false
}

// RoundFloat rounds a float to n decimal places.
func RoundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
