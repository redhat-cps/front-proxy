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
	kcpproxyoptions "github.com/kcp-dev/kcp/pkg/proxy/options"
	"github.com/spf13/pflag"
)

type Options struct {
	KCPProxyOptions *kcpproxyoptions.Options
	RateLimit       *RateLimit
}

func NewOptions() *Options {
	o := &Options{
		KCPProxyOptions: kcpproxyoptions.NewOptions(),
		RateLimit:       NewRateLimit(),
	}
	return o
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.KCPProxyOptions.AddFlags(fs)
	o.RateLimit.AddFlags(fs)
}

func (o *Options) Complete() error {
	return o.KCPProxyOptions.Complete()
}

func (o *Options) Validate() []error {
	var errs []error

	errs = append(errs, o.KCPProxyOptions.Validate()...)
	errs = append(errs, o.RateLimit.Validate()...)

	return errs
}
