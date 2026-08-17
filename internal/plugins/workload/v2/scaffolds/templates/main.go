// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/utils"
)

const (
	defaultMainPath = utils.DefaultMainPath

	importMarker    = "imports"
	addSchemeMarker = "scheme"
	setupMarker     = "reconcilers"
	webhookMarker   = "webhook"
)

var _ machinery.Template = &Main{}

// Main adds API-specific scaffolding to main.go.
type Main struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.DomainMixin
	machinery.RepositoryMixin
}

func (f *Main) SetTemplateDefaults() error {
	f.Path = defaultMainPath

	f.TemplateBody = fmt.Sprintf(mainTemplate,
		machinery.NewMarkerFor(f.Path, importMarker),
		machinery.NewMarkerFor(f.Path, addSchemeMarker),
		machinery.NewMarkerFor(f.Path, setupMarker),
		machinery.NewMarkerFor(f.Path, webhookMarker),
	)

	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

var _ machinery.Inserter = &MainUpdater{}

type MainUpdater struct {
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	// Flags to indicate which parts need to be included when updating the file
	WireResource, WireController, WireWebhook bool
}

func (*MainUpdater) GetPath() string {
	return defaultMainPath
}

func (*MainUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

func (f *MainUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(defaultMainPath, importMarker),
		machinery.NewMarkerFor(defaultMainPath, addSchemeMarker),
		machinery.NewMarkerFor(defaultMainPath, setupMarker),
		machinery.NewMarkerFor(defaultMainPath, webhookMarker),
	}
}

const (
	apiImportCodeFragment = `%s "%s"
`
	controllerImportCodeFragment = `"%s/controllers"
`
	multiGroupControllerImportCodeFragment = `%scontrollers "%s/controllers/%s"
`
	addschemeCodeFragment = `utilruntime.Must(%s.AddToScheme(scheme))
`
	reconcilerSetupCodeFragment = `controllers.New%sReconciler(mgr),
`
	multiGroupReconcilerSetupCodeFragment = `%scontrollers.New%sReconciler(mgr),
`
	webhookImportCodeFragment = `%s "%s/internal/webhook/%s"
`
	//nolint:gosec
	multiGroupWebhookImportCodeFragment = `%s "%s/internal/webhook/%s/%s"
`
	webhookSetupCodeFragment = `if err := %s.Setup%sWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "%s")
		os.Exit(1)
	}
`
)

//nolint:gocyclo
func (f *MainUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	const options = 4

	fragments := make(machinery.CodeFragmentsMap, options)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// Generate import code fragments
	imports := make([]string, 0)
	if f.WireResource {
		imports = append(imports, fmt.Sprintf(apiImportCodeFragment, f.Resource.ImportAlias(), f.Resource.Path))
	}

	if f.WireController {
		if !f.MultiGroup || f.Resource.Group == "" {
			imports = append(imports, fmt.Sprintf(controllerImportCodeFragment, f.Repo))
		} else {
			imports = append(imports, fmt.Sprintf(multiGroupControllerImportCodeFragment,
				f.Resource.PackageName(), f.Repo, f.Resource.Group))
		}
	}

	// Generate add scheme code fragments
	addScheme := make([]string, 0)
	if f.WireResource {
		addScheme = append(addScheme, fmt.Sprintf(addschemeCodeFragment, f.Resource.ImportAlias()))
	}

	// Generate setup code fragments
	setup := make([]string, 0)

	if f.WireController {
		if !f.MultiGroup || f.Resource.Group == "" {
			setup = append(setup, fmt.Sprintf(reconcilerSetupCodeFragment, f.Resource.Kind))
		} else {
			setup = append(
				setup, fmt.Sprintf(
					multiGroupReconcilerSetupCodeFragment,
					f.Resource.PackageName(),
					f.Resource.Kind,
				),
			)
		}
	}

	// Generate webhook wiring code fragments.  These are statements rather than literal
	// elements, so they are kept separate from the reconciler setup fragments above.
	webhookSetup := make([]string, 0)

	if f.WireWebhook {
		alias := f.webhookImportAlias()

		if f.Resource.Group == "" {
			imports = append(imports, fmt.Sprintf(webhookImportCodeFragment,
				alias, f.Repo, f.Resource.Version))
		} else {
			imports = append(imports, fmt.Sprintf(multiGroupWebhookImportCodeFragment,
				alias, f.Repo, f.Resource.Group, f.Resource.Version))
		}

		webhookSetup = append(webhookSetup, fmt.Sprintf(webhookSetupCodeFragment,
			alias, f.Resource.Kind, f.Resource.Kind))
	}

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(defaultMainPath, importMarker)] = imports
	}

	if len(addScheme) != 0 {
		fragments[machinery.NewMarkerFor(defaultMainPath, addSchemeMarker)] = addScheme
	}

	if len(setup) != 0 {
		fragments[machinery.NewMarkerFor(defaultMainPath, setupMarker)] = setup
	}

	if len(webhookSetup) != 0 {
		fragments[machinery.NewMarkerFor(defaultMainPath, webhookMarker)] = webhookSetup
	}

	return fragments
}

