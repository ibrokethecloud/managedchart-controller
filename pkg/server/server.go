package server

import (
	"context"
	"os"

	mgmtv3 "github.com/harvester/managedchart-controller/pkg/apis/management.cattle.io/v3"
	"github.com/harvester/managedchart-controller/pkg/controller/managedchart"
	fleetfactory "github.com/harvester/managedchart-controller/pkg/generated/controllers/fleet.cattle.io"
	mgmtfactory "github.com/harvester/managedchart-controller/pkg/generated/controllers/management.cattle.io"
	fleetv1alpha1 "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/lasso/pkg/cache"
	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/wrangler/v3/pkg/apply"
	"github.com/rancher/wrangler/v3/pkg/leader"
	"github.com/rancher/wrangler/v3/pkg/schemes"
	"github.com/rancher/wrangler/v3/pkg/start"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

var (
	localSchemeBuilder = runtime.SchemeBuilder{
		v1beta1.AddToScheme,
	}
	AddToScheme = localSchemeBuilder.AddToScheme
	Scheme      = runtime.NewScheme()
)

func init() {
	utilruntime.Must(AddToScheme(Scheme))
	utilruntime.Must(schemes.AddToScheme(Scheme))
	utilruntime.Must(apiregistrationv1.AddToScheme(Scheme))
	utilruntime.Must(mgmtv3.AddToScheme(Scheme))
	utilruntime.Must(fleetv1alpha1.AddToScheme(Scheme))
	if debug := os.Getenv("DEBUG_LOGGING"); debug == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func Setup(ctx context.Context, restConfig *rest.Config) error {

	clientFactory, err := client.NewSharedClientFactory(restConfig, nil)
	if err != nil {
		return err
	}

	k8sclient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	cacheFactory := cache.NewSharedCachedFactory(clientFactory, nil)

	mgmtFactory, err := mgmtfactory.NewFactoryFromConfigWithOptions(restConfig, &mgmtfactory.FactoryOptions{SharedCacheFactory: cacheFactory})
	if err != nil {
		return err
	}

	bundleFactory, err := fleetfactory.NewFactoryFromConfigWithOptions(restConfig, &fleetfactory.FactoryOptions{SharedCacheFactory: cacheFactory})
	if err != nil {
		return err
	}

	applyInterface, err := apply.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	// leader election and start caches
	leader.RunOrDie(ctx, "harvester-system", "managedchart-controller", k8sclient, func(ctx context.Context) {
		// register managedChart controller
		managedchart.Register(ctx, mgmtFactory.Management().V3(), bundleFactory.Fleet().V1alpha1().Bundle(), applyInterface)
	})

	return start.All(ctx, 2, bundleFactory, mgmtFactory)
}
