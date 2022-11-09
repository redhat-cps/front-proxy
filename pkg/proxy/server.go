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
	"os"
	"time"

	coreinformers "github.com/kcp-dev/client-go/informers"
	coreclient "github.com/kcp-dev/client-go/kubernetes"
	kcpproxy "github.com/kcp-dev/kcp/pkg/proxy"
	kcpproxyfilters "github.com/kcp-dev/kcp/pkg/proxy/filters"

	"github.com/redhat-cps/front-proxy/pkg/proxy/usergating"
	"github.com/redhat-cps/front-proxy/pkg/proxy/usergating/ruleset"
)

type Server struct {
	KCPProxyServer            *kcpproxy.Server
	CoreSharedInformerFactory coreinformers.SharedInformerFactory
	CompletedConfig           CompletedConfig
	UserGateController        *usergating.Controller
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

	if c.Options.UserGate.Enabled {
		// setup the shared informer factory

		// TODO(csams): we could (should?) use some workspace other than root
		rootShardCoreInformerClient, err := coreclient.NewForConfig(s.CompletedConfig.KCPConfig.RootShardConfig)
		if err != nil {
			return s, fmt.Errorf("failed to create client for informers: %w", err)
		}
		s.CoreSharedInformerFactory = coreinformers.NewSharedInformerFactoryWithOptions(rootShardCoreInformerClient, 30*time.Minute)

		var rules *ruleset.RuleSet
		if c.Options.UserGate.DefaultRulesFile != "" {
			// load the default rule set from a file passed in options
			data, err := os.ReadFile(c.Options.UserGate.DefaultRulesFile)
			if err != nil {
				return nil, err
			}
			rules, err = ruleset.FromYAML(data)
			if err != nil {
				return nil, err
			}
		}
		// create the controller to watch for updates to the usergating secret in root:kcp-system
		s.UserGateController = usergating.NewController(ctx, s.CoreSharedInformerFactory, rules)
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
	// optionally start the user gate controller and wrap preparedServer.Handler with the
	// gate filter
	if s.CompletedConfig.Options.UserGate.Enabled {
		go s.UserGateController.Start(ctx, 2)
		s.CoreSharedInformerFactory.Start(ctx.Done())
		s.CoreSharedInformerFactory.WaitForCacheSync(ctx.Done())

		unauthorizedHandler := kcpproxyfilters.NewUnauthorizedHandler()
		s.KCPProxyServer.Handler = usergating.WithUserGating(s.KCPProxyServer.Handler, unauthorizedHandler, s.UserGateController.GetRuleSet)
	}

	// installs the default handler chain around preparedServer.Handler and
	// actually starts the server
	return s.RunKCPProxyServer(ctx)
}
