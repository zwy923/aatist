package handler

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aatist/backend/internal/platform/config"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/middleware"
	"github.com/aatist/backend/internal/user/model"
	"github.com/aatist/backend/internal/user/repository"
	"github.com/aatist/backend/internal/user/service"
	"github.com/aatist/backend/pkg/errs"
	"github.com/aatist/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	avatarFormField = "avatar"
	maxAvatarSize   = 5 * 1024 * 1024 // 5MB
	bannerFormField = "banner"
	maxBannerSize   = 10 * 1024 * 1024 // 10MB
)

var allowedAvatarTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService        service.AuthService
	profileSvc         service.ProfileService
	savedItemSvc       service.SavedItemService
	userServiceRepo    repository.UserServiceRepository
	notificationClient service.NotificationClient
	emailVerifSvc             *service.EmailVerificationService
	passwordResetSvc          *service.PasswordResetService
	mq                        interface {
		PublishEmailVerification(message interface{}) error
		PublishPasswordReset(message interface{}) error
	}
	disableEmailVerification  bool   // When true, do not send verification email after registration
	logger                    *log.Logger
	googleOAuth               config.GoogleOAuthConfig
	oauthFrontendBase         string // FRONTEND_URL — redirect after Google OAuth (hash tokens)
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authService service.AuthService,
	profileService service.ProfileService,
	savedItemService service.SavedItemService,
	userServiceRepo repository.UserServiceRepository,
	notificationClient service.NotificationClient,
	emailVerifSvc *service.EmailVerificationService,
	passwordResetSvc *service.PasswordResetService,
	mq interface {
		PublishEmailVerification(message interface{}) error
		PublishPasswordReset(message interface{}) error
	},
	disableEmailVerification bool,
	logger *log.Logger,
	googleOAuth config.GoogleOAuthConfig,
	oauthFrontendBase string,
) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		profileSvc:         profileService,
		savedItemSvc:       savedItemService,
		userServiceRepo:    userServiceRepo,
		notificationClient: notificationClient,
		emailVerifSvc:            emailVerifSvc,
		passwordResetSvc:         passwordResetSvc,
		mq:                       mq,
		disableEmailVerification: disableEmailVerification,
		logger:                   logger,
		googleOAuth:              googleOAuth,
		oauthFrontendBase:        strings.TrimRight(oauthFrontendBase, "/"),
	}
}

