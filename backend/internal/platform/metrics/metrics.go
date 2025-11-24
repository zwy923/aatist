package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// LoginSuccessTotal counts successful logins
	LoginSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "login_success_total",
		Help: "Total number of successful logins",
	})

	// LoginFailureTotal counts failed logins
	LoginFailureTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "login_failure_total",
		Help: "Total number of failed logins",
	})

	// AccountLockedTotal counts account lock events
	AccountLockedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "account_locked_total",
		Help: "Total number of account lock events",
	})

	// JWTRefreshTotal counts JWT refresh operations
	JWTRefreshTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jwt_refresh_total",
		Help: "Total number of JWT refresh operations",
	})

	// RegisterSuccessTotal counts successful registrations
	RegisterSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "register_success_total",
		Help: "Total number of successful registrations",
	})

	// AvatarUploadSuccessTotal counts successful avatar uploads
	AvatarUploadSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "avatar_upload_success_total",
		Help: "Total number of successful avatar uploads",
	})

	// AvatarUploadFailureTotal counts failed avatar uploads
	AvatarUploadFailureTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "avatar_upload_failure_total",
		Help: "Total number of failed avatar uploads",
	})
)
