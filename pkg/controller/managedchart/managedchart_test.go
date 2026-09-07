package managedchart

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	mgmtv3 "github.com/harvester/managedchart-controller/pkg/apis/management.cattle.io/v3"
	"github.com/harvester/managedchart-controller/pkg/generated/clientset/versioned/fake"
	"github.com/harvester/managedchart-controller/pkg/util/fakeclient"
	fleetv1alpha1 "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ManagedChartBundleGeneration(t *testing.T) {
	assert := require.New(t)
	handler := http.FileServer(http.Dir("../../helm/testdata"))
	server := httptest.NewServer(handler)

	mc := &mgmtv3.ManagedChart{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managed-chart",
			Namespace: "default",
		},
		Spec: mgmtv3.ManagedChartSpec{
			RepoName: server.URL,
			Chart:    "harvester-cloud-provider",
			Version:  "0.2.14",
		},
		Status: mgmtv3.ManagedChartStatus{},
	}

	// initialise clients to verify reconcile
	client := fake.NewSimpleClientset(mc)
	bundleCache := fakeclient.BundleCache(client.FleetV1alpha1().Bundles)

	h := &Handler{
		bundleCache: bundleCache,
	}

	bundle, status, err := h.OnChange(mc, mc.Status)
	assert.NoError(err)
	assert.Len(bundle, 1, "expected to find 1 object")
	_, ok := bundle[0].(*fleetv1alpha1.Bundle)
	assert.True(ok, "expected bundle[0] to be of type *fleetv1alpha1.Bundle")
	assert.NotNil(status)
}
