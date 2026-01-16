package handler

import (
	"strings"

	"github.com/aatist/backend/internal/community/model"
)

type listPostsQuery struct {
	Category string `form:"category"`
	Limit    int    `form:"limit"`
	Offset   int    `form:"offset"`
	Search   string `form:"search"`
	Sort     string `form:"sort"`
}

type postMutationRequest struct {
	Title    string   `json:"title" binding:"required,min=3"`
	Content  string   `json:"content" binding:"required"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

type commentMutationRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID *int64 `json:"parent_id"`
}

func normalizeCategory(category string) model.PostCategory {
	if cat, ok := parseCategory(category); ok {
		return cat
	}
	return model.PostCategoryGeneral
}

func parseCategory(category string) (model.PostCategory, bool) {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "":
		return model.PostCategoryGeneral, false
	case "general":
		return model.PostCategoryGeneral, true
	case "study_tips", "study-tips", "studytips":
		return model.PostCategoryStudyTips, true
	case "events":
		return model.PostCategoryEvents, true
	case "housing":
		return model.PostCategoryHousing, true
	case "food", "cafes", "food_cafes", "food-cafes":
		return model.PostCategoryFoodCafes, true
	case "sports", "hobbies", "sports_hobbies", "sports-hobbies":
		return model.PostCategorySportsHobbies, true
	case "random":
		return model.PostCategoryRandom, true
	case "projects":
		return model.PostCategoryProjects, true
	case "other":
		return model.PostCategoryOther, true
	case "sticky":
		return model.PostCategorySticky, true
	default:
		return model.PostCategoryGeneral, false
	}
}
