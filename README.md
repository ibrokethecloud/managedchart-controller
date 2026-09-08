# managedchart-controller

Standalone controller to manage lifecycle of existing managedcharts

The controller is only to be used within Harvester and not as a standalone project.

It is built with assumption that `managedchart.Spec.RepoName` points to the internal harvester cluster repo svc.

The controller only supports helm v3 repos, and fetches and downloads the charts and updates the associated bundles
