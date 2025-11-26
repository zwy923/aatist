package handler

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/platform/middleware"
	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/internal/user/service"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"github.com/aalto-talent-network/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	avatarFormField = "avatar"
	maxAvatarSize   = 5 * 1024 * 1024 // 5MB
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
	notificationClient service.NotificationClient
	emailVerifSvc      *service.EmailVerificationService
	passwordResetSvc   *service.PasswordResetService
	mq                 interface {
		PublishEmailVerification(message interface{}) error
		PublishPasswordReset(message interface{}) error
	}
	logger *log.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authService service.AuthService,
	profileService service.ProfileService,
	savedItemService service.SavedItemService,
	notificationClient service.NotificationClient,
	emailVerifSvc *service.EmailVerificationService,
	passwordResetSvc *service.PasswordResetService,
	mq interface {
		PublishEmailVerification(message interface{}) error
		PublishPasswordReset(message interface{}) error
	},
	logger *log.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		profileSvc:         profileService,
		savedItemSvc:       savedItemService,
		notificationClient: notificationClient,
		emailVerifSvc:      emailVerifSvc,
		passwordResetSvc:   passwordResetSvc,
		mq:                 mq,
		logger:             logger,
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
	var studentIDPtr, schoolPtr, facultyPtr, majorPtr *string
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

	// Send verification email to all users (all users need email verification)
	// School email domains (e.g., @aalto.fi) get role_verified = true, but still need email verification
	if h.emailVerifSvc != nil && h.mq != nil {
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
	user, tokens, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, ip)
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
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	if err := h.profileSvc.EnsureAvatarUploadRateLimit(c.Request.Context(), userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

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

	profile, err := h.profileSvc.UploadAvatar(c.Request.Context(), userID, reader, fileHeader.Size, contentType, filename)
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
		"avatar_url":      profile.AvatarURL,
		"file_size":       fileHeader.Size,
		"last_updated_at": lastUpdated,
		"user":            mapUserToResponse(profile),
	}))
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
		ID:                user.ID,
		Email:             user.Email,
		Name:              user.Name,
		AvatarURL:         user.AvatarURL,
		Role:              user.Role.String(),
		Bio:               user.Bio,
		ProfileVisibility: user.ProfileVisibility.String(),
		IsVerifiedEmail:   user.IsVerifiedEmail,
		RoleVerified:      user.RoleVerified,
		OAuthProvider:      user.OAuthProvider,
		CreatedAt:          user.CreatedAt.Format(time.RFC3339),
		// Student/Alumni fields
		StudentID:          user.StudentID,
		School:             user.School,
		Faculty:            user.Faculty,
		Major:              user.Major,
		WeeklyHours:        user.WeeklyHours,
		WeeklyAvailability: user.WeeklyAvailability,
		Skills:             user.Skills,
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
	c.JSON(http.StatusOK, response.Success(publicProfile))
}

// GetAvailabilityHandler returns availability information for the authenticated user
func (h *AuthHandler) GetAvailabilityHandler(c *gin.Context) {
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

	c.JSON(http.StatusOK, response.Success(gin.H{
		"weekly_hours":        user.WeeklyHours,
		"weekly_availability": user.WeeklyAvailability,
	}))
}

// UpdateAvailabilityHandler updates availability information
func (h *AuthHandler) UpdateAvailabilityHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, errs.ErrUnauthorized, "unauthorized")
		return
	}

	var req struct {
		WeeklyHours        *int                        `json:"weekly_hours"`
		WeeklyAvailability *[]model.WeeklyAvailability `json:"weekly_availability"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	input := service.UpdateProfileInput{
		WeeklyHours: req.WeeklyHours,
	}
	if req.WeeklyAvailability != nil {
		wa := model.WeeklyAvailabilityArray(*req.WeeklyAvailability)
		input.WeeklyAvailability = &wa
	}

	user, err := h.profileSvc.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"weekly_hours":        user.WeeklyHours,
		"weekly_availability": user.WeeklyAvailability,
	}))
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
		c.JSON(http.StatusOK, response.Success(items))
		return
	}

	items, err := h.savedItemSvc.GetSavedItems(c.Request.Context(), userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(items))
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

	var req struct {
		ItemID   int64  `json:"item_id" binding:"required"`
		ItemType string `json:"item_type" binding:"required,oneof=project opportunity user event"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, errs.ErrInvalidInput, err.Error())
		return
	}

	itemType := model.SavedItemType(req.ItemType)
	if err := h.savedItemSvc.UnsaveItem(c.Request.Context(), userID, req.ItemID, itemType); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "item unsaved successfully"}))
}

// mapUserToPublicResponse maps user to public response (excludes sensitive fields)
func mapUserToPublicResponse(user *model.User) gin.H {
	resp := gin.H{
		"id":                 user.ID,
		"name":               user.Name,
		"avatar_url":         user.AvatarURL,
		"role":               user.Role.String(),
		"bio":                user.Bio,
		"profile_visibility": user.ProfileVisibility.String(),
		"created_at":         user.CreatedAt.Format(time.RFC3339),
	}

	// Add student/alumni fields if applicable
	if user.Role.IsStudentRole() {
		resp["school"] = user.School
		resp["faculty"] = user.Faculty
		resp["major"] = user.Major
		resp["weekly_hours"] = user.WeeklyHours
		resp["weekly_availability"] = user.WeeklyAvailability
		resp["skills"] = user.Skills
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