// RegisterHandler handles user registration
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	ip := h.getClientIP(c)
	ctx := c.Request.Context()
	role := h.normalizeRole(req.Role)

	// Prepare student/alumni fields
	var studentIDPtr, preferredNamePtr, schoolPtr, facultyPtr, majorPtr *string
	// Prepare organization fields
	var orgNamePtr, orgBioPtr, contactTitlePtr *string
	var isAffiliatedPtr *bool
	var orgSizePtr *int

	if req.Profile != nil {
		// Student/Alumni fields
		if v := strings.TrimSpace(req.Profile.StudentID); v != "" {
			value := v
			studentIDPtr = &value
		}
		if v := strings.TrimSpace(req.Profile.PreferredName); v != "" {
			value := v
			preferredNamePtr = &value
		}
		if v := strings.TrimSpace(req.Profile.School); v != "" {
			value := v
			schoolPtr = &value
		}
		if v := strings.TrimSpace(req.Profile.Faculty); v != "" {
			value := v
			facultyPtr = &value
		}
		if v := strings.TrimSpace(req.Profile.Major); v != "" {
			value := v
			majorPtr = &value
		}
		// Organization fields
		if v := strings.TrimSpace(req.Profile.OrganizationName); v != "" {
			value := v
			orgNamePtr = &value
		}
		if v := strings.TrimSpace(req.Profile.OrganizationBio); v != "" {
			value := v
			orgBioPtr = &value
		}
		if v := strings.TrimSpace(req.Profile.ContactTitle); v != "" {
			value := v
			contactTitlePtr = &value
		}
		if req.Profile.IsAffiliatedWithSchool {
			isAffiliatedPtr = &req.Profile.IsAffiliatedWithSchool
		}
		if req.Profile.OrgSize != nil {
			orgSizePtr = req.Profile.OrgSize
		}
	}

	input := service.RegisterInput{
		Email:                  req.Email,
		Password:               req.Password,
		Name:                   req.Name,
		IP:                     ip,
		Role:                   role,
		StudentID:              studentIDPtr,
		PreferredName:          preferredNamePtr,
		School:                 schoolPtr,
		Faculty:                facultyPtr,
		Major:                  majorPtr,
		OrganizationName:       orgNamePtr,
		OrganizationBio:        orgBioPtr,
		ContactTitle:           contactTitlePtr,
		IsAffiliatedWithSchool: isAffiliatedPtr,
		OrgSize:                orgSizePtr,
	}

	user, tokens, err := h.authService.Register(ctx, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Send verification email unless disabled by config (e.g. when SendGrid quota is exceeded)
	if !h.disableEmailVerification && h.emailVerifSvc != nil && h.mq != nil {
		token, err := h.emailVerifSvc.GenerateVerificationToken(ctx, user.ID, user.Email)
		if err != nil {
			h.logger.Error("Failed to generate verification token", zap.Error(err))
			// Non-critical, continue with registration
		} else {
			// Publish to MQ for async email sending
			emailMsg := model.EmailVerificationRequest{
				UserID: user.ID,
				Email:  user.Email,
				Name:   user.Name,
				Token:  token,
			}
			if err := h.mq.PublishEmailVerification(emailMsg); err != nil {
				h.logger.Error("Failed to publish email verification message", zap.Error(err))
				// Non-critical, continue with registration
			}
		}
	}

	if user.RoleVerified {
		h.logger.Info("User registered with verified school email (role_verified=true)",
			zap.Int64("user_id", user.ID),
			zap.String("email", user.Email),
		)
	}

	h.respondSuccess(c, http.StatusCreated, user, tokens)
}

// LoginHandler handles user login
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	ip := h.getClientIP(c)
	user, tokens, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, ip, req.LoginType)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, user, tokens)
}

// RefreshTokenHandler handles token refresh
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	tokens, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	}))
}

// LogoutHandler handles user logout
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "logged out successfully"}))
}

// VerifyEmailHandler handles email verification via POST
func (h *AuthHandler) VerifyEmailHandler(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	h.processEmailVerification(c, req.Token)
}

// VerifyEmailGetHandler handles email verification via GET (token in query)
func (h *AuthHandler) VerifyEmailGetHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "token is required")
		return
	}

	h.processEmailVerification(c, token)
}

func (h *AuthHandler) processEmailVerification(c *gin.Context, token string) {
	if err := h.authService.VerifyEmail(c.Request.Context(), token); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "email verified successfully"}))
}

// ForgotPasswordHandler handles forgot password request
func (h *AuthHandler) ForgotPasswordHandler(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	ctx := c.Request.Context()

	// Generate reset token
	token, userID, userName, err := h.passwordResetSvc.GenerateResetToken(ctx, req.Email)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// If token is empty, user not found - but we still return success for security
	// (don't reveal if email exists)
	if token == "" {
		c.JSON(http.StatusOK, response.Success(gin.H{
			"message": "if the email exists, a password reset link will be sent",
		}))
		return
	}

	// Publish to MQ for async email sending
	if h.mq != nil {
		emailMsg := model.PasswordResetRequest{
			UserID: userID,
			Email:  req.Email,
			Name:   userName,
			Token:  token,
		}
		if err := h.mq.PublishPasswordReset(emailMsg); err != nil {
			h.logger.Error("Failed to publish password reset message", zap.Error(err))
			// Non-critical for the user, continue
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message": "if the email exists, a password reset link will be sent",
	}))
}

