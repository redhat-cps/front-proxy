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
)

type UserGate struct {
	Enabled          bool
	DefaultRulesFile string
}

func NewUserGate() *UserGate {
	return &UserGate{
		Enabled:          false,
		DefaultRulesFile: "",
	}
}

func (o *UserGate) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "enable-user-gating", false, "Enabled user gating.")
	fs.StringVar(&o.DefaultRulesFile, "user-gating-rules", "", "Provide a YAML file containing user gating rules.")
}

func (c *UserGate) Validate() []error {
	return nil
}
