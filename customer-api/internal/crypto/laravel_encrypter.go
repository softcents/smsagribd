package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// LaravelEncrypter handles encryption and decryption compatible with
// Laravel's Illuminate\Encryption\Encrypter (AES-256-CBC with HMAC-SHA256).
type LaravelEncrypter struct {
	key []byte
}

// NewLaravelEncrypter creates a new encrypter from a Laravel APP_KEY.
// The key can be in "base64:<key>" format or raw 32 bytes.
func NewLaravelEncrypter(appKey string) (*LaravelEncrypter, error) {
	appKey = strings.TrimSpace(appKey)
	if strings.HasPrefix(appKey, "base64:") {
		k, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(appKey, "base64:"))
		if err != nil {
			return nil, fmt.Errorf("invalid base64 key: %w", err)
		}
		if len(k) != 32 {
			return nil, fmt.Errorf("key length must be 32 bytes, got %d", len(k))
		}
		return &LaravelEncrypter{key: k}, nil
	}

	k := []byte(appKey)
	if len(k) != 32 {
		return nil, fmt.Errorf("key length must be 32 bytes, got %d", len(k))
	}
	return &LaravelEncrypter{key: k}, nil
}

type laravelPayload struct {
	IV    string `json:"iv"`
	Value string `json:"value"`
	MAC   string `json:"mac"`
	Tag   string `json:"tag"`
}

// Decrypt decrypts a base64-encoded or raw JSON Laravel payload.
func (e *LaravelEncrypter) Decrypt(encrypted string) ([]byte, error) {
	encrypted = strings.TrimSpace(encrypted)
	jsonBytes, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		// Fallback: try raw JSON if not outer-base64 encoded
		jsonBytes = []byte(encrypted)
	}

	var payload laravelPayload
	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Verify HMAC-SHA256 over IV + Value
	mac := hmac.New(sha256.New, e.key)
	mac.Write([]byte(payload.IV + payload.Value))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedMAC), []byte(payload.MAC)) {
		return nil, errors.New("MAC mismatch: payload was tampered with or key is incorrect")
	}

	iv, err := base64.StdEncoding.DecodeString(payload.IV)
	if err != nil {
		return nil, fmt.Errorf("invalid IV base64: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(payload.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid Value base64: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	if len(plaintext) == 0 {
		return nil, errors.New("decrypted plaintext is empty")
	}
	padding := int(plaintext[len(plaintext)-1])
	if padding < 1 || padding > aes.BlockSize || padding > len(plaintext) {
		return nil, errors.New("invalid PKCS7 padding")
	}
	for i := len(plaintext) - padding; i < len(plaintext); i++ {
		if int(plaintext[i]) != padding {
			return nil, errors.New("invalid PKCS7 padding bytes")
		}
	}
	plaintext = plaintext[:len(plaintext)-padding]

	return plaintext, nil
}

// DecryptString decrypts and returns as string.
func (e *LaravelEncrypter) DecryptString(encrypted string) (string, error) {
	data, err := e.Decrypt(encrypted)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DecryptArray decrypts and parses a PHP serialized or JSON map.
func (e *LaravelEncrypter) DecryptArray(encrypted string) (map[string]any, error) {
	data, err := e.Decrypt(encrypted)
	if err != nil {
		return nil, err
	}
	return ParsePhpOrJson(data)
}

// Encrypt encrypts data into Laravel's standard AES-256-CBC base64 envelope.
func (e *LaravelEncrypter) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Add PKCS7 padding
	padding := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	paddedPlaintext := append(plaintext, padtext...)

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	b64IV := base64.StdEncoding.EncodeToString(iv)
	b64Value := base64.StdEncoding.EncodeToString(ciphertext)

	mac := hmac.New(sha256.New, e.key)
	mac.Write([]byte(b64IV + b64Value))
	hexMAC := hex.EncodeToString(mac.Sum(nil))

	payload := laravelPayload{
		IV:    b64IV,
		Value: b64Value,
		MAC:   hexMAC,
		Tag:   "",
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

// EncryptArray serializes a map as JSON and encrypts it in Laravel format.
func (e *LaravelEncrypter) EncryptArray(m map[string]any) (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal map: %w", err)
	}
	return e.Encrypt(data)
}

// ParsePhpOrJson parses either JSON or PHP serialized data into map[string]any.
func ParsePhpOrJson(data []byte) (map[string]any, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		var res map[string]any
		if err := json.Unmarshal(trimmed, &res); err == nil {
			return res, nil
		}
	}

	p := &phpUnserializer{buf: trimmed, pos: 0}
	val, err := p.unserialize()
	if err != nil {
		return nil, fmt.Errorf("php unserialize error: %w", err)
	}
	resMap, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", val)
	}
	return resMap, nil
}

