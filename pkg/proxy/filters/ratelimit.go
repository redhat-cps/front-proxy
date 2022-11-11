/*
Copyright 2022 The KCP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filters

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog/v2"
)

const (
	// Do we need the Retry-After value to be user configurable?
	retryAfter      = "1"
	cleanupInterval = 1
)

type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimitFilter struct {
	RequestLimit int
	BurstLimit   int
	excludeRegex *regexp.Regexp
	limiters     map[string]*userLimiter
	lock         sync.Mutex
}

func NewRateLimitFilter(requestLimit, burstLimit int, excludePattern string) (*RateLimitFilter, error) {
	var re *regexp.Regexp
	if excludePattern != "" {
		var err error
		re, err = regexp.Compile(excludePattern)
		if err != nil {
			return nil, err
		}
	}
	return &RateLimitFilter{
		RequestLimit: requestLimit,
		BurstLimit:   burstLimit,
		excludeRegex: re,
		limiters:     make(map[string]*userLimiter),
	}, nil
}

// Get a limiter for a given user if it exists,
// otherwise create one and store in the map
func (r *RateLimitFilter) getLimiter(user string) *userLimiter {
	r.lock.Lock()
	defer r.lock.Unlock()

	limiter, exists := r.limiters[user]
	if !exists {
		newLimiter := rate.NewLimiter(rate.Limit(r.RequestLimit), r.BurstLimit)
		limiter = &userLimiter{newLimiter, time.Now()}
		r.limiters[user] = limiter
	}
	limiter.lastSeen = time.Now()
	return limiter
}

func (r *RateLimitFilter) PeriodicCleanup(ctx context.Context) {
	logger := klog.FromContext(ctx)
	if err := wait.PollInfiniteWithContext(ctx, time.Minute, func(ctx context.Context) (bool, error) {
		r.lock.Lock()
		for user, limiter := range r.limiters {
			if time.Since(limiter.lastSeen) > cleanupInterval*time.Minute {
				logger.WithValues("user", user).Info("removing limiter for user")
				delete(r.limiters, user)
			}
		}
		r.lock.Unlock()
		return false, nil
	}); err != nil {
		logger.Error(err, "error polling for RateLimitFilter cleanup")
	}
}

func (r *RateLimitFilter) rateLimitUser(user string) bool {
	if r.excludeRegex != nil {
		if r.excludeRegex.MatchString(user) {
			return false
		}
	}
	return true
}

// WithRateLimitAuthenticatedUser limits the number of requests per-user
func (r *RateLimitFilter) WithRateLimitAuthenticatedUser(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, ok := request.UserFrom(ctx)
		if !ok {
			logger := klog.FromContext(ctx).WithValues("proxy", "ratelimiter")
			logger.Error(errors.New("no user in context"), "can't detect user from context")
			return
		}
		userName := user.GetName()
		if r.rateLimitUser(userName) {
			ul := r.getLimiter(userName)
			if !ul.limiter.Allow() {
				tooManyRequests(req, w)
				return
			}
		}
		handler.ServeHTTP(w, req)
	})
}

func tooManyRequests(req *http.Request, w http.ResponseWriter) {
	// Return a 429 status indicating "Too Many Requests"
	w.Header().Set("Retry-After", retryAfter)
	http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
}