// ResetPasswordHandler handles password reset with token
func (h *AuthHandler) ResetPasswordHandler(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	if err := h.passwordResetSvc.VerifyTokenAndReset(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "password reset successfully"}))
}

// GetCurrentUserHandler returns profile of the authenticated user.
func (h *AuthHandler) GetCurrentUserHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	user, err := h.profileSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(mapUserToResponse(user)))
}

// UpdateCurrentUserHandler updates profile of the authenticated user.
func (h *AuthHandler) UpdateCurrentUserHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	if err := h.profileSvc.EnsureProfileUpdateRateLimit(c.Request.Context(), userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	profile, err := h.profileSvc.UpdateProfile(c.Request.Context(), userID, req.ToServiceInput())
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(mapUserToResponse(profile)))
}

// UploadAvatarHandler handles avatar uploads via multipart/form-data.
func (h *AuthHandler) UploadAvatarHandler(c *gin.Context) {
	if h.logger == nil {
		// Temporary fallback to avoid secondary panic from logging
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logger is nil"})
		return
	}

	if h.profileSvc == nil {
		h.logger.Error("profileSvc is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "profile service not configured"})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	h.logger.Info("UploadAvatar: before rate limit", zap.Int64("user_id", userID))
	if err := h.profileSvc.EnsureAvatarUploadRateLimit(c.Request.Context(), userID); err != nil {
		h.handleServiceError(c, err)
		return
	}
	h.logger.Info("UploadAvatar: after rate limit")

	fileHeader, err := c.FormFile(avatarFormField)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "avatar file is required")
		return
	}

	if fileHeader.Size == 0 || fileHeader.Size > maxAvatarSize {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "avatar size exceeds limit (5MB)")
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "failed to open uploaded file")
		return
	}
	defer src.Close()

	sniff := make([]byte, 512)
	n, err := io.ReadFull(src, sniff)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "failed to read uploaded file")
		return
	}
	if n == 0 {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "empty file")
		return
	}

	contentType := http.DetectContentType(sniff[:n])
	if _, ok := allowedAvatarTypes[contentType]; !ok {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "unsupported image type")
		return
	}

	reader := io.MultiReader(bytes.NewReader(sniff[:n]), src)
	filename := fileHeader.Filename
	if ext := filepath.Ext(filename); ext == "" {
		filename = filename + guessExtensionForType(contentType)
	}

	h.logger.Info("UploadAvatar: before upload", zap.Int64("user_id", userID))
	profile, err := h.profileSvc.UploadAvatar(c.Request.Context(), userID, reader, fileHeader.Size, contentType, filename)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	h.logger.Info("UploadAvatar: after upload")

	var updatedAt string
	var lastUpdated *string
	if !profile.UpdatedAt.IsZero() {
		updatedAt = profile.UpdatedAt.Format(time.RFC3339)
	}
	if updatedAt != "" {
		lastUpdated = &updatedAt
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"avatar_url":      profile.AvatarURL,
		"file_size":       fileHeader.Size,
		"last_updated_at": lastUpdated,
		"user":            mapUserToResponse(profile),
	}))
}

