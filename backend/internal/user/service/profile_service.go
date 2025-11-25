package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/aalto-talent-network/backend/internal/platform/cache"
	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/aalto-talent-network/backend/internal/platform/metrics"
	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/internal/user/repository"
	"github.com/aalto-talent-network/backend/pkg/errs"
	"go.uber.org/zap"
)

// UpdateProfileInput represents profile fields that can be updated.
type UpdateProfileInput struct {
	Name               *string
	Nickname           *string
	AvatarURL          *string
	StudentID          *string
	School             *string
	Faculty            *string
	Major              *string
	WeeklyHours        *int
	EmotionalStatus    *string
	WeeklyAvailability *model.WeeklyAvailabilityArray
	Bio                *string
	ProfileVisibility  *model.ProfileVisibility
	Skills             *[]model.Skill
}

type ProfileService interface {
	GetProfile(ctx context.Context, userID int64) (*model.User, error)
	GetUserSummary(ctx context.Context, userID int64) (*UserSummary, error)
	UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (*model.User, error)
	UploadAvatar(ctx context.Context, userID int64, reader io.Reader, size int64, contentType, filename string) (*model.User, error)
	EnsureProfileUpdateRateLimit(ctx context.Context, userID int64) error
	EnsureAvatarUploadRateLimit(ctx context.Context, userID int64) error
	FilterSkillsByPrefix(ctx context.Context, userID int64, prefix string) ([]model.Skill, error)
}

type profileService struct {
	userRepo          repository.UserRepository
	fileClient        FileServiceClient // Changed from storage to fileClient
	redis             *cache.Redis
	logger            *log.Logger
	avatarURLPrefix   string
	profileRateLimit  rateLimitConfig
	avatarUploadLimit rateLimitConfig
}

type rateLimitConfig struct {
	KeyFormat string
	Limit     int
	Window    time.Duration
}

// NewProfileService creates a new profile service.
func NewProfileService(
	userRepo repository.UserRepository,
	fileClient FileServiceClient, // Changed from storage to fileClient
	redis *cache.Redis,
	logger *log.Logger,
	avatarURLPrefix string,
) ProfileService {
	if avatarURLPrefix != "" {
		avatarURLPrefix = strings.TrimRight(avatarURLPrefix, "/")
	}
	return &profileService{
		userRepo:        userRepo,
		fileClient:      fileClient,
		redis:           redis,
		logger:          logger,
		avatarURLPrefix: avatarURLPrefix,
		profileRateLimit: rateLimitConfig{
			KeyFormat: "rate:profile_update:%d",
			Limit:     5,
			Window:    time.Minute,
		},
		avatarUploadLimit: rateLimitConfig{
			KeyFormat: "rate:avatar:%d",
			Limit:     3,
			Window:    time.Minute,
		},
	}
}

func (s *profileService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *profileService) GetUserSummary(ctx context.Context, userID int64) (*UserSummary, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserSummary{
		ID:        user.ID,
		Name:      user.Name,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
		Role:      user.Role.String(),
		School:    user.School,
		Major:     user.Major,
	}, nil
}

func (s *profileService) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (*model.User, error) {
	if input.AvatarURL != nil {
		return nil, errs.NewAppError(errors.New("avatar update forbidden"), http.StatusBadRequest, "avatar_url is not directly editable")
	}

	fields := make(map[string]interface{})

	if input.Name != nil {
		fields["name"] = valueOrNil(normalizeOptionalStringWithLimit(input.Name, 100))
	}
	if input.Nickname != nil {
		fields["nickname"] = valueOrNil(normalizeOptionalStringWithLimit(input.Nickname, 100))
	}
	if input.StudentID != nil {
		fields["student_id"] = valueOrNil(normalizeOptionalStringWithLimit(input.StudentID, 64))
	}
	if input.School != nil {
		fields["school"] = valueOrNil(normalizeOptionalStringWithLimit(input.School, 255))
	}
	if input.Faculty != nil {
		fields["faculty"] = valueOrNil(normalizeOptionalStringWithLimit(input.Faculty, 255))
	}
	if input.Major != nil {
		fields["major"] = valueOrNil(normalizeMajor(input.Major))
	}
	if input.Bio != nil {
		fields["bio"] = valueOrNil(normalizeOptionalStringWithLimit(input.Bio, maxBioLength))
	}
	if input.Skills != nil {
		normalizedSkills := sanitizeSkillsWithLevel(*input.Skills)
		skills := model.Skills(normalizedSkills)
		fields["skills"] = skills
	}
	if input.WeeklyHours != nil {
		fields["weekly_hours"] = *input.WeeklyHours
	}
	if input.EmotionalStatus != nil {
		fields["emotional_status"] = valueOrNil(normalizeOptionalStringWithLimit(input.EmotionalStatus, 50))
	}
	if input.WeeklyAvailability != nil {
		fields["weekly_availability"] = *input.WeeklyAvailability
	}
	if input.ProfileVisibility != nil {
		if input.ProfileVisibility.IsValid() {
			fields["profile_visibility"] = input.ProfileVisibility.String()
		}
	}

	if len(fields) == 0 {
		return s.userRepo.FindByID(ctx, userID)
	}

	return s.userRepo.UpdateProfile(ctx, repository.ProfileUpdate{
		UserID: userID,
		Fields: fields,
	})
}

