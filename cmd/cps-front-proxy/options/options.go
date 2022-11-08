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
	"github.com/spf13/pflag"

	"k8s.io/component-base/config"
	"k8s.io/component-base/logs"

	cpsproxyoptions "github.com/redhat-cps/front-proxy/pkg/proxy/options"
)

/*
Options are organized in a tree with these at the root since they're for the
command that is used to start the proxy process. The CPS proxy options are
specific to CPS, but they also reference the options provided by the proxy
implemention in kcp.
*/
type Options struct {
	CPSProxy *cpsproxyoptions.Options
	Logs     *logs.Options
}

func NewOptions() *Options {
	o := &Options{
		CPSProxy: cpsproxyoptions.NewOptions(),
		Logs:     logs.NewOptions(),
	}

	// Default to -v=2
	o.Logs.Config.Verbosity = config.VerbosityLevel(2)
	return o
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.CPSProxy.AddFlags(fs)
	o.Logs.AddFlags(fs)
}

func (o *Options) Complete() error {
	if err := o.CPSProxy.Complete(); err != nil {
		return err
	}

	return nil
}

func (o *Options) Validate() []error {
	var errs []error

	errs = append(errs, o.CPSProxy.Validate()...)

	return errs
}
