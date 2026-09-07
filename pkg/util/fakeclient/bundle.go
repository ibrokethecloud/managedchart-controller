package fakeclient

import (
	"context"

	v1alpha1 "github.com/harvester/managedchart-controller/pkg/generated/clientset/versioned/typed/fleet.cattle.io/v1alpha1"
	fleetv1alpha1 "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/wrangler/v3/pkg/generic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type BundleCache func() v1alpha1.BundleInterface

func (m BundleCache) Get(name, namespace string) (*fleetv1alpha1.Bundle, error) {
	return m().Get(context.TODO(), name, metav1.GetOptions{})
}

func (m BundleCache) List(namespace string, selector labels.Selector) ([]*fleetv1alpha1.Bundle, error) {
	list, err := m().List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	result := make([]*fleetv1alpha1.Bundle, 0, len(list.Items))
	for _, node := range list.Items {
		obj := node
		result = append(result, &obj)
	}
	return result, err
}

func (m BundleCache) AddIndexer(_ string, _ generic.Indexer[*fleetv1alpha1.Bundle]) {
	panic("implement me")
}

func (m BundleCache) GetByIndex(indexName, key string) ([]*fleetv1alpha1.Bundle, error) {
	panic("implement me")
}