// UploadProfileBannerHandler handles profile cover / banner uploads (multipart field "banner").
func (h *AuthHandler) UploadProfileBannerHandler(c *gin.Context) {
	if h.logger == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logger is nil"})
		return
	}
	if h.profileSvc == nil {
		h.logger.Error("profileSvc is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "profile service not configured"})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	if err := h.profileSvc.EnsureBannerUploadRateLimit(c.Request.Context(), userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	fileHeader, err := c.FormFile(bannerFormField)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "banner file is required")
		return
	}

	if fileHeader.Size == 0 || fileHeader.Size > maxBannerSize {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "banner size exceeds limit (10MB)")
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "failed to open uploaded file")
		return
	}
	defer src.Close()

	sniff := make([]byte, 512)
	n, err := io.ReadFull(src, sniff)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "failed to read uploaded file")
		return
	}
	if n == 0 {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "empty file")
		return
	}

	contentType := http.DetectContentType(sniff[:n])
	if _, ok := allowedAvatarTypes[contentType]; !ok {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "unsupported image type")
		return
	}

	reader := io.MultiReader(bytes.NewReader(sniff[:n]), src)
	filename := fileHeader.Filename
	if ext := filepath.Ext(filename); ext == "" {
		filename = filename + guessExtensionForType(contentType)
	}

	profile, err := h.profileSvc.UploadProfileBanner(c.Request.Context(), userID, reader, fileHeader.Size, contentType, filename)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	var updatedAt string
	var lastUpdated *string
	if !profile.UpdatedAt.IsZero() {
		updatedAt = profile.UpdatedAt.Format(time.RFC3339)
	}
	if updatedAt != "" {
		lastUpdated = &updatedAt
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"banner_url":      profile.BannerURL,
		"file_size":       fileHeader.Size,
		"last_updated_at": lastUpdated,
		"user":            mapUserToResponse(profile),
	}))
}

// AddUserSkillHandler handles POST /users/me/skills
func (h *AuthHandler) AddUserSkillHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req SkillInput
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	skill := model.Skill{
		Name:  strings.TrimSpace(req.Name),
		Level: strings.ToLower(strings.TrimSpace(req.Level)),
	}

	user, err := h.profileSvc.AddUserSkill(c.Request.Context(), userID, skill)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(mapUserToResponse(user)))
}

// RemoveUserSkillHandler handles DELETE /users/me/skills/:name
func (h *AuthHandler) RemoveUserSkillHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	skillName := c.Param("name")
	if skillName == "" {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "skill name is required")
		return
	}

	user, err := h.profileSvc.RemoveUserSkill(c.Request.Context(), userID, skillName)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(mapUserToResponse(user)))
}

// AddUserCourseHandler handles POST /users/me/courses
func (h *AuthHandler) AddUserCourseHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req model.Course
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	user, err := h.profileSvc.AddUserCourse(c.Request.Context(), userID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(mapUserToResponse(user)))
}

// RemoveUserCourseHandler handles DELETE /users/me/courses/:code
func (h *AuthHandler) RemoveUserCourseHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	courseCode := c.Param("code")
	if courseCode == "" {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "course code is required")
		return
	}

	user, err := h.profileSvc.RemoveUserCourse(c.Request.Context(), userID, courseCode)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(mapUserToResponse(user)))
}

// GetUserServicesHandler handles GET /users/me/services
func (h *AuthHandler) GetUserServicesHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}
	services, err := h.userServiceRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"services": services}))
}

