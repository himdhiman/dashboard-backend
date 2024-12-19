package limiter

import (
	"net/http"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/cache"
)

func RateLimiterMiddleware(redisClient *cache.CacheClient, logger logger.Logger, serviceName string) func(http.Handler) http.Handler {
	limiterService := NewRateLimiter(redisClient, logger)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			endpoint := r.URL.Path

			allowed, err := limiterService.Allow(r.Context(), serviceName, endpoint)
			if err != nil || !allowed {
				logger.Warn("Rate limit exceeded for endpoint:", endpoint)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
