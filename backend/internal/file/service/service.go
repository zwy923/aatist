package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aatist/backend/internal/file/model"
	"github.com/aatist/backend/internal/file/repository"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/storage"
	"github.com/aatist/backend/pkg/errs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PresignedUploadResponse represents the response for presigned URL generation
type PresignedUploadResponse struct {
	UploadURL  string    `json:"upload_url"`
	ObjectKey  string    `json:"object_key"`
	FileID     int64     `json:"file_id,omitempty"` // Set after confirmation
	ExpiresIn  int       `json:"expires_in"`         // Seconds
	PublicURL  string    `json:"public_url,omitempty"`
}

// FileService defines the interface for file operations
type FileService interface {
	UploadFile(ctx context.Context, userID int64, fileType model.FileType, reader io.Reader, size int64, contentType, filename string, metadata map[string]interface{}) (*model.File, error)
	GetFile(ctx context.Context, id int64) (*model.File, error)
	GetUserFiles(ctx context.Context, userID int64, fileType *model.FileType) ([]*model.File, error)
	DeleteFile(ctx context.Context, id int64, userID int64) error
	GeneratePresignedUploadURL(ctx context.Context, userID int64, fileType model.FileType, filename, contentType string, size int64) (*PresignedUploadResponse, error)
	ConfirmUpload(ctx context.Context, userID int64, objectKey, filename, contentType string, size int64) (*model.File, error)
}

type fileService struct {
	fileRepo repository.FileRepository
	storage  storage.ObjectStorage
	logger   *log.Logger
}

// NewFileService creates a new file service
func NewFileService(fileRepo repository.FileRepository, storage storage.ObjectStorage, logger *log.Logger) FileService {
	return &fileService{
		fileRepo: fileRepo,
		storage:  storage,
		logger:   logger,
	}
}

func (s *fileService) UploadFile(ctx context.Context, userID int64, fileType model.FileType, reader io.Reader, size int64, contentType, filename string, metadata map[string]interface{}) (*model.File, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("object storage not configured")
	}
	if size <= 0 {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "invalid file size")
	}

	// Validate file type
	if !isValidFileType(fileType) {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "invalid file type")
	}

	// Generate object key based on file type
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(filename)))
	if ext == "" {
		ext = guessExtension(contentType)
	}
	if ext == "" {
		ext = ".bin"
	}

	objectKey := generateObjectKey(fileType, userID, ext)

	// Upload to S3
	metadataMap := make(map[string]string)
	metadataMap["uploaded_at"] = time.Now().UTC().Format(time.RFC3339)
	metadataMap["user_id"] = fmt.Sprintf("%d", userID)
	metadataMap["file_type"] = string(fileType)

	if err := s.storage.Upload(ctx, objectKey, reader, size, contentType, metadataMap); err != nil {
		if s.logger != nil {
			s.logger.Error("file upload failed",
				zap.Int64("user_id", userID),
				zap.String("file_type", string(fileType)),
				zap.String("object_key", objectKey),
				zap.String("content_type", contentType),
				zap.Int64("size", size),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Build public URL
	url := s.storage.BuildPublicURL(objectKey)

	// Serialize metadata
	metadataJSON := ""
	if metadata != nil {
		bytes, err := json.Marshal(metadata)
		if err == nil {
			metadataJSON = string(bytes)
		}
	}

	// Create file record
	file := &model.File{
		UserID:      userID,
		Type:        fileType,
		ObjectKey:   objectKey,
		URL:         url,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		Metadata:    metadataJSON,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		// If DB insert fails, try to delete the uploaded object to maintain consistency
		if deleteErr := s.storage.Delete(ctx, objectKey); deleteErr != nil {
			if s.logger != nil {
				s.logger.Warn("failed to delete S3 object after DB insert failure",
					zap.String("object_key", objectKey),
					zap.Error(deleteErr),
				)
			}
		}
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("file uploaded",
			zap.Int64("file_id", file.ID),
			zap.Int64("user_id", userID),
			zap.String("file_type", string(fileType)),
			zap.String("object_key", objectKey),
			zap.String("content_type", contentType),
			zap.Int64("size", size),
		)
	}

	return file, nil
}

func (s *fileService) GetFile(ctx context.Context, id int64) (*model.File, error) {
	return s.fileRepo.FindByID(ctx, id)
}

func (s *fileService) GetUserFiles(ctx context.Context, userID int64, fileType *model.FileType) ([]*model.File, error) {
	if fileType != nil {
		return s.fileRepo.FindByUserIDAndType(ctx, userID, *fileType)
	}
	return s.fileRepo.FindByUserID(ctx, userID)
}

func (s *fileService) DeleteFile(ctx context.Context, id int64, userID int64) error {
	// Verify file exists and belongs to user
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if file.UserID != userID {
		return errs.NewAppError(errs.ErrForbidden, 403, "file does not belong to user").WithCode(errs.CodeForbidden)
	}

	// Clear references (avatar_url, cover_image_url) before deletion
	if err := s.fileRepo.ClearReferencesByURL(ctx, file.URL, file.Type, file.UserID); err != nil {
		if s.logger != nil {
			s.logger.Warn("failed to clear file references",
				zap.Int64("file_id", id),
				zap.String("url", file.URL),
				zap.Error(err),
			)
		}
		// Continue with deletion; references may be stale
	}

	// Delete from S3 first (before DB deletion to maintain consistency)
	// If S3 deletion fails, we still want to log it but continue with DB deletion
	// to maintain idempotency (if called again, DB delete will be no-op)
	if s.storage != nil && file.ObjectKey != "" {
		if err := s.storage.Delete(ctx, file.ObjectKey); err != nil {
			// Log error but continue with DB deletion for idempotency
			if s.logger != nil {
				s.logger.Warn("failed to delete S3 object, continuing with DB deletion",
					zap.Int64("file_id", id),
					zap.String("object_key", file.ObjectKey),
					zap.Error(err),
				)
			}
		}
	}

	// Delete DB record (idempotent - if already deleted, will return not found)
	if err := s.fileRepo.Delete(ctx, id); err != nil {
		return err
	}

	if s.logger != nil {
		s.logger.Info("file deleted",
			zap.Int64("file_id", id),
			zap.Int64("user_id", userID),
			zap.String("object_key", file.ObjectKey),
		)
	}

	return nil
}

// GeneratePresignedUploadURL generates a pre-signed URL for direct client uploads
func (s *fileService) GeneratePresignedUploadURL(ctx context.Context, userID int64, fileType model.FileType, filename, contentType string, size int64) (*PresignedUploadResponse, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("object storage not configured")
	}

	// Validate file type
	if !isValidFileType(fileType) {
		return nil, errs.NewAppError(errs.ErrInvalidInput, 400, "invalid file type")
	}

	// Validate file size based on type
	if err := validateFileSize(fileType, size); err != nil {
		return nil, err
	}

	// Validate content type
	if err := validateContentType(fileType, contentType); err != nil {
		return nil, err
	}

	// Generate object key
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(filename)))
	if ext == "" {
		ext = guessExtension(contentType)
	}
	if ext == "" {
		ext = ".bin"
	}
	objectKey := generateObjectKey(fileType, userID, ext)

	// Generate presigned URL (expires in 1 hour)
	expires := 1 * time.Hour
	uploadURL, err := s.storage.PresignedPutURL(ctx, objectKey, expires)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Create file record in "pending" state (we'll update it after upload confirmation)
	file := &model.File{
		UserID:      userID,
		Type:        fileType,
		ObjectKey:   objectKey,
		URL:         "", // Will be set after confirmation
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		Metadata:    "", // Can be updated later
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Build public URL
	publicURL := s.storage.BuildPublicURL(objectKey)

	return &PresignedUploadResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		FileID:    file.ID,
		ExpiresIn: int(expires.Seconds()),
		PublicURL: publicURL,
	}, nil
}

