package kyc

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

type KycService struct {
	db          *sql.DB
	storageRoot string
}

func NewKycService(db *sql.DB, storageRoot string) *KycService {
	if storageRoot == "" {
		storageRoot = "./storage/app/private"
	}
	return &KycService{
		db:          db,
		storageRoot: storageRoot,
	}
}

func (s *KycService) CreatePhonePhotoSession(c fiber.Ctx) error {
	sessionID := uuid.New().String()
	sessionToken := uuid.New().String() + "-" + uuid.New().String()
	expiresAt := time.Now().Add(30 * time.Minute)

	_, err := s.db.ExecContext(c.Context(), `
		INSERT INTO phone_photo_sessions (id, session_token, status, expires_at, created_at, updated_at)
		VALUES (?, ?, 'pending', ?, NOW(), NOW())
	`, sessionID, sessionToken, expiresAt)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
	}

	captureURL := fmt.Sprintf("%s/phone-photo/%s", c.BaseURL(), sessionToken)

	pngBytes, err := qrcode.Encode(captureURL, qrcode.Medium, 220)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate QR code"})
	}

	qrDataURI := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(pngBytes))

	return c.JSON(fiber.Map{
		"success":       true,
		"session_token": sessionToken,
		"capture_url":   captureURL,
		"qr_data_uri":   qrDataURI,
		"expires_at":    expiresAt.Format(time.RFC3339),
	})
}

