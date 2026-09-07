package main

import (
	"os"

	v3 "github.com/harvester/managedchart-controller/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
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
			"fleet.cattle.io": {
				Types: []interface{}{
					v1alpha1.Bundle{},
				},
				GenerateTypes:   false,
				GenerateClients: true,
			},
		},
	})
}
