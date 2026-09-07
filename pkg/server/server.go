package server

import (
	"sync"

	catalogv1 "github.com/harvester/managedchart-controller/pkg/generated/controllers/catalog.cattle.io/v1"
	fleetv1alpha1 "github.com/harvester/managedchart-controller/pkg/generated/controllers/fleet.cattle.io/v1alpha1"
	mgmtv3 "github.com/harvester/managedchart-controller/pkg/generated/controllers/management.cattle.io/v3"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/lasso/pkg/dynamic"
	"github.com/rancher/rancher/pkg/catalogv2/content"
	"github.com/rancher/rancher/pkg/catalogv2/helmop"
	"github.com/rancher/rancher/pkg/catalogv2/system"
	"github.com/rancher/rancher/pkg/namespace"
	"github.com/rancher/steve/pkg/accesscontrol"
	"github.com/rancher/wrangler/v3/pkg/apply"
	"github.com/rancher/wrangler/v3/pkg/leader"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Context struct {
	RESTConfig *rest.Config

	Apply   apply.Apply
	Dynamic *dynamic.Controller
	Mgmt    mgmtv3.Interface
	Fleet   fleetv1alpha1.Interface

	Catalog           catalogv1.Interface
	ControllerFactory controller.SharedControllerFactory

	K8s *namespace.Clientset

	ASL                     accesscontrol.AccessSetLookup
	ClientConfig            clientcmd.ClientConfig
	CachedDiscovery         discovery.CachedDiscoveryInterface
	RESTMapper              meta.RESTMapper
	SharedControllerFactory controller.SharedControllerFactory
	leadership              *leader.Manager
	controllerLock          *sync.Mutex

	RESTClientGetter      genericclioptions.RESTClientGetter
	CatalogContentManager *content.Manager
	HelmOperations        *helmop.Operations
	SystemChartsManager   *system.Manager

	started bool
}
