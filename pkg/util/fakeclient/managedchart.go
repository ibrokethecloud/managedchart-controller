package fakeclient

import (
	"context"

	mgmtv3api "github.com/harvester/managedchart-controller/pkg/apis/management.cattle.io/v3"
	mgmtv3client "github.com/harvester/managedchart-controller/pkg/generated/clientset/versioned/typed/management.cattle.io/v3"
	"github.com/rancher/wrangler/v3/pkg/generic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type ManagedChartCache func(string) mgmtv3client.ManagedChartInterface

func (m ManagedChartCache) Get(namespace, name string) (*mgmtv3api.ManagedChart, error) {
	return m(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (m ManagedChartCache) List(namespace string, selector labels.Selector) ([]*mgmtv3api.ManagedChart, error) {
	list, err := m(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	result := make([]*mgmtv3api.ManagedChart, 0, len(list.Items))
	for _, node := range list.Items {
		obj := node
		result = append(result, &obj)
	}
	return result, err
}

func (m ManagedChartCache) AddIndexer(_ string, _ generic.Indexer[*mgmtv3api.ManagedChart]) {
	panic("implement me")
}

func (m ManagedChartCache) GetByIndex(indexName, key string) ([]*mgmtv3api.ManagedChart, error) {
	panic("implement me")
}

type ManagedChartClient func(string) mgmtv3client.ManagedChartInterface

func (m ManagedChartClient) Update(managedChart *mgmtv3api.ManagedChart) (*mgmtv3api.ManagedChart, error) {
	return m(managedChart.Namespace).Update(context.TODO(), managedChart, metav1.UpdateOptions{})
}

func (m ManagedChartClient) Get(namespace, name string, options metav1.GetOptions) (*mgmtv3api.ManagedChart, error) {
	return m(namespace).Get(context.TODO(), name, options)
}

func (m ManagedChartClient) Create(managedChart *mgmtv3api.ManagedChart) (*mgmtv3api.ManagedChart, error) {
	return m(managedChart.Namespace).Create(context.TODO(), managedChart, metav1.CreateOptions{})
}

func (m ManagedChartClient) Delete(_, _ string, _ *metav1.DeleteOptions) error {
	panic("implement me")
}

func (m ManagedChartClient) List(_ string, _ metav1.ListOptions) (*mgmtv3api.ManagedChartList, error) {
	panic("implement me")
}

func (m ManagedChartClient) UpdateStatus(managedChart *mgmtv3api.ManagedChart) (*mgmtv3api.ManagedChartStatus, error) {
	panic("implement me")
}

func (m ManagedChartClient) Watch(_ string, _ metav1.ListOptions) (watch.Interface, error) {
	panic("implement me")
}

func (m ManagedChartClient) Patch(_, _ string, _ types.PatchType, _ []byte, _ ...string) (result *mgmtv3api.ManagedChart, err error) {
	panic("implement me")
}

func (m ManagedChartClient) WithImpersonation(_ rest.ImpersonationConfig) (generic.ClientInterface[*mgmtv3api.ManagedChart, *mgmtv3api.ManagedChartList], error) {
	panic("implement me")
}

func (m ManagedChartClient) DeleteCollection(_ context.Context, _ metav1.DeleteOptions, _ metav1.List) error {
	panic("implement me")
}
