package integration

import (
	"fmt"

	mgmtv3 "github.com/harvester/managedchart-controller/pkg/apis/management.cattle.io/v3"
	"github.com/harvester/managedchart-controller/pkg/controller/managedchart"
	fleetv1alpha1 "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ensure managed charts is synced to bundle", func() {
	mc1 := mgmtv3.ManagedChart{}

	BeforeEach(func() {
		mc1 = mgmtv3.ManagedChart{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mc1",
				Namespace: "harvester-system",
			},
			Spec: mgmtv3.ManagedChartSpec{
				Chart:       "harvester-cloud-provider",
				Version:     "0.2.4",
				RepoName:    RepoURL,
				ReleaseName: "harvester-cloud-provider",
			},
			Status: mgmtv3.ManagedChartStatus{},
		}
		// Setup code before each test goes here
		// create managedchart object
		err := k8sClient.Create(ctx, &mc1)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		// Cleanup code after each test goes here
		err := k8sClient.Delete(ctx, &mc1)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should create the fleet bundle correctly", func() {
		By("checking if the fleet bundle is created correctly", func() {
			Eventually(func() error {
				bundleObj := &fleetv1alpha1.Bundle{}
				bundleName := managedchart.SafeConcatName(managedchart.MaxHelmReleaseNameLength, "mcc", mc1.Name)
				err := k8sClient.Get(ctx, types.NamespacedName{Name: bundleName, Namespace: mc1.Namespace}, bundleObj)
				return err
			}, "30s", "5s").ShouldNot(HaveOccurred())
		})
	})
})

var _ = Describe("ensure bundle is removed when chart is deleted", func() {
	mc2 := mgmtv3.ManagedChart{}

	BeforeEach(func() {
		mc2 = mgmtv3.ManagedChart{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mc2",
				Namespace: "harvester-system",
			},
			Spec: mgmtv3.ManagedChartSpec{
				Chart:       "harvester-cloud-provider",
				Version:     "0.2.4",
				RepoName:    RepoURL,
				ReleaseName: "harvester-cloud-provider",
			},
			Status: mgmtv3.ManagedChartStatus{},
		}
		// Setup code before each test goes here
		// create managedchart object
		err := k8sClient.Create(ctx, &mc2)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should delete the fleet bundle correctly", func() {
		By("delete managedchart", func() {
			Eventually(func() error {
				return k8sClient.Delete(ctx, &mc2)
			}, "30s", "5s").ShouldNot(HaveOccurred())
		})

		By("checking if the fleet bundle is deleted correctly", func() {
			Eventually(func() error {
				bundleObj := &fleetv1alpha1.Bundle{}
				bundleName := managedchart.SafeConcatName(managedchart.MaxHelmReleaseNameLength, "mcc", mc2.Name)
				err := k8sClient.Get(ctx, types.NamespacedName{Name: bundleName, Namespace: mc2.Namespace}, bundleObj)
				if err != nil {
					if apierrors.IsNotFound(err) {
						return nil
					}
					return err
				}
				return fmt.Errorf("bundle %s still exists", bundleName)
			}, "30s", "5s").ShouldNot(HaveOccurred())
		})
	})
})