// CreateUserServiceHandler handles POST /users/me/services
func (h *AuthHandler) CreateUserServiceHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Category          string   `json:"category" binding:"required,max=100"`
		ExperienceSummary string   `json:"experience_summary" binding:"omitempty,max=500"`
		Title             string   `json:"title" binding:"omitempty,max=200"`
		Description       string   `json:"description" binding:"omitempty,max=5000"`
		ShortDescription  string   `json:"short_description" binding:"omitempty,max=500"`
		PriceType         string   `json:"price_type" binding:"omitempty,max=128"`
		PriceMin          *int     `json:"price_min" binding:"omitempty,gte=0"`
		PriceMax          *int     `json:"price_max" binding:"omitempty,gte=0"`
		MediaURLs         []string `json:"media_urls"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}
	normalizedPriceType, err := model.NormalizePriceType(req.PriceType)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}
	expSummary := strings.TrimSpace(req.ExperienceSummary)
	if expSummary == "" {
		expSummary = strings.TrimSpace(req.Description)
	}
	if expSummary == "" {
		expSummary = strings.TrimSpace(req.ShortDescription)
	}
	s := &model.UserService{
		UserID:            userID,
		Category:          strings.TrimSpace(req.Category),
		ExperienceSummary: expSummary,
		Title:             strings.TrimSpace(req.Title),
		Description:       strings.TrimSpace(req.Description),
		ShortDescription:  strings.TrimSpace(req.ShortDescription),
		PriceType:         normalizedPriceType,
		PriceMin:          req.PriceMin,
		PriceMax:          req.PriceMax,
		MediaURLs:         model.StringArray(req.MediaURLs),
	}
	if err := h.userServiceRepo.Create(c.Request.Context(), s); err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(s))
}

// UpdateUserServiceHandler handles PATCH /users/me/services/:id
func (h *AuthHandler) UpdateUserServiceHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid service id")
		return
	}
	var req struct {
		Category          string   `json:"category" binding:"omitempty,max=100"`
		ExperienceSummary string   `json:"experience_summary" binding:"omitempty,max=500"`
		Title             string   `json:"title" binding:"omitempty,max=200"`
		Description       string   `json:"description" binding:"omitempty,max=5000"`
		ShortDescription  string   `json:"short_description" binding:"omitempty,max=500"`
		PriceType         string   `json:"price_type" binding:"omitempty,max=128"`
		PriceMin          *int     `json:"price_min" binding:"omitempty,gte=0"`
		PriceMax          *int     `json:"price_max" binding:"omitempty,gte=0"`
		MediaURLs         []string `json:"media_urls"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}
	s, err := h.userServiceRepo.FindByID(c.Request.Context(), id, userID)
	if err != nil {
		h.respondError(c, http.StatusNotFound, errs.ErrInvalidInput, "service not found")
		return
	}
	if req.Category != "" {
		s.Category = strings.TrimSpace(req.Category)
	}
	if req.ExperienceSummary != "" {
		s.ExperienceSummary = strings.TrimSpace(req.ExperienceSummary)
	}
	if req.Title != "" {
		s.Title = strings.TrimSpace(req.Title)
	}
	if req.Description != "" {
		s.Description = strings.TrimSpace(req.Description)
	}
	if req.ShortDescription != "" {
		s.ShortDescription = strings.TrimSpace(req.ShortDescription)
	}
	if req.PriceType != "" {
		normalizedPriceType, nerr := model.NormalizePriceType(req.PriceType)
		if nerr != nil {
			h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, nerr.Error())
			return
		}
		s.PriceType = normalizedPriceType
	}
	if req.PriceMin != nil {
		s.PriceMin = req.PriceMin
	}
	if req.PriceMax != nil {
		s.PriceMax = req.PriceMax
	}
	if req.MediaURLs != nil {
		s.MediaURLs = model.StringArray(req.MediaURLs)
	}
	if s.PriceMin != nil && *s.PriceMin < 0 {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "price_min must not be negative")
		return
	}
	if s.PriceMax != nil && *s.PriceMax < 0 {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "price_max must not be negative")
		return
	}
	if err := h.userServiceRepo.Update(c.Request.Context(), s); err != nil {
		h.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(s))
}

// DeleteUserServiceHandler handles DELETE /users/me/services/:id
func (h *AuthHandler) DeleteUserServiceHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid service id")
		return
	}
	if err := h.userServiceRepo.Delete(c.Request.Context(), id, userID); err != nil {
		h.respondError(c, http.StatusNotFound, errs.ErrInvalidInput, "service not found")
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "deleted"}))
}

// SearchSkillsHandler handles GET /skills
func (h *AuthHandler) SearchSkillsHandler(c *gin.Context) {
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	skills, err := h.profileSvc.SearchSkills(c.Request.Context(), query, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(skills))
}

// SearchCoursesHandler handles GET /courses
func (h *AuthHandler) SearchCoursesHandler(c *gin.Context) {
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	courses, err := h.profileSvc.SearchCourses(c.Request.Context(), query, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(courses))
}

