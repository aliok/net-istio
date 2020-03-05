/*
Copyright 2019 The Knative Authors

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

package main

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/net-istio/pkg/defaults"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	defaultconfig "knative.dev/serving/pkg/apis/config"
)

// TODO(nghia): Validate config-istio
// func NewConfigValidationController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
// 	return configmaps.NewAdmissionController(ctx,

// 		// Name of the configmap webhook.
// 		fmt.Sprintf("config.webhook.%s.knative.dev", system.Namespace()),

// 		// The path on which to serve the webhook.
// 		"/config-validation",

// 		// The configmaps to validate.
// 		configmap.Constructors{
// 			logging.ConfigMapName(): logging.NewConfigFromConfigMap,
// 		},
// 	)
// }

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	appsv1.SchemeGroupVersion.WithKind("Deployment"): &defaults.IstioDeployment{},
}

// NewDefaultingAdmissionController adds default values to the watched types
func NewDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	// Decorate contexts with the current state of the config.
	store := defaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	store.WatchConfigs(cmw)

	return defaulting.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"webhook.istio.networking.internal.knative.dev",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func main() {
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "istio-webhook",
		Port:        8443,
		SecretName:  "istio-webhook-certs",
	})

	sharedmain.WebhookMainWithContext(
		ctx, "istio-webhook",
		certificates.NewController,
		NewDefaultingAdmissionController,
		// TODO(nghia): NewConfigValidationController,
	)
}
