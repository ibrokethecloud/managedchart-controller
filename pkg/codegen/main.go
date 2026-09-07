package main

import (
	"os"

	v1 "github.com/harvester/managedchart-controller/pkg/apis/catalog.cattle.io/v1"
	v3 "github.com/harvester/managedchart-controller/pkg/apis/management.cattle.io/v3"
	controllergen "github.com/rancher/wrangler/v3/pkg/controller-gen"
	"github.com/rancher/wrangler/v3/pkg/controller-gen/args"
)

func main() {
	os.Unsetenv("GOPATH")
	controllergen.Run(args.Options{
		OutputPackage: "github.com/harvester/managedchart-controller/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"management.cattle.io": {
				Types: []interface{}{
					v3.ManagedChart{},
				},
				GenerateTypes:   true,
				GenerateClients: true,
			},
			"catalog.cattle.io": {
				Types: []interface{}{
					v1.ClusterRepo{},
				},
				GenerateTypes:   true,
				GenerateClients: true,
			},
		},
	})
}