func (s *profileService) UploadAvatar(ctx context.Context, userID int64, reader io.Reader, size int64, contentType, filename string) (*model.User, error) {
	if s.fileClient == nil {
		return nil, fmt.Errorf("file service client not configured")
	}
	if size <= 0 {
		return nil, fmt.Errorf("invalid file size")
	}

	if _, err := s.userRepo.FindByID(ctx, userID); err != nil {
		return nil, err
	}

	// Upload file via file-service
	fileResp, err := s.fileClient.UploadFile(ctx, userID, "avatar", reader, size, contentType, filename)
	if err != nil {
		metrics.AvatarUploadFailureTotal.Inc()
		if s.logger != nil {
			s.logger.Error("avatar upload failed",
				zap.Int64("user_id", userID),
				zap.String("content_type", contentType),
				zap.Int64("size", size),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	url := fileResp.Data.URL
	if s.avatarURLPrefix != "" && !strings.HasPrefix(url, s.avatarURLPrefix) {
		metrics.AvatarUploadFailureTotal.Inc()
		return nil, fmt.Errorf("unexpected avatar url domain")
	}

	updatedUser, err := s.userRepo.UpdateAvatarURL(ctx, userID, url)
	if err != nil {
		metrics.AvatarUploadFailureTotal.Inc()
		return nil, err
	}

	metrics.AvatarUploadSuccessTotal.Inc()
	if s.logger != nil {
		s.logger.Info("avatar uploaded",
			zap.Int64("user_id", userID),
			zap.String("file_id", fmt.Sprintf("%d", fileResp.Data.ID)),
			zap.String("content_type", contentType),
			zap.Int64("size", size),
		)
	}

	return updatedUser, nil
}

func (s *profileService) EnsureProfileUpdateRateLimit(ctx context.Context, userID int64) error {
	return s.enforceRateLimit(ctx, s.profileRateLimit, userID)
}

func (s *profileService) EnsureAvatarUploadRateLimit(ctx context.Context, userID int64) error {
	return s.enforceRateLimit(ctx, s.avatarUploadLimit, userID)
}

func (s *profileService) FilterSkillsByPrefix(ctx context.Context, userID int64, prefix string) ([]model.Skill, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if prefix == "" {
		return user.Skills, nil
	}

	prefix = strings.ToLower(strings.TrimSpace(prefix))
	var matches []model.Skill
	for _, skill := range user.Skills {
		if strings.HasPrefix(strings.ToLower(skill.Name), prefix) {
			matches = append(matches, skill)
		}
	}
	// Sort by name
	sort.Slice(matches, func(i, j int) bool {
		return strings.ToLower(matches[i].Name) < strings.ToLower(matches[j].Name)
	})
	return matches, nil
}

func (s *profileService) enforceRateLimit(ctx context.Context, cfg rateLimitConfig, userID int64) error {
	if s.redis == nil {
		return nil
	}
	client := s.redis.GetClient()
	key := fmt.Sprintf(cfg.KeyFormat, userID)

	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("rate limit check failed", zap.Error(err))
		}
		return nil
	}

	if count == 1 {
		client.Expire(ctx, key, cfg.Window)
	}

	if count > int64(cfg.Limit) {
		return errs.NewAppError(errs.ErrRateLimitExceeded, http.StatusTooManyRequests, "too many requests, please try again later")
	}

	return nil
}

func sanitizeSkillsWithLevel(skills []model.Skill) []model.Skill {
	seen := make(map[string]struct{})
	var normalized []model.Skill
	for _, skill := range skills {
		name := strings.ToLower(strings.TrimSpace(skill.Name))
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}

		// Normalize level
		level := strings.ToLower(strings.TrimSpace(skill.Level))
		if level != "expert" && level != "advanced" && level != "intermediate" {
			level = "intermediate" // Default
		}

		normalized = append(normalized, model.Skill{
			Name:  name,
			Level: level,
		})
	}
	// Sort by name
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].Name < normalized[j].Name
	})
	return normalized
}

func sanitizeStringArray(arr []string) []string {
	seen := make(map[string]struct{})
	var normalized []string
	for _, item := range arr {
		trimmed := strings.ToLower(strings.TrimSpace(item))
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	sort.Strings(normalized)
	return normalized
}

func valueOrNil(value *string) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func normalizeProjects(projects []model.Project) []model.Project {
	nowYear := time.Now().Year() + 1
	normalized := make([]model.Project, 0, len(projects))
	for _, project := range projects {
		title := strings.TrimSpace(project.Title)
		description := strings.TrimSpace(project.Description)
		if title == "" || description == "" {
			continue
		}

		var year *int
		if project.Year != nil && *project.Year >= 1900 && *project.Year <= nowYear {
			y := *project.Year
			year = &y
		}

		normalized = append(normalized, model.Project{
			Title:       title,
			ClientName:  normalizeString(project.ClientName),
			Description: description,
			Tags:        sanitizeStringArray(project.Tags),
			Year:        year,
		})
	}
	return normalized
}

func normalizeMajor(value *string) *string {
	trimmed := normalizeOptionalStringWithLimit(value, 255)
	if trimmed == nil {
		return nil
	}
	lower := strings.ToLower(*trimmed)
	formatted := strings.Title(lower)
	return &formatted
}

func normalizeString(value string) string {
	return strings.TrimSpace(value)
}

func normalizeOptionalString(value *string) *string {
	return normalizeOptionalStringWithLimit(value, 0)
}

const maxBioLength = 2000

func normalizeOptionalStringWithLimit(value *string, maxLen int) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	if maxLen > 0 && len(trimmed) > maxLen {
		trimmed = trimmed[:maxLen]
	}
	return &trimmed
}

func guessExt(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
