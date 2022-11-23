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

package options

import (
	"fmt"
	"regexp"

	"github.com/spf13/pflag"
)

const (
	defaultBurstLimit = 5
)

type RateLimit struct {
	RequestLimit   int
	BurstLimit     int
	ExcludePattern string
}

func NewRateLimit() *RateLimit {
	return &RateLimit{
		RequestLimit:   0,
		BurstLimit:     defaultBurstLimit,
		ExcludePattern: "",
	}
}

func (o *RateLimit) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&o.RequestLimit, "ratelimit-request-limit", o.RequestLimit, "Rate limit requests (per user, per second) when non-zero.")
	fs.IntVar(&o.BurstLimit, "ratelimit-burst-limit", defaultBurstLimit, "Rate limit burst limit (bucket size).")
	fs.StringVar(&o.ExcludePattern, "ratelimit-exclude-pattern", "", "Regex for usernames to exclude from rate limiting.")
}

func (o *RateLimit) Validate() []error {
	var errs []error
	if o.ExcludePattern != "" {
		_, err := regexp.Compile(o.ExcludePattern)
		if err != nil {
			errs = append(errs, fmt.Errorf("Invalid regex passed to --ratelimit-exclude-pattern: %v", err))
		}
	}
	return errs
}
