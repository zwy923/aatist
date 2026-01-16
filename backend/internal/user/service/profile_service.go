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

	"github.com/aatist/backend/internal/platform/cache"
	"github.com/aatist/backend/internal/platform/log"
	"github.com/aatist/backend/internal/platform/metrics"
	"github.com/aatist/backend/internal/user/model"
	"github.com/aatist/backend/internal/user/repository"
	"github.com/aatist/backend/pkg/errs"
	"go.uber.org/zap"
)

// UpdateProfileInput represents profile fields that can be updated.
type UpdateProfileInput struct {
	// Common fields
	Name              *string
	AvatarURL         *string
	Bio               *string
	ProfileVisibility *model.ProfileVisibility
	// Student/Alumni fields
	StudentID           *string
	School              *string
	Faculty             *string
	Major               *string
	WeeklyHours         *int
	WeeklyAvailability  *model.WeeklyAvailabilityArray
	Skills              *[]model.Skill
	PortfolioVisibility *model.PortfolioVisibility
	// Organization fields
	OrganizationName       *string
	OrganizationBio        *string
	ContactTitle           *string
	IsAffiliatedWithSchool *bool
	OrgSize                *int
}

type ProfileService interface {
	GetProfile(ctx context.Context, userID int64) (*model.User, error)
	GetUserSummary(ctx context.Context, userID int64) (*UserSummary, error)
	UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (*model.User, error)
	UploadAvatar(ctx context.Context, userID int64, reader io.Reader, size int64, contentType, filename string) (*model.User, error)
	EnsureProfileUpdateRateLimit(ctx context.Context, userID int64) error
	EnsureAvatarUploadRateLimit(ctx context.Context, userID int64) error
	FilterSkillsByPrefix(ctx context.Context, userID int64, prefix string) ([]model.Skill, error)
	SearchSkills(ctx context.Context, query string, limit int) ([]model.SkillMetadata, error)
	SearchCourses(ctx context.Context, query string, limit int) ([]model.CourseMetadata, error)
	SearchTags(ctx context.Context, tagType string, query string, limit int) ([]model.TagMetadata, error)
	SearchUsers(ctx context.Context, filter repository.UserSearchFilter) ([]*model.User, error)

	// User skills/courses maintenance
	AddUserSkill(ctx context.Context, userID int64, skill model.Skill) (*model.User, error)
	RemoveUserSkill(ctx context.Context, userID int64, skillName string) (*model.User, error)
	AddUserCourse(ctx context.Context, userID int64, course model.Course) (*model.User, error)
	RemoveUserCourse(ctx context.Context, userID int64, courseCode string) (*model.User, error)
	// Courses are currently not in the user model, I'll add them to the model first if needed,
	// but the request says POST /users/me/courses.
	// Looking at model.User, it doesn't have Courses. I should add it.
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
) (ProfileService, error) {
	if userRepo == nil {
		return nil, fmt.Errorf("userRepo is required")
	}
	if fileClient == nil {
		return nil, fmt.Errorf("fileClient is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

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
	}, nil
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

	// Get current user to check role
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]interface{})

	// Common fields (all roles)
	if input.Name != nil {
		fields["name"] = valueOrNil(normalizeOptionalStringWithLimit(input.Name, 100))
	}
	if input.Bio != nil {
		fields["bio"] = valueOrNil(normalizeOptionalStringWithLimit(input.Bio, maxBioLength))
	}
	if input.ProfileVisibility != nil {
		if input.ProfileVisibility.IsValid() {
			fields["profile_visibility"] = input.ProfileVisibility.String()
		}
	}

	// Student/Alumni fields (only for student/alumni roles)
	if user.Role.IsStudentRole() {
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
		if input.Skills != nil {
			normalizedSkills := sanitizeSkillsWithLevel(*input.Skills)
			skills := model.Skills(normalizedSkills)
			fields["skills"] = skills
		}
		if input.WeeklyHours != nil {
			fields["weekly_hours"] = *input.WeeklyHours
		}
		if input.WeeklyAvailability != nil {
			fields["weekly_availability"] = *input.WeeklyAvailability
		}
		if input.PortfolioVisibility != nil {
			if input.PortfolioVisibility.IsValid() {
				fields["portfolio_visibility"] = input.PortfolioVisibility.String()
			}
		}
	}

	// Organization fields (only for org roles)
	if user.Role.IsOrgRole() {
		if input.OrganizationName != nil {
			fields["organization_name"] = valueOrNil(normalizeOptionalStringWithLimit(input.OrganizationName, 255))
		}
		if input.OrganizationBio != nil {
			fields["organization_bio"] = valueOrNil(normalizeOptionalStringWithLimit(input.OrganizationBio, maxBioLength))
		}
		if input.ContactTitle != nil {
			fields["contact_title"] = valueOrNil(normalizeOptionalStringWithLimit(input.ContactTitle, 100))
		}
		if input.IsAffiliatedWithSchool != nil {
			fields["is_affiliated_with_school"] = *input.IsAffiliatedWithSchool
		}
		if input.OrgSize != nil {
			fields["org_size"] = *input.OrgSize
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
		panic("fileClient is nil in UploadAvatar")
	}
	if size <= 0 {
		return nil, fmt.Errorf("invalid file size")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Upload file via file-service
	fileResp, err := s.fileClient.UploadFile(ctx, userID, user.Role.String(), user.Email, "avatar", reader, size, contentType, filename)
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
	if s.redis == nil {
		// If you want to catch this as a panic for debugging:
		// panic("redis is nil in EnsureAvatarUploadRateLimit")
		s.logger.Warn("redis is nil, skipping rate limit", zap.Int64("user_id", userID))
		return nil
	}
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

func (s *profileService) SearchSkills(ctx context.Context, query string, limit int) ([]model.SkillMetadata, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.userRepo.SearchSkills(ctx, query, limit)
}

func (s *profileService) SearchCourses(ctx context.Context, query string, limit int) ([]model.CourseMetadata, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.userRepo.SearchCourses(ctx, query, limit)
}

func (s *profileService) SearchTags(ctx context.Context, tagType string, query string, limit int) ([]model.TagMetadata, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.userRepo.SearchTags(ctx, tagType, query, limit)
}

func (s *profileService) SearchUsers(ctx context.Context, filter repository.UserSearchFilter) ([]*model.User, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.userRepo.SearchUsers(ctx, filter)
}

func (s *profileService) AddUserSkill(ctx context.Context, userID int64, skill model.Skill) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if skill already exists
	for _, s := range user.Skills {
		if strings.EqualFold(s.Name, skill.Name) {
			return user, nil // Already exists
		}
	}

	user.Skills = append(user.Skills, skill)
	return s.userRepo.UpdateProfile(ctx, repository.ProfileUpdate{
		UserID: userID,
		Fields: map[string]interface{}{"skills": user.Skills},
	})
}

func (s *profileService) RemoveUserSkill(ctx context.Context, userID int64, skillName string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	newSkills := make(model.Skills, 0, len(user.Skills))
	found := false
	for _, s := range user.Skills {
		if strings.EqualFold(s.Name, skillName) {
			found = true
			continue
		}
		newSkills = append(newSkills, s)
	}

	if !found {
		return user, nil
	}

	return s.userRepo.UpdateProfile(ctx, repository.ProfileUpdate{
		UserID: userID,
		Fields: map[string]interface{}{"skills": newSkills},
	})
}

func (s *profileService) AddUserCourse(ctx context.Context, userID int64, course model.Course) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, c := range user.Courses {
		if strings.EqualFold(c.Code, course.Code) {
			return user, nil
		}
	}

	user.Courses = append(user.Courses, course)
	return s.userRepo.UpdateProfile(ctx, repository.ProfileUpdate{
		UserID: userID,
		Fields: map[string]interface{}{"courses": user.Courses},
	})
}

func (s *profileService) RemoveUserCourse(ctx context.Context, userID int64, courseCode string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	newCourses := make(model.Courses, 0, len(user.Courses))
	found := false
	for _, c := range user.Courses {
		if strings.EqualFold(c.Code, courseCode) {
			found = true
			continue
		}
		newCourses = append(newCourses, c)
	}

	if !found {
		return user, nil
	}

	return s.userRepo.UpdateProfile(ctx, repository.ProfileUpdate{
		UserID: userID,
		Fields: map[string]interface{}{"courses": newCourses},
	})
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