func (s *KycService) ShowCapturePage(c fiber.Ctx) error {
	token := c.Params("token")

	var status string
	var expiresAt time.Time
	err := s.db.QueryRowContext(c.Context(), `
		SELECT status, expires_at 
		FROM phone_photo_sessions 
		WHERE session_token = ?
	`, token).Scan(&status, &expiresAt)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("Invalid or expired photo session.")
	}

	if time.Now().After(expiresAt) {
		return c.Status(http.StatusGone).SendString("This photo session has expired. Please start over from your registration page.")
	}

	if status == "completed" {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(`
			<!DOCTYPE html>
			<html><head><title>Photo Already Uploaded</title><meta name="viewport" content="width=device-width, initial-scale=1"></head>
			<body style="font-family:sans-serif;text-align:center;padding:40px;">
				<h2>Photo Already Uploaded</h2>
				<p>Your photo has already been uploaded for this session. Please return to your computer to finish registration.</p>
			</body></html>
		`)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Capture Photo - SMS WhiteInfo</title>
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<style>
				body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; padding: 20px; text-align: center; background: #f9fafb; }
				.card { background: white; border-radius: 12px; padding: 24px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); max-width: 400px; margin: 40px auto; }
				.btn { background: #4f46e5; color: white; border: none; padding: 12px 24px; border-radius: 8px; font-weight: 600; cursor: pointer; display: inline-block; margin-top: 15px; }
				#preview { width: 100%%; max-height: 250px; object-fit: cover; border-radius: 8px; margin-top: 15px; display: none; }
			</style>
		</head>
		<body>
			<div class="card">
				<h3>Take Photo for Verification</h3>
				<p style="color:#6b7280; font-size:14px;">Take a clear photo or selfie to complete your verification.</p>
				<input type="file" id="fileInput" accept="image/*" capture="user" style="display:none;" />
				<button class="btn" onclick="document.getElementById('fileInput').click()">Take Photo</button>
				<img id="preview" />
				<button id="uploadBtn" class="btn" style="background:#059669; display:none;" onclick="upload()">Confirm & Upload</button>
				<p id="msg" style="margin-top:15px; font-weight:bold;"></p>
			</div>
			<script>
				let photoData = "";
				document.getElementById('fileInput').addEventListener('change', function(e) {
					const file = e.target.files[0];
					if (file) {
						const reader = new FileReader();
						reader.onload = function(evt) {
							photoData = evt.target.result;
							const preview = document.getElementById('preview');
							preview.src = photoData;
							preview.style.display = 'block';
							document.getElementById('uploadBtn').style.display = 'inline-block';
						};
						reader.readAsDataURL(file);
					}
				});
				function upload() {
					document.getElementById('msg').innerText = "Uploading...";
					fetch('/phone-photo/%s/upload', {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ photo: photoData })
					})
					.then(r => r.json())
					.then(data => {
						if (data.success) {
							document.getElementById('msg').style.color = "#059669";
							document.getElementById('msg').innerText = "Uploaded successfully! You can now return to your computer.";
							document.getElementById('uploadBtn').style.display = 'none';
						} else {
							document.getElementById('msg').style.color = "#dc2626";
							document.getElementById('msg').innerText = data.message || "Upload failed.";
						}
					})
					.catch(() => {
						document.getElementById('msg').style.color = "#dc2626";
						document.getElementById('msg').innerText = "Network error. Please try again.";
					});
				}
			</script>
		</body>
		</html>
	`, token)

	return c.SendString(html)
}

func (s *KycService) UploadPhoto(c fiber.Ctx) error {
	token := c.Params("token")

	var req struct {
		Photo string `json:"photo"`
	}
	if err := c.Bind().JSON(&req); err != nil || req.Photo == "" {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"message": "Photo data is required.",
		})
	}

	if !strings.HasPrefix(req.Photo, "data:image") {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"message": "Invalid photo format.",
		})
	}

	var status string
	var expiresAt time.Time
	err := s.db.QueryRowContext(c.Context(), `
		SELECT status, expires_at 
		FROM phone_photo_sessions 
		WHERE session_token = ?
	`, token).Scan(&status, &expiresAt)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Invalid session.",
		})
	}

	if time.Now().After(expiresAt) {
		return c.Status(http.StatusGone).JSON(fiber.Map{
			"success": false,
			"message": "Session expired.",
		})
	}

	if status == "completed" {
		return c.JSON(fiber.Map{
			"success":           true,
			"message":           "Photo already uploaded.",
			"already_completed": true,
		})
	}

	_, err = s.db.ExecContext(c.Context(), `
		UPDATE phone_photo_sessions 
		SET photo_data = ?, status = 'completed', updated_at = NOW() 
		WHERE session_token = ?
	`, req.Photo, token)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to save photo.",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Photo uploaded successfully! You can now return to your computer to continue registration.",
	})
}

func (s *KycService) CheckStatus(c fiber.Ctx) error {
	token := c.Params("token")

	var status string
	var photoData sql.NullString
	var expiresAt time.Time

	err := s.db.QueryRowContext(c.Context(), `
		SELECT status, photo_data, expires_at 
		FROM phone_photo_sessions 
		WHERE session_token = ?
	`, token).Scan(&status, &photoData, &expiresAt)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"status":  "not_found",
			"message": "Invalid session.",
		})
	}

	if time.Now().After(expiresAt) {
		return c.JSON(fiber.Map{
			"success": true,
			"status":  "expired",
			"message": "Session has expired.",
		})
	}

	if status == "completed" {
		return c.JSON(fiber.Map{
			"success":    true,
			"status":     "completed",
			"photo_data": photoData.String,
		})
	}

	expiresInSec := int(time.Until(expiresAt).Seconds())
	if expiresInSec < 0 {
		expiresInSec = 0
	}

	return c.JSON(fiber.Map{
		"success":            true,
		"status":             "pending",
		"message":            "Waiting for photo upload.",
		"expires_in_seconds": expiresInSec,
	})
}

func (s *KycService) StorageRoot() string {
	return s.storageRoot
}

func (s *KycService) StreamKycFile(c fiber.Ctx) error {
	userParam := c.Params("user")
	field := c.Params("field")

	allowedFields := map[string]bool{
		"front_nid":          true,
		"back_nid":           true,
		"live_photo":         true,
		"trade_license_file": true,
	}

	if !allowedFields[field] {
		return c.Status(http.StatusNotFound).SendString("File not found.")
	}

	var relPath sql.NullString
	err := s.db.QueryRowContext(c.Context(), fmt.Sprintf("SELECT %s FROM users WHERE id = ?", field), userParam).Scan(&relPath)
	if err != nil || !relPath.Valid || relPath.String == "" {
		return c.Status(http.StatusNotFound).SendString("File not found.")
	}

	rawPath := relPath.String
	// Prevent path traversal
	if strings.Contains(rawPath, "..") {
		return c.Status(http.StatusNotFound).SendString("File not found.")
	}

	// Normalize if path in DB includes legacy "storage/app/private/" or "app/private/"
	normalized := strings.TrimPrefix(rawPath, "storage/app/private/")
	normalized = strings.TrimPrefix(normalized, "app/private/")
	normalized = strings.TrimPrefix(normalized, "/")

	fullPath := normalized
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(s.storageRoot, normalized)
	}

	// Verify that the resolved path is within storageRoot to prevent traversal
	cleanRoot := filepath.Clean(s.storageRoot)
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, cleanRoot) {
		return c.Status(http.StatusNotFound).SendString("File not found.")
	}

	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return c.Status(http.StatusNotFound).SendString("File not found on storage.")
	}

	c.Set("Cache-Control", "private, no-store")
	c.Set("X-Content-Type-Options", "nosniff")

	lower := strings.ToLower(cleanPath)
	if strings.HasSuffix(lower, ".pdf") {
		c.Set("Content-Type", "application/pdf")
	} else if strings.HasSuffix(lower, ".png") {
		c.Set("Content-Type", "image/png")
	} else if strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
		c.Set("Content-Type", "image/jpeg")
	}

	return c.SendFile(cleanPath)
}

func (s *KycService) DownloadFailedReport(c fiber.Ctx, userID int64, isAdmin bool) error {
	batchParam := c.Params("batch")
	filename := c.Params("filename")

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return c.Status(http.StatusNotFound).SendString("File not found.")
	}

	var batchOwner int64
	var reportPath sql.NullString
	err := s.db.QueryRowContext(c.Context(), "SELECT user_id, failed_report_path FROM batches WHERE id = ?", batchParam).Scan(&batchOwner, &reportPath)
	if err != nil || !reportPath.Valid || reportPath.String == "" {
		return c.Status(http.StatusNotFound).SendString("Report not found.")
	}

	if !isAdmin && batchOwner != userID {
		return c.Status(http.StatusForbidden).SendString("Access denied.")
	}

	expectedBase := filepath.Base(reportPath.String)
	if expectedBase != filename {
		return c.Status(http.StatusNotFound).SendString("Report not found.")
	}

	fullPath := reportPath.String
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(s.storageRoot, fullPath)
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return c.Status(http.StatusNotFound).SendString("Report file does not exist on disk.")
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.SendFile(fullPath)
}
