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
	kcpproxy "github.com/kcp-dev/kcp/pkg/proxy"

	"github.com/redhat-cps/front-proxy/pkg/proxy/options"
)

type Config struct {
	Options   *options.Options
	KCPConfig *kcpproxy.Config
}

type completedConfig struct {
	Options   *options.Options
	KCPConfig *kcpproxy.CompletedConfig
}

type CompletedConfig struct {
	*completedConfig
}

func NewConfig(o *options.Options) (*Config, error) {
	kcpconfig, err := kcpproxy.NewConfig(o.KCPProxyOptions)
	if err != nil {
		return nil, err
	}
	return &Config{
		Options:   o,
		KCPConfig: kcpconfig,
	}, nil
}

func (c *Config) Complete() (CompletedConfig, error) {
	completedKcpConfig, err := c.KCPConfig.Complete()
	if err != nil {
		return CompletedConfig{}, nil
	}
	return CompletedConfig{
		&completedConfig{
			Options:   c.Options,
			KCPConfig: &completedKcpConfig,
		},
	}, nil
}
