package hostgw

import (
	"github.com/rancher/go-rancher-metadata/metadata"
)

func getCurrentRouteEntries(host metadata.Host) (map[string]string, error) {
	return make(map[string]string), nil
}

func getDesiredRouteEntries(selfHost metadata.Host, allHosts []metadata.Host) (map[string]string, error) {
	return make(map[string]string), nil
}

func updateRoutes(oldEntries map[string]string, newEntries map[string]string) error {
	return nil
}
