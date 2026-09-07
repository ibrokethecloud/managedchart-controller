package helm

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/repo"
)

func DownloadIndex(repoURL string) (*repo.IndexFile, error) {
	client := &http.Client{}
	defer client.CloseIdleConnections()

	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return nil, err
	}

	parsedURL.RawPath = path.Join(parsedURL.RawPath, "index.yaml")
	parsedURL.Path = path.Join(parsedURL.Path, "index.yaml")

	url := parsedURL.String()
	logrus.Infof("Downloading repo index from %s", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("failed to close response body: %v", err)
		}
	}()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Marshall to file to ensure it matches the schema and this component doesn't just
	// become a "fetch any file" service.
	index := &repo.IndexFile{}
	if err := yaml.Unmarshal(bytes, index); err != nil {
		logrus.Errorf("failed to unmarshal %s: %v", url, err)
		return nil, fmt.Errorf("failed to parse response from %s", url)
	}

	if index.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}

	return index, nil
}

func ReadChart(repoURL, chart, version string) (io.ReadCloser, error) {
	index, err := DownloadIndex(repoURL)
	if err != nil {
		return nil, err
	}

	chartEntry, ok := index.Entries[chart]
	if !ok {
		return nil, fmt.Errorf("chart %s not found in index", chart)
	}
	var chartVersion *repo.ChartVersion
	for _, v := range chartEntry {
		if v.Version == version {
			chartVersion = v
			break
		}
	}
	if chartVersion == nil {
		return nil, fmt.Errorf("version %s of chart %s not found in index", version, chart)
	}
	chartURL := chartVersion.URLs[0]
	client := http.Client{}
	defer client.CloseIdleConnections()

	u, err := url.Parse(chartURL)
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		base, err := url.Parse(repoURL)
		if err != nil {
			return nil, err
		}
		// Prevent ResolveReference from stripping the last element
		// of the path by ensuring it has a trailing slash
		base.Path = strings.TrimSuffix(base.Path, "/") + "/"
		u = base.ResolveReference(u)
		// Retain the query string of the repository URL as it might
		// contain an access credential.
		u.RawQuery = base.RawQuery
	}

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("failed to close response body: %v", err)
		}
	}()

	data, err := io.ReadAll(resp.Body)
	return io.NopCloser(bytes.NewBuffer(data)), err
}
