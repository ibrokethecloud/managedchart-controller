package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	log "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	mgmtv3 "github.com/harvester/managedchart-controller/pkg/apis/management.cattle.io/v3"
	"github.com/harvester/managedchart-controller/pkg/controller/managedchart"
	"github.com/harvester/managedchart-controller/pkg/crd"
	fleetfactory "github.com/harvester/managedchart-controller/pkg/generated/controllers/fleet.cattle.io"
	mgmtfactory "github.com/harvester/managedchart-controller/pkg/generated/controllers/management.cattle.io"
	fleetv1alpha1 "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/lasso/pkg/cache"
	lassoClient "github.com/rancher/lasso/pkg/client"
	"github.com/rancher/wrangler/v3/pkg/apply"
	"github.com/rancher/wrangler/v3/pkg/start"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	scheme    = runtime.NewScheme()
	ctx       context.Context
	cancel    context.CancelFunc
	defaultNS = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "harvester-system",
		},
	}
	// RepoURL is populated by the http test and used for subsequent tests
	RepoURL    = ""
	httpServer *httptest.Server
)

// Declarations for Ginkgo DSL
var Fail = ginkgo.Fail
var Describe = ginkgo.Describe
var It = ginkgo.It
var By = ginkgo.By
var BeforeEach = ginkgo.BeforeEach
var AfterEach = ginkgo.AfterEach
var BeforeSuite = ginkgo.BeforeSuite
var AfterSuite = ginkgo.AfterSuite
var RunSpecs = ginkgo.RunSpecs
var GinkgoWriter = ginkgo.GinkgoWriter

// Declarations for Gomega Matchers
var RegisterFailHandler = gomega.RegisterFailHandler
var Equal = gomega.Equal
var Expect = gomega.Expect
var BeNil = gomega.BeNil
var HaveOccurred = gomega.HaveOccurred
var BeEmpty = gomega.BeEmpty
var Eventually = gomega.Eventually

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t,
		"Controller Suite",
	)
}

var _ = BeforeSuite(func() {
	var err error
	log.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.TODO())
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDInstallOptions: envtest.CRDInstallOptions{
			Paths: []string{filepath.Join("..", "manifests", "fleet-crd")},
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = apiregistrationv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = mgmtv3.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = fleetv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = crd.Create(ctx, cfg)
	Expect(err).NotTo(HaveOccurred())

	handler := http.FileServer(http.Dir("../../pkg/helm/testdata"))
	httpServer = httptest.NewServer(handler)
	RepoURL = httpServer.URL

	// setup harvester-system namespace
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	err = k8sClient.Create(ctx, defaultNS)
	Expect(err).NotTo(HaveOccurred())

	clientFactory, err := lassoClient.NewSharedClientFactory(cfg, nil)
	Expect(err).NotTo(HaveOccurred())

	cacheFactory := cache.NewSharedCachedFactory(clientFactory, nil)

	mgmtFactory, err := mgmtfactory.NewFactoryFromConfigWithOptions(cfg, &mgmtfactory.FactoryOptions{SharedCacheFactory: cacheFactory})
	Expect(err).NotTo(HaveOccurred())

	bundleFactory, err := fleetfactory.NewFactoryFromConfigWithOptions(cfg, &fleetfactory.FactoryOptions{SharedCacheFactory: cacheFactory})
	Expect(err).NotTo(HaveOccurred())

	applyInterface, err := apply.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	managedchart.Register(ctx, mgmtFactory.Management().V3(), bundleFactory.Fleet().V1alpha1().Bundle(), applyInterface)

	err = start.All(ctx, 2, bundleFactory, mgmtFactory)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
	httpServer.Close()
})
