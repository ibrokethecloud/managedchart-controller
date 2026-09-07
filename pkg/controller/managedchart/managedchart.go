package managedchart

import (
	tar2 "archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"sort"
	"strings"

	v3 "github.com/harvester/managedchart-controller/pkg/apis/management.cattle.io/v3"
	fleetcontrollers "github.com/harvester/managedchart-controller/pkg/generated/controllers/fleet.cattle.io/v1alpha1"
	mgmtcontrollers "github.com/harvester/managedchart-controller/pkg/generated/controllers/management.cattle.io/v3"
	"github.com/harvester/managedchart-controller/pkg/helm"
	"github.com/pkg/errors"
	"github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/wrangler/v3/pkg/apply"
	"github.com/rancher/wrangler/v3/pkg/name"
	"github.com/rancher/wrangler/v3/pkg/relatedresource"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	MaxHelmReleaseNameLength = 53
)

func Register(ctx context.Context, mgmtController mgmtcontrollers.Interface, bundleController fleetcontrollers.BundleController, apply apply.Apply) {
	h := &Handler{
		bundleCache: bundleController.Cache(),
	}

	relatedresource.Watch(ctx,
		"mcc-from-bundle-trigger",
		relatedresource.OwnerResolver(true, v3.SchemeGroupVersion.String(), "ManagedChart"),
		mgmtController.ManagedChart(),
		bundleController)
	mgmtcontrollers.RegisterManagedChartGeneratingHandler(ctx,
		mgmtController.ManagedChart(),
		apply.
			WithSetOwnerReference(true, true).
			WithCacheTypes(
				mgmtController.ManagedChart(),
				bundleController),
		"Defined",
		"mcc-bundle",
		h.OnChange,
		nil)
}

type Handler struct {
	bundleCache fleetcontrollers.BundleCache
}

func (h *Handler) OnChange(mcc *v3.ManagedChart, status v3.ManagedChartStatus) ([]runtime.Object, v3.ManagedChartStatus, error) {
	chart, err := helm.ReadChart(mcc.Spec.RepoName, mcc.Spec.Chart, mcc.Spec.Version)
	if err != nil {
		return nil, status, err
	}

	defer func() {
		if err := chart.Close(); err != nil {
			logrus.Errorf("failed to close chart reader: %v", err)
		}
	}()

	bundle := &v1alpha1.Bundle{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SafeConcatName(MaxHelmReleaseNameLength, "mcc", mcc.Name),
			Namespace: mcc.Namespace,
		},
		Spec: v1alpha1.BundleSpec{
			BundleDeploymentOptions: v1alpha1.BundleDeploymentOptions{
				DefaultNamespace: mcc.Spec.DefaultNamespace,
				TargetNamespace:  mcc.Spec.TargetNamespace,
				Helm: &v1alpha1.HelmOptions{
					ReleaseName:    name.Limit(mcc.Spec.ReleaseName, MaxHelmReleaseNameLength),
					Version:        mcc.Spec.Version,
					TimeoutSeconds: mcc.Spec.TimeoutSeconds,
					Values:         mcc.Spec.Values,
					Force:          mcc.Spec.Force,
					TakeOwnership:  mcc.Spec.TakeOwnership,
					MaxHistory:     mcc.Spec.MaxHistory,
				},
				ServiceAccount: mcc.Spec.ServiceAccount,
				Diff:           mcc.Spec.Diff,
			},
			Paused:          mcc.Spec.Paused,
			RolloutStrategy: mcc.Spec.RolloutStrategy,
			Targets:         mcc.Spec.Targets,
		},
	}

	gz, err := gzip.NewReader(chart)
	if err != nil {
		return nil, status, err
	}

	tar := tar2.NewReader(gz)
	for {
		next, err := tar.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, status, err
		}

		if next.Typeflag != tar2.TypeReg {
			continue
		}

		buf := &bytes.Buffer{}
		gz := gzip.NewWriter(buf)
		if _, err := io.Copy(gz, tar); err != nil {
			return nil, status, err
		}
		if err := gz.Close(); err != nil {
			return nil, status, err
		}

		name := strings.TrimPrefix(next.Name, mcc.Spec.Chart+"/")
		bundle.Spec.Resources = append(bundle.Spec.Resources, v1alpha1.BundleResource{
			Name:     name,
			Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
			Encoding: "base64+gz",
		})
	}

	sort.Slice(bundle.Spec.Resources, func(i, j int) bool {
		return bundle.Spec.Resources[i].Name < bundle.Spec.Resources[j].Name
	})

	status, err = h.updateStatus(status, bundle)
	return []runtime.Object{
		bundle,
	}, status, err
}

func (h *Handler) updateStatus(status v3.ManagedChartStatus, bundle *v1alpha1.Bundle) (v3.ManagedChartStatus, error) {
	bundle, err := h.bundleCache.Get(bundle.Namespace, bundle.Name)
	if apierrors.IsNotFound(err) {
		return status, nil
	} else if err != nil {
		return status, err
	}

	status.BundleStatus = bundle.Status
	return status, nil
}

func SafeConcatName(maxLength int, name ...string) string {

	hashLength := 6

	fullPath := strings.Join(name, "-")
	if len(fullPath) <= maxLength {
		return fullPath
	}

	if maxLength == 0 {
		return ""
	}

	if maxLength <= 5 {
		return fullPath[:maxLength]
	}

	digest := sha256.Sum256([]byte(fullPath))

	// since we trailingCharacterIndex the string in the middle, the last char may not be compatible with what is expected in k8s
	// we are checking and if necessary removing the last char
	trailingCharacterIndex := maxLength - (hashLength + 1)
	if trailingCharacterIndex < 0 {
		trailingCharacterIndex = 0
	}
	c := fullPath[trailingCharacterIndex]

	if 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		remainingString := fullPath[0 : maxLength-(hashLength)]
		hash := hex.EncodeToString(digest[0:])[0 : hashLength-1]
		if remainingString == "" {
			// if we've completely converted the input into a hash don't append '-'
			return hash
		}
		return remainingString + "-" + hash
	}

	return fullPath[0:maxLength-(hashLength+1)] + "-" + hex.EncodeToString(digest[0:])[0:hashLength]
}
