package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// PostCategory represents the category of a discussion post.
type PostCategory string

const (
	PostCategoryGeneral       PostCategory = "general"
	PostCategoryStudyTips     PostCategory = "study_tips"
	PostCategoryEvents        PostCategory = "events"
	PostCategoryProjects      PostCategory = "projects"
	PostCategoryFoodCafes     PostCategory = "food_cafes"
	PostCategoryHousing       PostCategory = "housing"
	PostCategorySportsHobbies PostCategory = "sports_hobbies"
	PostCategoryRandom        PostCategory = "random"
	PostCategoryOther         PostCategory = "other"
	PostCategorySticky        PostCategory = "sticky"
)

// validPostCategories keeps the allowed category list aligned with DB-level CHECK constraints.
var validPostCategories = map[PostCategory]struct{}{
	PostCategoryGeneral:       {},
	PostCategoryStudyTips:     {},
	PostCategoryEvents:        {},
	PostCategoryProjects:      {},
	PostCategoryFoodCafes:     {},
	PostCategoryHousing:       {},
	PostCategorySportsHobbies: {},
	PostCategoryRandom:        {},
	PostCategoryOther:         {},
	PostCategorySticky:        {},
}

// IsValid reports whether the category is part of the supported list.
func (pc PostCategory) IsValid() bool {
	_, ok := validPostCategories[pc]
	return ok
}

// IsSticky indicates whether the post should be treated as a sticky note (special display).
func (pc PostCategory) IsSticky() bool {
	return pc == PostCategorySticky
}

// StringArray is a JSONB-backed slice of strings.
type StringArray []string

func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(sa)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for StringArray: %T", value)
	}

	if len(bytes) == 0 {
		*sa = nil
		return nil
	}

	var temp []string
	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}
	*sa = temp
	return nil
}

// DiscussionPost represents a community discussion post.
type DiscussionPost struct {
	ID           int64        `db:"id" json:"id"`
	UserID       int64        `db:"user_id" json:"user_id"`
	Title        string       `db:"title" json:"title"`
	Content      string       `db:"content" json:"content"`
	Category     PostCategory `db:"category" json:"category"`
	Tags         StringArray  `db:"tags" json:"tags"`
	LikeCount    int64        `db:"like_count" json:"like_count"`
	CommentCount int64        `db:"comment_count" json:"comment_count"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`

	// Enriched fields (joined from users table)
	AuthorName    string `db:"author_name" json:"author_name"`
	AuthorAvatar  string `db:"author_avatar" json:"author_avatar"`
	AuthorFaculty string `db:"author_faculty" json:"author_faculty"`

	// Client-side state
	HasLiked bool `db:"-" json:"has_liked"`
}
