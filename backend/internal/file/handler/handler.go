package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/aatist/backend/internal/file/model"
	"github.com/aatist/backend/internal/file/service"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	maxFileSize = 50 * 1024 * 1024 // 50MB default max size
	maxMemory   = 32 << 20         // 32MB for multipart form parsing
)

var allowedImageTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/gif":  {},
}

var allowedDocumentTypes = map[string]struct{}{
	"application/pdf": {},
	"text/plain":      {},
}

// FileHandler handles file-related HTTP requests
type FileHandler struct {
	fileService service.FileService
	logger      *log.Logger
}

// NewFileHandler creates a new file handler
func NewFileHandler(fileService service.FileService, logger *log.Logger) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		logger:      logger,
	}
}

// UploadFileHandler handles file uploads
// POST /api/v1/files/upload (user-facing) or POST /api/v1/internal/file/upload (internal)
func (h *FileHandler) UploadFileHandler(c *gin.Context) {
	// Get userID from context (set by TrustGatewayMiddleware or RequireGatewayAuth)
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.CodeUnauthorized, "unauthorized")
		return
	}

	// Parse file type from query parameter
	fileTypeStr := c.Query("type")
	if fileTypeStr == "" {
		fileTypeStr = string(model.FileTypeOther)
	}
	fileType := model.FileType(fileTypeStr)

	// Set max memory for multipart form (32MB)
	c.Request.ParseMultipartForm(maxMemory)

	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, "file is required")
		return
	}

	// Validate file size
	if fileHeader.Size == 0 {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, "empty file")
		return
	}

	// Validate file size using service layer validation
	if err := h.validateFileSize(fileType, fileHeader.Size); err != nil {
		h.handleServiceError(c, err)
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, "failed to open uploaded file")
		return
	}
	defer src.Close()

	// Detect content type using magic number (first 512 bytes)
	sniff := make([]byte, 512)
	n, err := io.ReadFull(src, sniff)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, "failed to read uploaded file")
		return
	}
	if n == 0 {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, "empty file")
		return
	}

	// Validate magic number to prevent content-type spoofing
	contentType := http.DetectContentType(sniff[:n])
	if !h.isValidContentType(fileType, contentType) {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, "unsupported file type or content type mismatch")
		return
	}

	// Use LimitReader to prevent memory exhaustion from oversized files
	// Limit to the declared file size + small buffer
	limitedReader := io.LimitReader(src, fileHeader.Size+1024)
	reader := io.MultiReader(bytes.NewReader(sniff[:n]), limitedReader)

	// Ensure filename has extension
	filename := fileHeader.Filename
	if ext := filepath.Ext(filename); ext == "" {
		filename = filename + guessExtension(contentType)
	}

	// Upload file
	file, err := h.fileService.UploadFile(c.Request.Context(), userID, fileType, reader, fileHeader.Size, contentType, filename, nil)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(file))
}

// GetFileHandler handles file retrieval
// GET /api/v1/files/:id
func (h *FileHandler) GetFileHandler(c *gin.Context) {
	id, err := h.parseID(c, "id")
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, "invalid file id")
		return
	}

	file, err := h.fileService.GetFile(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(file))
}

// GetUserFilesHandler handles listing user files
// GET /api/v1/files?type=avatar
func (h *FileHandler) GetUserFilesHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.CodeUnauthorized, "unauthorized")
		return
	}

	var fileType *model.FileType
	if typeStr := c.Query("type"); typeStr != "" {
		ft := model.FileType(typeStr)
		fileType = &ft
	}

	files, err := h.fileService.GetUserFiles(c.Request.Context(), userID, fileType)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(files))
}

// DeleteFileHandler handles file deletion
// DELETE /api/v1/files/:id
func (h *FileHandler) DeleteFileHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.CodeUnauthorized, "unauthorized")
		return
	}

	id, err := h.parseID(c, "id")
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, "invalid file id")
		return
	}

	if err := h.fileService.DeleteFile(c.Request.Context(), id, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "file deleted"}))
}