type phpUnserializer struct {
	buf []byte
	pos int
}

func (p *phpUnserializer) unserialize() (any, error) {
	if p.pos >= len(p.buf) {
		return nil, errors.New("unexpected EOF")
	}
	tp := p.buf[p.pos]
	p.pos++
	if p.pos >= len(p.buf) || (p.buf[p.pos] != ':' && tp != 'N') {
		if tp == 'N' && p.pos < len(p.buf) && p.buf[p.pos] == ';' {
			p.pos++
			return nil, nil
		}
		return nil, fmt.Errorf("malformed type at pos %d", p.pos)
	}
	if tp != 'N' {
		p.pos++ // skip ':'
	}

	switch tp {
	case 's':
		lenStr, err := p.readUntil(':')
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(lenStr)
		if err != nil {
			return nil, err
		}
		if p.pos >= len(p.buf) || p.buf[p.pos] != '"' {
			return nil, errors.New("missing opening quote for string")
		}
		p.pos++
		if p.pos+length > len(p.buf) {
			return nil, errors.New("string length exceeds buffer")
		}
		val := string(p.buf[p.pos : p.pos+length])
		p.pos += length
		if p.pos >= len(p.buf) || p.buf[p.pos] != '"' {
			return nil, errors.New("missing closing quote for string")
		}
		p.pos++
		if p.pos >= len(p.buf) || p.buf[p.pos] != ';' {
			return nil, errors.New("missing semicolon after string")
		}
		p.pos++
		return val, nil

	case 'i':
		valStr, err := p.readUntil(';')
		if err != nil {
			return nil, err
		}
		return strconv.ParseInt(valStr, 10, 64)

	case 'd':
		valStr, err := p.readUntil(';')
		if err != nil {
			return nil, err
		}
		return strconv.ParseFloat(valStr, 64)

	case 'b':
		valStr, err := p.readUntil(';')
		if err != nil {
			return nil, err
		}
		return valStr == "1", nil

	case 'a':
		countStr, err := p.readUntil(':')
		if err != nil {
			return nil, err
		}
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, err
		}
		if p.pos >= len(p.buf) || p.buf[p.pos] != '{' {
			return nil, errors.New("missing opening brace for array")
		}
		p.pos++
		m := make(map[string]any, count)
		for i := 0; i < count; i++ {
			k, err := p.unserialize()
			if err != nil {
				return nil, fmt.Errorf("failed reading array key %d: %w", i, err)
			}
			v, err := p.unserialize()
			if err != nil {
				return nil, fmt.Errorf("failed reading array val for key %v: %w", k, err)
			}
			m[fmt.Sprintf("%v", k)] = v
		}
		if p.pos >= len(p.buf) || p.buf[p.pos] != '}' {
			return nil, errors.New("missing closing brace for array")
		}
		p.pos++
		return m, nil

	case 'N':
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported type '%c' at pos %d", tp, p.pos)
	}
}

func (p *phpUnserializer) readUntil(ch byte) (string, error) {
	idx := bytes.IndexByte(p.buf[p.pos:], ch)
	if idx == -1 {
		return "", errors.New("delimiter not found")
	}
	res := string(p.buf[p.pos : p.pos+idx])
	p.pos += idx + 1
	return res, nil
}
