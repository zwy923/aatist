package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// WaitlistHandler serves the waitlist signup form and ad-attribution
// tracking used by the temporary user-research landing page. It talks to
// the database directly rather than going through a repository/service
// layer, since this is a short-lived experiment rather than a long-term
// domain.
type WaitlistHandler struct {
	db *sqlx.DB
}

func NewWaitlistHandler(db *sqlx.DB) *WaitlistHandler {
	return &WaitlistHandler{db: db}
}

// tracking holds the ad-attribution fields shared by page views and
// waitlist signups: UTM params, referrer, and user agent.
type tracking struct {
	UTMSource   string `json:"utm_source"`
	UTMMedium   string `json:"utm_medium"`
	UTMCampaign string `json:"utm_campaign"`
	UTMContent  string `json:"utm_content"`
	UTMTerm     string `json:"utm_term"`
	Referrer    string `json:"referrer"`
	UserAgent   string `json:"user_agent"`
}

type createWaitlistRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Interest string `json:"interest" binding:"required,oneof=talent opportunity"`
	tracking
}

type consentRequest struct {
	Consent string `json:"consent" binding:"required,oneof=agreed declined"`
}

type pageViewRequest struct {
	Tab string `json:"tab" binding:"required,oneof=talent opportunity"`
	tracking
}

// nullIfEmpty converts an unset (zero-value) tracking field to SQL NULL
// instead of storing an empty string, so "no attribution data" reads
// consistently across old and new rows.
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (h *WaitlistHandler) CreateHandler(c *gin.Context) {
	var req createWaitlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int64
	err := h.db.QueryRowxContext(c.Request.Context(),
		`INSERT INTO waitlist_entries
			(email, interest, utm_source, utm_medium, utm_campaign, utm_content, utm_term, referrer, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		req.Email, req.Interest, nullIfEmpty(req.UTMSource), nullIfEmpty(req.UTMMedium), nullIfEmpty(req.UTMCampaign),
		nullIfEmpty(req.UTMContent), nullIfEmpty(req.UTMTerm), nullIfEmpty(req.Referrer), nullIfEmpty(req.UserAgent),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join waitlist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *WaitlistHandler) ConsentHandler(c *gin.Context) {
	id := c.Param("id")

	var req consentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.db.ExecContext(c.Request.Context(),
		`UPDATE waitlist_entries SET consent = $1, updated_at = NOW() WHERE id = $2`,
		req.Consent, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record consent"})
		return
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "waitlist entry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// PageViewHandler logs a visit to the landing page regardless of whether the
// visitor signs up, so conversion rate (page views -> waitlist signups) can
// be computed per ad/campaign.
func (h *WaitlistHandler) PageViewHandler(c *gin.Context) {
	var req pageViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.ExecContext(c.Request.Context(),
		`INSERT INTO page_views
			(utm_source, utm_medium, utm_campaign, utm_content, utm_term, referrer, user_agent, tab)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		nullIfEmpty(req.UTMSource), nullIfEmpty(req.UTMMedium), nullIfEmpty(req.UTMCampaign), nullIfEmpty(req.UTMContent),
		nullIfEmpty(req.UTMTerm), nullIfEmpty(req.Referrer), nullIfEmpty(req.UserAgent), req.Tab,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log page view"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}
