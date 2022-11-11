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

package proxy

import (
	"context"
	"fmt"

	kcpproxy "github.com/kcp-dev/kcp/pkg/proxy"

	"github.com/redhat-cps/front-proxy/pkg/proxy/filters"
)

type Server struct {
	KCPProxyServer  *kcpproxy.Server
	CompletedConfig CompletedConfig
}

func NewServer(ctx context.Context, c CompletedConfig) (*Server, error) {
	kcpServer, err := kcpproxy.NewServer(ctx, *c.KCPConfig)
	if err != nil {
		return nil, err
	}

	s := &Server{
		KCPProxyServer:  kcpServer,
		CompletedConfig: c,
	}

	return s, nil
}

// preparedServer is a private wrapper that enforces a call of PrepareRun() before Run can be invoked.
type preparedServer struct {
	*Server
	RunKCPProxyServer func(context.Context) error
}

// PrepareRun is basically a no-op, but we contort a little to get the Run func
// out of the KCP proxy delegate
func (s *Server) PrepareRun(ctx context.Context) (preparedServer, error) {
	p, err := s.KCPProxyServer.PrepareRun(ctx)
	if err != nil {
		return preparedServer{}, err
	}
	return preparedServer{
		Server:            s,
		RunKCPProxyServer: p.Run,
	}, nil
}

func (s preparedServer) Run(ctx context.Context) error {
	// optionally rate-limit requests per user
	if s.CompletedConfig.Options.RateLimit.RequestLimit != 0 {
		rateLimitFilter, err := filters.NewRateLimitFilter(
			s.CompletedConfig.Options.RateLimit.RequestLimit,
			s.CompletedConfig.Options.RateLimit.BurstLimit,
			s.CompletedConfig.Options.RateLimit.ExcludePattern)
		s.KCPProxyServer.Handler = rateLimitFilter.WithRateLimitAuthenticatedUser(s.KCPProxyServer.Handler)
		if err != nil {
			return fmt.Errorf("failed to create RateLimitFilter: %w", err)
		}
		go rateLimitFilter.PeriodicCleanup(ctx)
	}

	// installs the default handler chain around preparedServer.Handler and
	// actually starts the server
	return s.RunKCPProxyServer(ctx)
}