// SearchTagsHandler handles GET /tags
func (h *AuthHandler) SearchTagsHandler(c *gin.Context) {
	tagType := c.Query("type")
	if tagType == "" {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "tag type is required")
		return
	}
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tags, err := h.profileSvc.SearchTags(c.Request.Context(), tagType, query, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(tags))
}

// Helper methods

func (h *AuthHandler) respondSuccess(c *gin.Context, statusCode int, user *model.User, tokens *service.Tokens) {
	data := AuthResponse{
		User:         mapUserToResponse(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	c.JSON(statusCode, response.Success(data))
}

func (h *AuthHandler) respondError(c *gin.Context, statusCode int, err error, message string) {
	appErr := errs.NewAppError(err, statusCode, message).WithCode(errs.GetErrorCode(err))
	c.JSON(statusCode, response.Error(appErr))
}

func (h *AuthHandler) handleServiceError(c *gin.Context, err error) {
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

func (h *AuthHandler) normalizeRole(role string) model.Role {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "alumni":
		return model.RoleAlumni
	case "org_person", "org-person", "organization_person":
		return model.RoleOrgPerson
	case "org_team", "org-team", "organization_team", "organization":
		return model.RoleOrgTeam
	default:
		return model.RoleStudent
	}
}

func (h *AuthHandler) getClientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = c.GetHeader("X-Real-IP")
	}
	if ip == "" {
		ip = c.ClientIP()
	}
	return ip
}

func mapUserToResponse(user *model.User) UserResponse {
	resp := UserResponse{
		ID:                   user.ID,
		Email:                user.Email,
		Name:                 user.Name,
		PreferredName:        user.PreferredName,
		AvatarURL:            user.AvatarURL,
		BannerURL:            user.BannerURL,
		Role:                 user.Role.String(),
		Bio:                  user.Bio,
		Website:              user.Website,
		LinkedIn:             user.LinkedIn,
		Behance:              user.Behance,
		Languages:            user.Languages,
		ProfessionalInterests: user.ProfessionalInterests,
		ProfileVisibility:    user.ProfileVisibility.String(),
		IsVerifiedEmail:   user.IsVerifiedEmail,
		RoleVerified:      user.RoleVerified,
		OAuthProvider:     user.OAuthProvider,
		CreatedAt:         user.CreatedAt.Format(time.RFC3339),
		// Student/Alumni fields
		StudentID:           user.StudentID,
		School:              user.School,
		Faculty:             user.Faculty,
		Major:               user.Major,
		Skills:              user.Skills,
		Courses:             user.Courses,
		PortfolioVisibility: user.PortfolioVisibility.String(),
		// Organization fields
		OrganizationName:       user.OrganizationName,
		OrganizationBio:        user.OrganizationBio,
		ContactTitle:           user.ContactTitle,
		IsAffiliatedWithSchool: user.IsAffiliatedWithSchool,
		OrgSize:                user.OrgSize,
	}

	if user.LastLoginAt != nil {
		lastLogin := user.LastLoginAt.Format(time.RFC3339)
		resp.LastLoginAt = &lastLogin
	}

	return resp
}

func guessExtensionForType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}

// GetUserByIDHandler returns public user information by ID
func (h *AuthHandler) GetUserByIDHandler(c *gin.Context) {
	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid user id")
		return
	}

	user, err := h.profileSvc.GetProfile(c.Request.Context(), req.ID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Check visibility
	var viewerEmail *string
	// Try to get email from context (set by middleware)
	if email, exists := c.Get("email"); exists {
		if e, ok := email.(string); ok {
			viewerEmail = &e
		}
	}

	// Check if viewer can access this profile
	if !user.ProfileVisibility.CanView(viewerEmail) {
		h.respondError(c, http.StatusForbidden, errs.ErrUnauthorized, "profile is not accessible")
		return
	}

	// Return public profile (exclude sensitive fields)
	publicProfile := mapUserToPublicResponse(user)
	// Include service offerings for talent profiles
	if user.Role.IsStudentRole() {
		services, _ := h.userServiceRepo.FindByUserID(c.Request.Context(), req.ID)
		if services != nil {
			publicProfile["services"] = services
		}
	}
	c.JSON(http.StatusOK, response.Success(publicProfile))
}

