package hostgw

import (
	"net"

	log "github.com/Sirupsen/logrus"
	winroute "github.com/llparse/win-route"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher-metadata/metadata"
)

const (
	ProviderName = "hostgw"

	changeCheckInterval = 5
	perHostSubnetLabel  = "io.rancher.network.per_host_subnet.subnet"
)

type HostGw struct {
	m metadata.Client
	r *winroute.NetRoute
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

	i := winroute.MustResolveInterface(net.ParseIP(selfHost.AgentIP))
	iface, err := p.r.GetInterfaceByIndex(uint32(i.Index))
	if err != nil {
		return err
	}

	currentRoutes, err := p.getCurrentRouteEntries(iface, selfHost)
	if err != nil {
		return errors.Wrap(err, "Failed to getCurrentRouteEntries")
	}
	desiredRoutes, err := p.getDesiredRouteEntries(iface, selfHost, allHosts)
	if err != nil {
		return errors.Wrap(err, "Failed to getDesiredRouteEntries")
	}
	err = p.updateRoutes(currentRoutes, desiredRoutes)
	if err != nil {
		return errors.Wrap(err, "Failed to updateRoutes")
	}
	return err
}

func (p *HostGw) getCurrentRouteEntries(iface *winroute.IPInterfaceEntry, host metadata.Host) (map[string]winroute.IPForwardRow, error) {
	routes, err := p.r.GetRoutes()
	if err != nil {
		return nil, err
	}

	routeEntries := make(map[string]winroute.IPForwardRow)
	for _, route := range routes {
		// Skip routes on other interfaces
		if iface.InterfaceIndex != route.ForwardIfIndex {
			continue
		}
		// Skip link-local routes
		if winroute.Inet_aton(host.AgentIP, false) == route.ForwardNextHop {
			continue
		}

		gwIP := winroute.Inet_ntoa(route.ForwardNextHop, false)
		routeEntries[gwIP] = route
	}

	p.logRouteEntries(routeEntries, "getCurrentRouteEntries")
	return routeEntries, nil
}

func (p *HostGw) getDesiredRouteEntries(iface *winroute.IPInterfaceEntry, selfHost metadata.Host, allHosts []metadata.Host) (map[string]winroute.IPForwardRow, error) {
	routeEntries := make(map[string]winroute.IPForwardRow)

	for _, h := range allHosts {
		// Link-local routes already in place
		if h.UUID == selfHost.UUID {
			continue
		}

		dest, mask, err := ParseIPNet(h)
		if err != nil {
			return nil, err
		}
		r := winroute.IPForwardRow{
			ForwardDest:    winroute.Inet_aton(dest.String(), false),
			ForwardMask:    winroute.Inet_aton(mask.String(), false),
			ForwardNextHop: winroute.Inet_aton(h.AgentIP, false),
			ForwardIfIndex: iface.InterfaceIndex,
			ForwardType:    3,
			ForwardProto:   3,
			ForwardMetric1: iface.Metric, // route metric is 0 (+ interface metric)
		}
		routeEntries[h.AgentIP] = r
	}

	p.logRouteEntries(routeEntries, "getDesiredRouteEntries")
	return routeEntries, nil
}

func (p *HostGw) updateRoutes(oldEntries map[string]winroute.IPForwardRow, newEntries map[string]winroute.IPForwardRow) error {
	return nil
}

func ParseIPNet(host metadata.Host) (net.IP, net.IP, error) {
	_, ipNet, err := net.ParseCIDR(host.Labels[perHostSubnetLabel])
	if err != nil {
		return nil, nil, err
	}
	return ipNet.IP, net.IP(ipNet.Mask), nil
}

func (p *HostGw) logRouteEntries(entries map[string]winroute.IPForwardRow, action string) {
	if log.GetLevel() == log.DebugLevel {
		for _, route := range entries {
			log.WithFields(log.Fields{
				"dest":    winroute.Inet_ntoa(route.ForwardDest, false),
				"mask":    winroute.Inet_ntoa(route.ForwardMask, false),
				"gateway": winroute.Inet_ntoa(route.ForwardNextHop, false),
			}).Debug(action)
		}
	}
}