// webhookImportAlias returns the import alias for the generated webhook package.  The
// alias is prefixed to keep it distinct from the API package alias, which main.go also
// imports and which resolves to the same trailing path elements.
func (f *MainUpdater) webhookImportAlias() string {
	if f.Resource.Group == "" {
		return "webhook" + f.Resource.Version
	}

	return "webhook" + f.Resource.ImportAlias()
}

//nolint:lll
const mainTemplate = `{{ .Boilerplate }}

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	%s
)

type ReconcilerInitializer interface {
	GetName() string
	SetupWithManager(ctrl.Manager) error
}

var (
	scheme = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	%s
}

func main() {
	// webhook flags
	var webhookCertPath, webhookCertName, webhookCertKey string
	flag.StringVar(&webhookCertPath, "webhook-cert-path", "", "The directory that contains the webhook certificate.")
	flag.StringVar(&webhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	flag.StringVar(&webhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")

	// metrics flags
	var secureMetrics bool
	var metricsAddr string
	var metricsCertPath, metricsCertName, metricsCertKey string
	flag.BoolVar(&secureMetrics, "metrics-secure", false, "If set the metrics endpoint is served securely")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&metricsCertPath, "metrics-cert-path", "", "The directory that contains the metrics server certificate.")
	flag.StringVar(&metricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	flag.StringVar(&metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")

	// health probe flags
	var probeAddr string
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")

	// generic server flags
	var enableLeaderElection bool
	var enableHTTP2 bool
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// only print a given warning the first time we receive it
	rest.SetDefaultWarningHandler(
		rest.NewWarningWriter(os.Stderr, rest.WarningWriterOptions{
			Deduplicate: true,
		}),
	)

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and 
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	// Initial webhook TLS options
	webhookTLSOpts := tlsOpts
	webhookServerOptions := webhook.Options{
		TLSOpts: webhookTLSOpts,
	}

	if len(webhookCertPath) > 0 {
		setupLog.Info("Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path", webhookCertPath, "webhook-cert-name", webhookCertName, "webhook-cert-key", webhookCertKey)

		webhookServerOptions.CertDir = webhookCertPath
		webhookServerOptions.CertName = webhookCertName
		webhookServerOptions.KeyName = webhookCertKey
	}

	webhookServer := webhook.NewServer(webhookServerOptions)

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@{{ .ControllerRuntimeVersion }}/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts: tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@{{ .ControllerRuntimeVersion }}/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "{{ hashFNV .Repo }}.{{ .Domain }}",
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	reconcilers := []ReconcilerInitializer{
		%s
	}

	for _, reconciler := range reconcilers {
		if err = reconciler.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", reconciler.GetName())
			os.Exit(1)
		}
	}

	%s

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
`