// ConfirmUpload confirms that a file was uploaded using a presigned URL
func (s *fileService) ConfirmUpload(ctx context.Context, userID int64, objectKey, filename, contentType string, size int64) (*model.File, error) {
	// Find file by objectKey
	file, err := s.fileRepo.FindByObjectKey(ctx, objectKey)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if file.UserID != userID {
		return nil, errs.NewAppError(errs.ErrForbidden, 403, "file does not belong to user").WithCode(errs.CodeForbidden)
	}

	// Update file with final URL and metadata
	file.URL = s.storage.BuildPublicURL(objectKey)
	file.Filename = filename
	file.ContentType = contentType
	file.Size = size

	// Update in database
	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("file upload confirmed",
			zap.Int64("file_id", file.ID),
			zap.Int64("user_id", userID),
			zap.String("object_key", objectKey),
		)
	}

	return file, nil
}

func validateFileSize(fileType model.FileType, size int64) error {
	var maxSize int64
	switch fileType {
	case model.FileTypeAvatar:
		maxSize = 5 * 1024 * 1024 // 5MB
	case model.FileTypeProjectCover, model.FileTypePostImage, model.FileTypeProfileBanner:
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

func validateContentType(fileType model.FileType, contentType string) error {
	allowedTypes := getAllowedContentTypes(fileType)
	if len(allowedTypes) == 0 {
		return nil // No restriction
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return nil
		}
	}

	return errs.NewAppError(errs.ErrInvalidInput, 400, fmt.Sprintf("content type %s not allowed for file type %s", contentType, fileType))
}

func getAllowedContentTypes(fileType model.FileType) []string {
	switch fileType {
	case model.FileTypeAvatar, model.FileTypeProjectCover, model.FileTypePostImage, model.FileTypeProfileBanner:
		return []string{"image/jpeg", "image/png", "image/webp", "image/gif"}
	case model.FileTypeResume:
		return []string{"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}
	case model.FileTypeAIOutput, model.FileTypeOther:
		return []string{} // More permissive
	default:
		return []string{}
	}
}

func isValidFileType(fileType model.FileType) bool {
	switch fileType {
	case model.FileTypeAvatar, model.FileTypeProjectCover, model.FileTypeProfileBanner, model.FileTypePostImage,
		model.FileTypeResume, model.FileTypeAIOutput, model.FileTypeOther:
		return true
	default:
		return false
	}
}

func generateObjectKey(fileType model.FileType, userID int64, ext string) string {
	uuidStr := uuid.New().String()
	switch fileType {
	case model.FileTypeAvatar:
		return fmt.Sprintf("avatars/%d/%s%s", userID, uuidStr, ext)
	case model.FileTypeProfileBanner:
		return fmt.Sprintf("banners/%d/%s%s", userID, uuidStr, ext)
	case model.FileTypeProjectCover:
		return fmt.Sprintf("projects/%d/%s%s", userID, uuidStr, ext)
	case model.FileTypePostImage:
		return fmt.Sprintf("posts/%d/%s%s", userID, uuidStr, ext)
	case model.FileTypeResume:
		return fmt.Sprintf("resumes/%d/%s%s", userID, uuidStr, ext)
	case model.FileTypeAIOutput:
		return fmt.Sprintf("ai/%d/%s%s", userID, uuidStr, ext)
	default:
		return fmt.Sprintf("files/%d/%s%s", userID, uuidStr, ext)
	}
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
	case "text/plain":
		return ".txt"
	default:
		return ""
	}
}