// GeneratePresignedUploadURLHandler handles presigned URL generation for direct client uploads
// POST /api/v1/files/presigned-upload
func (h *FileHandler) GeneratePresignedUploadURLHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.CodeUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Type        string `json:"type" binding:"required"`
		Filename    string `json:"filename" binding:"required"`
		ContentType string `json:"content_type" binding:"required"`
		Size        int64  `json:"size" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, err.Error())
		return
	}

	fileType := model.FileType(req.Type)
	resp, err := h.fileService.GeneratePresignedUploadURL(c.Request.Context(), userID, fileType, req.Filename, req.ContentType, req.Size)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(resp))
}

// ConfirmUploadHandler confirms that a file was uploaded using a presigned URL
// POST /api/v1/files/confirm-upload
func (h *FileHandler) ConfirmUploadHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.CodeUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ObjectKey   string `json:"object_key" binding:"required"`
		Filename    string `json:"filename" binding:"required"`
		ContentType string `json:"content_type" binding:"required"`
		Size        int64  `json:"size" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.CodeInvalidInput, err.Error())
		return
	}

	file, err := h.fileService.ConfirmUpload(c.Request.Context(), userID, req.ObjectKey, req.Filename, req.ContentType, req.Size)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(file))
}

func (h *FileHandler) validateFileSize(fileType model.FileType, size int64) error {
	var maxSize int64
	switch fileType {
	case model.FileTypeAvatar:
		maxSize = 5 * 1024 * 1024 // 5MB
	case model.FileTypeProjectCover, model.FileTypePostImage:
		maxSize = 10 * 1024 * 1024 // 10MB
	case model.FileTypeResume:
		maxSize = 5 * 1024 * 1024 // 5MB
	case model.FileTypeAIOutput, model.FileTypeOther:
		maxSize = 100 * 1024 * 1024 // 100MB
	default:
		maxSize = 50 * 1024 * 1024 // 50MB default
	}

	if size > maxSize {
		return errs.NewAppError(errs.ErrInvalidInput, 400, fmt.Sprintf("file size exceeds limit (%dMB)", maxSize/(1024*1024)))
	}

	return nil
}

func (h *FileHandler) isValidContentType(fileType model.FileType, contentType string) bool {
	switch fileType {
	case model.FileTypeAvatar, model.FileTypeProjectCover, model.FileTypePostImage:
		_, ok := allowedImageTypes[contentType]
		return ok
	case model.FileTypeResume:
		_, ok := allowedDocumentTypes[contentType]
		return ok
	case model.FileTypeAIOutput, model.FileTypeOther:
		// More permissive for other types
		return true
	default:
		return false
	}
}

func (h *FileHandler) parseID(c *gin.Context, param string) (int64, error) {
	idStr := c.Param(param)
	var id int64
	_, err := fmt.Sscanf(idStr, "%d", &id)
	return id, err
}

func (h *FileHandler) respondError(c *gin.Context, status int, errCode string, message string) {
	c.JSON(status, response.Error(errs.NewAppError(errs.ErrInvalidInput, status, message).WithCode(errCode)))
}

func (h *FileHandler) handleServiceError(c *gin.Context, err error) {
	statusCode := errs.ToHTTPStatus(err)
	var message string
	var code string

	if appErr, ok := err.(*errs.AppError); ok {
		message = appErr.Message
		code = appErr.Code
		if code == "" {
			code = errs.GetErrorCode(err)
		}
	} else {
		message = err.Error()
		code = errs.GetErrorCode(err)
	}

	h.logger.Error("Service error",
		zap.Error(err),
		zap.Int("status_code", statusCode),
		zap.String("error_code", code),
	)

	appErr := errs.NewAppError(err, statusCode, message).WithCode(code)
	c.JSON(statusCode, response.Error(appErr))
}

func guessExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}
