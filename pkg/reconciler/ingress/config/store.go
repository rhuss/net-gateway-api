/*
Copyright 2021 The Knative Authors

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

package config

import (
	"context"

	network "knative.dev/networking/pkg"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
)

type cfgKey struct{}

// Config is the configuration for the route reconciler.
type Config struct {
	Network *network.Config
	Gateway *Gateway
}

// FromContext obtains a Config injected into the passed context.
func FromContext(ctx context.Context) *Config {
	return ctx.Value(cfgKey{}).(*Config)
}

// FromContextOrDefaults is like FromContext, but when no Config is attached it
// returns a Config populated with the defaults for each of the Config fields.
func FromContextOrDefaults(ctx context.Context) *Config {
	cfg := FromContext(ctx)
	if cfg == nil {
		cfg = &Config{}
	}
	return cfg
}

// ToContext stores the configuration Config in the passed context.
func ToContext(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, cfgKey{}, c)
}

// Store is a typed wrapper around configmap.Untyped store to handle our configmaps.
//
// +k8s:deepcopy-gen=false
type Store struct {
	*configmap.UntypedStore
}

// NewStore creates a configmap.UntypedStore based config store.
//
// logger must be non-nil implementation of configmap.Logger (commonly used
// loggers conform)
//
// onAfterStore is a variadic list of callbacks to run
// after the ConfigMap has been processed and stored.
//
// See also: configmap.NewUntypedStore().
func NewStore(ctx context.Context, onAfterStore ...func(name string, value interface{})) *Store {
	logger := logging.FromContext(ctx)

	store := &Store{
		UntypedStore: configmap.NewUntypedStore(
			"gateway-api",
			logger,
			configmap.Constructors{
				GatewayConfigName:  NewGatewayFromConfigMap,
				network.ConfigName: network.NewConfigFromConfigMap,
			},
			onAfterStore...,
		),
	}

	return store
}

// ToContext stores the configuration Store in the passed context.
func (s *Store) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, s.Load())
}

// Load creates a Config for this store.
func (s *Store) Load() *Config {
	config := &Config{
		Gateway: s.UntypedLoad(GatewayConfigName).(*Gateway).DeepCopy(),
		Network: s.UntypedLoad(network.ConfigName).(*network.Config).DeepCopy(),
	}
	return config
}
