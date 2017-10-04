package hostgw

import (
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	winroute "github.com/llparse/win-route"
	"github.com/rancher/go-rancher-metadata/metadata"
)

const (
	ProviderName = "hostgw"

	changeCheckInterval = 5
	perHostSubnetLabel  = "io.rancher.network.per_host_subnet.subnet"
)

type HostGw struct {
	m metadata.Client
	r winroute.NetRoute
}

func New(m metadata.Client) (*HostGw, error) {
	o := &HostGw{
		m: m,
		r: winroute.NewNetRoute(),
	}
	return o, nil
}

func (p *HostGw) Start() {
	go p.m.OnChange(changeCheckInterval, p.onChangeNoError)
}

// TODO p.r.Close() when a stop signal is received

func (p *HostGw) onChangeNoError(version string) {
	if err := p.Reload(); err != nil {
		log.Errorf("Failed to apply host route : %v", err)
	}
}

func (p *HostGw) Reload() error {
	log.Debug("HostGW: reload")
	if err := p.configure(); err != nil {
		return errors.Wrap(err, "Failed to reload hostgw routes")
	}
	return nil
}

func (p *HostGw) configure() error {
	log.Debug("HostGW: reload")

	selfHost, err := p.m.GetSelfHost()
	if err != nil {
		return errors.Wrap(err, "Failed to get self host from metadata")
	}
	allHosts, err := p.m.GetHosts()
	if err != nil {
		return errors.Wrap(err, "Failed to get all hosts from metadata")
	}

	currentRoutes, err := p.getCurrentRouteEntries(selfHost)
	if err != nil {
		return errors.Wrap(err, "Failed to getCurrentRouteEntries")
	}
	desiredRoutes, err := p.getDesiredRouteEntries(selfHost, allHosts)
	if err != nil {
		return errors.Wrap(err, "Failed to getDesiredRouteEntries")
	}
	err = p.updateRoutes(currentRoutes, desiredRoutes)
	if err != nil {
		return errors.Wrap(err, "Failed to updateRoutes")
	}
	return err
}

func (p *HostGw) getCurrentRouteEntries(host metadata.Host) (map[string]winroute.IPForwardRow, error) {
	i := winroute.MustResolveInterface(net.ParseIP(host.AgentIP))

	intf, err := p.r.GetInterfaceByIndex(uint32(i.Index))
	if err != nil {
		return nil, err
	}

	routes, err := p.r.GetRoutes()
	if err != nil {
		return nil, err
	}

	routeEntries := make(map[string]winroute.IPForwardRow)
	for _, route := range routes {
		// Ignore routes on other interfaces
		if intf.InterfaceIndex != route.ForwardIfIndex {
			continue
		}

		gwIP := winroute.Inet_ntoa(route.ForwardNextHop, false)
		routeEntries[gwIP] = route
	}
	log.Debugf("%+v", routeEntries)
	return routeEntries, nil
}

func (p *HostGw) getDesiredRouteEntries(selfHost metadata.Host, allHosts []metadata.Host) (map[string]winroute.IPForwardRow, error) {
	return make(map[string]winroute.IPForwardRow), nil
}

func (p *HostGw) updateRoutes(oldEntries map[string]winroute.IPForwardRow, newEntries map[string]winroute.IPForwardRow) error {
	return nil
}
