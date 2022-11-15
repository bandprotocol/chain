package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

var cacher = cache.New(time.Hour, time.Hour)

func NewRateLimitMiddleware(keyFunc func(*gin.Context) (string, error)) gin.HandlerFunc {
	return func(gc *gin.Context) {
		// get key for cache
		key, err := keyFunc(gc)
		if err != nil {
			gc.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// set limiter for the key
		limiter, ok := cacher.Get(key)
		if !ok {
			// 1 request per hour
			limiter = rate.NewLimiter(rate.Every(time.Hour), 1)
			cacher.Set(key, limiter, 0)
		}

		// check if it reaches limit or not
		ok = limiter.(*rate.Limiter).Allow()
		if !ok {
			gc.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		gc.Next()
	}
}