// GetSavedItemsHandler returns all saved items for the authenticated user
func (h *AuthHandler) GetSavedItemsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	itemType := c.Query("type")
	if itemType != "" {
		savedItemType := model.SavedItemType(itemType)
		if savedItemType != model.SavedItemTypeProject &&
			savedItemType != model.SavedItemTypeOpportunity &&
			savedItemType != model.SavedItemTypeUser &&
			savedItemType != model.SavedItemTypeEvent {
			h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid item type")
			return
		}
		items, err := h.savedItemSvc.GetSavedItemsByType(c.Request.Context(), userID, savedItemType)
		if err != nil {
			h.handleServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, response.Success(gin.H{
			"items": items,
		}))
		return
	}

	items, err := h.savedItemSvc.GetSavedItems(c.Request.Context(), userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"items": items,
	}))
}

// SaveItemHandler saves an item
func (h *AuthHandler) SaveItemHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ItemID   int64  `json:"item_id" binding:"required"`
		ItemType string `json:"item_type" binding:"required,oneof=project opportunity user event"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	itemType := model.SavedItemType(req.ItemType)
	if err := h.savedItemSvc.SaveItem(c.Request.Context(), userID, req.ItemID, itemType); err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Send notification if saving a user profile
	if itemType == model.SavedItemTypeUser && h.notificationClient != nil {
		// Get saver info
		saver, err := h.profileSvc.GetProfile(c.Request.Context(), userID)
		if err == nil {
			// Notify the saved user
			service.NotifyProfileSaved(h.notificationClient, c.Request.Context(), req.ItemID, userID, saver.Name)
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "item saved successfully"}))
}

// UnsaveItemHandler unsaves an item
func (h *AuthHandler) UnsaveItemHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	// 1. Check for ID in path (e.g., DELETE /users/me/saved/:id)
	if idStr := c.Param("id"); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			if err := h.savedItemSvc.UnsaveItemByID(c.Request.Context(), userID, id); err != nil {
				h.handleServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, response.Success(gin.H{"message": "item unsaved successfully"}))
			return
		}
	}

	// 2. Check for type and targetId in query params (e.g., DELETE /users/me/saved?type=...&targetId=...)
	itemTypeStr := c.Query("type")
	targetIdStr := c.Query("targetId")
	if itemTypeStr != "" && targetIdStr != "" {
		targetId, err := strconv.ParseInt(targetIdStr, 10, 64)
		if err == nil {
			itemType := model.SavedItemType(itemTypeStr)
			if err := h.savedItemSvc.UnsaveItem(c.Request.Context(), userID, targetId, itemType); err != nil {
				h.handleServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, response.Success(gin.H{"message": "item unsaved successfully"}))
			return
		}
	}

	// 3. Fallback to JSON body (for backward compatibility)
	var req struct {
		ItemID   int64  `json:"item_id"`
		ItemType string `json:"item_type"`
	}
	if err := c.ShouldBindJSON(&req); err == nil && req.ItemID != 0 && req.ItemType != "" {
		itemType := model.SavedItemType(req.ItemType)
		if err := h.savedItemSvc.UnsaveItem(c.Request.Context(), userID, req.ItemID, itemType); err != nil {
			h.handleServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, response.Success(gin.H{"message": "item unsaved successfully"}))
		return
	}

	h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "missing saved item id or type/targetId")
}

// mapUserToPublicResponse maps user to public response (excludes sensitive fields)
func mapUserToPublicResponse(user *model.User) gin.H {
	resp := gin.H{
		"id":                   user.ID,
		"name":                 user.Name,
		"avatar_url":           user.AvatarURL,
		"banner_url":           user.BannerURL,
		"role":                 user.Role.String(),
		"bio":                  user.Bio,
		"role_verified":        user.RoleVerified,
		"website":              user.Website,
		"linkedin":             user.LinkedIn,
		"behance":              user.Behance,
		"languages":            user.Languages,
		"professional_interests": user.ProfessionalInterests,
		"profile_visibility":   user.ProfileVisibility.String(),
		"created_at":           user.CreatedAt.Format(time.RFC3339),
	}

	// Add student/alumni fields if applicable
	if user.Role.IsStudentRole() {
		if user.PreferredName != nil && strings.TrimSpace(*user.PreferredName) != "" {
			resp["preferred_name"] = strings.TrimSpace(*user.PreferredName)
		}
		resp["school"] = user.School
		resp["faculty"] = user.Faculty
		resp["major"] = user.Major
		resp["skills"] = user.Skills
		resp["courses"] = user.Courses
		resp["portfolio_visibility"] = user.PortfolioVisibility.String()
	}

	// Add organization fields if applicable
	if user.Role.IsOrgRole() {
		resp["organization_name"] = user.OrganizationName
		resp["organization_bio"] = user.OrganizationBio
		resp["contact_title"] = user.ContactTitle
		resp["is_affiliated_with_school"] = user.IsAffiliatedWithSchool
		resp["org_size"] = user.OrgSize
	}

	return resp
}

// CheckEmailHandler checks if an email is already registered
func (h *AuthHandler) CheckEmailHandler(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "email query parameter is required")
		return
	}

	exists, err := h.authService.CheckEmailExists(c.Request.Context(), email)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(CheckExistsResponse{Exists: exists}))
}

// ChangePasswordHandler handles password change for authenticated user
func (h *AuthHandler) ChangePasswordHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "password changed successfully"}))
}

// GetUserSummaryHandler returns a lightweight summary of a user's public profile
func (h *AuthHandler) GetUserSummaryHandler(c *gin.Context) {
	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid user id")
		return
	}

	// First check profile visibility
	user, err := h.profileSvc.GetProfile(c.Request.Context(), req.ID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Check visibility
	var viewerEmail *string
	if email, exists := c.Get("email"); exists {
		if e, ok := email.(string); ok {
			viewerEmail = &e
		}
	}

	// Check if viewer can access this profile
	if !user.ProfileVisibility.CanView(viewerEmail) {
		h.respondError(c, http.StatusForbidden, errs.ErrUnauthorized, "profile is not accessible")
		return
	}

	// Return summary (lightweight version)
	summary, err := h.profileSvc.GetUserSummary(c.Request.Context(), req.ID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(summary))
}

// SearchUsersHandler handles GET /users/search
func (h *AuthHandler) SearchUsersHandler(c *gin.Context) {
	var query struct {
		Query   string `form:"q"`
		Faculty string `form:"faculty"`
		School  string `form:"school"`
		Major   string `form:"major"`
		Role    string `form:"role"`
		Limit   int    `form:"limit"`
		Offset  int    `form:"offset"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, "invalid query parameters")
		return
	}

	filter := repository.UserSearchFilter{
		Query:         query.Query,
		Faculty:       query.Faculty,
		School:        query.School,
		Major:         query.Major,
		Role:          query.Role,
		Limit:         query.Limit,
		Offset:        query.Offset,
		ExcludeUserID: 0,
	}
	// Exclude current user from talent search (strict: cannot find own profile)
	if excludeID, ok := middleware.GetUserIDOptional(c); ok && excludeID > 0 {
		filter.ExcludeUserID = excludeID
	}

	users, err := h.profileSvc.SearchUsers(c.Request.Context(), filter)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Map to public response
	results := make([]gin.H, 0, len(users))
	for _, u := range users {
		results = append(results, mapUserToPublicResponse(u))
	}

	c.JSON(http.StatusOK, response.Success(results))
}
