/*
SPDX-License-Identifier: Apache-2.0

Copyright Contributors to the Submariner project.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fake

import (
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/submariner-io/admiral/pkg/slices"
	netlinkAPI "github.com/submariner-io/submariner/pkg/netlink"
	"github.com/vishvananda/netlink"
)

type linkType struct {
	netlink.Link
	isSetup bool
}

type basicType struct {
	mutex        sync.Mutex
	linkIndices  map[string]int
	links        map[string]*linkType
	routes       map[int][]netlink.Route
	neighbors    map[int][]netlink.Neigh
	rules        map[int][]netlink.Rule
	addrs        map[int][]netlink.Addr
	addrUpdateCh atomic.Value
}

type NetLink struct {
	netlinkAPI.Adapter
}

var _ netlinkAPI.Interface = &NetLink{}

type linkNotFoundError struct{}

func (e linkNotFoundError) Error() string {
	return "Link not found"
}

func (e linkNotFoundError) Is(err error) bool {
	_, ok := err.(netlink.LinkNotFoundError)
	return ok
}

func New() *NetLink {
	return &NetLink{
		Adapter: netlinkAPI.Adapter{Basic: &basicType{
			linkIndices: map[string]int{},
			links:       map[string]*linkType{},
			routes:      map[int][]netlink.Route{},
			neighbors:   map[int][]netlink.Neigh{},
			rules:       map[int][]netlink.Rule{},
			addrs:       map[int][]netlink.Addr{},
		}},
	}
}

func (n *NetLink) basic() *basicType {
	return n.Adapter.Basic.(*basicType)
}

func (n *NetLink) SetLinkIndex(name string, index int) {
	n.basic().mutex.Lock()
	defer n.basic().mutex.Unlock()

	n.basic().linkIndices[name] = index
}

func (n *basicType) LinkAdd(link netlink.Link) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if _, found := n.links[link.Attrs().Name]; found {
		return syscall.EEXIST
	}

	link.Attrs().Index = n.linkIndices[link.Attrs().Name]
	n.links[link.Attrs().Name] = &linkType{Link: link}

	return nil
}

func (n *basicType) LinkDel(link netlink.Link) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	delete(n.links, link.Attrs().Name)

	return nil
}

func (n *basicType) LinkByName(name string) (netlink.Link, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	link, found := n.links[name]
	if !found {
		return nil, linkNotFoundError{}
	}

	return link.Link, nil
}

func (n *basicType) LinkSetUp(link netlink.Link) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	l, found := n.links[link.Attrs().Name]
	if !found {
		return linkNotFoundError{}
	}

	l.isSetup = true

	return nil
}

func (n *basicType) sendAddrUpdate(linkIndex int, addr *netlink.Addr, added bool) {
	updateCh := n.addrUpdateCh.Load()
	if updateCh == nil {
		return
	}

	updateCh.(chan netlink.AddrUpdate) <- netlink.AddrUpdate{
		LinkAddress: *addr.IPNet,
		LinkIndex:   linkIndex,
		Flags:       addr.Flags,
		Scope:       addr.Scope,
		PreferedLft: addr.PreferedLft,
		ValidLft:    addr.ValidLft,
		NewAddr:     added,
	}
}

func (n *basicType) AddrAdd(link netlink.Link, addr *netlink.Addr) error {
	n.mutex.Lock()

	var added bool

	n.addrs[link.Attrs().Index], added = slices.AppendIfNotPresent(n.addrs[link.Attrs().Index], *addr, func(a netlink.Addr) string {
		return a.IPNet.String()
	})

	n.mutex.Unlock()

	if !added {
		return syscall.EEXIST
	}

	n.sendAddrUpdate(link.Attrs().Index, addr, true)

	return nil
}

func (n *basicType) AddrDel(link netlink.Link, addr *netlink.Addr) error {
	n.mutex.Lock()

	var removed bool

	n.addrs[link.Attrs().Index], removed = slices.Remove(n.addrs[link.Attrs().Index], *addr, func(a netlink.Addr) string {
		return a.IPNet.String()
	})

	n.mutex.Unlock()

	if !removed {
		return syscall.ENOENT
	}

	n.sendAddrUpdate(link.Attrs().Index, addr, false)

	return nil
}

func (n *basicType) AddrList(link netlink.Link, _ int) ([]netlink.Addr, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	return n.addrs[link.Attrs().Index], nil
}

func (n *basicType) AddrSubscribe(updateCh chan netlink.AddrUpdate, _ chan struct{}) error {
	n.addrUpdateCh.Store(updateCh)
	return nil
}

func (n *basicType) NeighAppend(neigh *netlink.Neigh) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.neighbors[neigh.LinkIndex] = append(n.neighbors[neigh.LinkIndex], *neigh)

	return nil
}

func (n *basicType) NeighDel(neigh *netlink.Neigh) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	neighbors := n.neighbors[neigh.LinkIndex]
	for i := range neighbors {
		if reflect.DeepEqual(neighbors[i], *neigh) {
			n.neighbors[neigh.LinkIndex] = append(neighbors[:i], neighbors[i+1:]...)
			break
		}
	}

	return nil
}

//nolint:gocritic // Ignore hugeParam.
func routeKey(r netlink.Route) string {
	k := ""
	if r.Gw != nil {
		k = r.Gw.String()
	}

	if r.Dst != nil {
		k += r.Dst.String()
	}

	if r.Src != nil {
		k += r.Src.String()
	}

	k += strconv.Itoa(r.Table)

	return k
}

func (n *basicType) RouteAdd(route *netlink.Route) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	var added bool

	n.routes[route.LinkIndex], added = slices.AppendIfNotPresent(n.routes[route.LinkIndex], *route, routeKey)
	if !added {
		return syscall.EEXIST
	}

	return nil
}

func (n *basicType) RouteDel(route *netlink.Route) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	routes := n.routes[route.LinkIndex]

	for i := range routes {
		if reflect.DeepEqual(routes[i].Dst, route.Dst) {
			n.routes[route.LinkIndex] = append(routes[:i], routes[i+1:]...)
			break
		}
	}

	return nil
}

func (n *basicType) RouteReplace(route *netlink.Route) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	index := slices.IndexOf(n.routes[route.LinkIndex], routeKey(*route), routeKey)
	if index >= 0 {
		n.routes[route.LinkIndex][index] = *route
	} else {
		n.routes[route.LinkIndex] = append(n.routes[route.LinkIndex], *route)
	}

	return nil
}

func (n *basicType) FlushRouteTable(table int) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	for index, routes := range n.routes {
		newRoutes := []netlink.Route{}

		for i := range routes {
			if routes[i].Table != table {
				newRoutes = append(newRoutes, routes[i])
			}
		}

		n.routes[index] = newRoutes
	}

	return nil
}

func (n *basicType) RouteGet(destination net.IP) ([]netlink.Route, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	routes := []netlink.Route{}
	for i := range n.routes {
		for j := range n.routes[i] {
			if n.routes[i][j].Dst != nil && reflect.DeepEqual(n.routes[i][j].Dst.IP, destination) {
				routes = append(routes, n.routes[i][j])
			}
		}
	}

	return routes, nil
}

func (n *basicType) RouteList(link netlink.Link, _ int) ([]netlink.Route, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	r := n.routes[link.Attrs().Index]
	to := make([]netlink.Route, len(r))
	copy(to, r)

	return to, nil
}

//nolint:gocritic // Ignore hugeParam.
func ruleKey(r netlink.Rule) string {
	k := ""
	if r.Src != nil {
		k = r.Src.String()
	}

	if r.Dst != nil {
		k += r.Dst.String()
	}

	return k
}

func (n *basicType) RuleAdd(rule *netlink.Rule) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	family := func(ip net.IP) int {
		if ip.To4() != nil {
			return netlink.FAMILY_V4
		} else if ip.To16() != nil {
			return netlink.FAMILY_V6
		}

		return 0
	}

	r := *rule
	if r.Src != nil {
		r.Family = family(r.Src.IP)
	} else if r.Dst != nil {
		r.Family = family(r.Dst.IP)
	}

	var added bool

	n.rules[rule.Table], added = slices.AppendIfNotPresent(n.rules[rule.Table], r, ruleKey)
	if !added {
		return os.ErrExist
	}

	return nil
}

func (n *basicType) RuleDel(rule *netlink.Rule) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	var removed bool

	n.rules[rule.Table], removed = slices.Remove(n.rules[rule.Table], *rule, ruleKey)
	if !removed {
		return os.ErrNotExist
	}

	return nil
}

func (n *basicType) RuleList(family int) ([]netlink.Rule, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	var rules []netlink.Rule
	for _, r := range n.rules {
		for i := range r {
			if r[i].Family == family {
				rules = append(rules, r[i])
			}
		}
	}

	return rules, nil
}

func (n *basicType) XfrmPolicyAdd(_ *netlink.XfrmPolicy) error {
	return nil
}

func (n *basicType) XfrmPolicyDel(_ *netlink.XfrmPolicy) error {
	return nil
}

func (n *basicType) XfrmPolicyList(_ int) ([]netlink.XfrmPolicy, error) {
	return []netlink.XfrmPolicy{}, nil
}

func (n *basicType) EnableLooseModeReversePathFilter(_ string) error {
	return nil
}

func (n *basicType) EnsureLooseModeIsConfigured(_ string) error {
	return nil
}

func (n *basicType) EnableForwarding(_ string) error {
	return nil
}

func (n *basicType) GetReversePathFilter(_ string) ([]byte, error) {
	return []byte("2"), nil
}

func (n *basicType) ConfigureTCPMTUProbe(_, _ string) error {
	return nil
}

func (n *NetLink) AwaitLink(name string) netlink.Link {
	var link netlink.Link

	Eventually(func() netlink.Link {
		link, _ = n.LinkByName(name)
		return link
	}, 5).ShouldNot(BeNil(), "Link %q not found", name)

	return link
}

func (n *NetLink) AwaitNoLink(name string) {
	Eventually(func() bool {
		_, err := n.LinkByName(name)
		return errors.Is(err, netlink.LinkNotFoundError{})
	}, 5).Should(BeTrue(), "Link %q exists", name)
}

func (n *NetLink) AwaitLinkSetup(name string) {
	Eventually(func() bool {
		n.basic().mutex.Lock()
		defer n.basic().mutex.Unlock()

		link, found := n.basic().links[name]
		if found {
			return link.isSetup
		}

		return false
	}, 5).Should(BeTrue(), "Link %q not setup", name)
}

func routeList[T any](n *NetLink, linkIndex, table int, f func(r *netlink.Route) *T) []T {
	routes, _ := n.RouteList(&netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{
			Index: linkIndex,
		},
	}, 0)

	result := []T{}

	for i := range routes {
		if routes[i].Table != table {
			continue
		}

		r := f(&routes[i])
		if r != nil {
			result = append(result, *r)
		}
	}

	return result
}

func (n *NetLink) routeDestList(linkIndex, table int) []net.IPNet {
	return routeList(n, linkIndex, table, func(r *netlink.Route) *net.IPNet {
		if r.Dst != nil {
			return r.Dst
		}

		return nil
	})
}

func (n *NetLink) AwaitDstRoutes(linkIndex, table int, destCIDRs ...string) {
	for _, destCIDR := range destCIDRs {
		_, expDest, _ := net.ParseCIDR(destCIDR)

		Eventually(func() []net.IPNet {
			return n.routeDestList(linkIndex, table)
		}, 5).Should(ContainElement(*expDest), "Route for %q not found", destCIDR)
	}
}

func (n *NetLink) AwaitNoDstRoutes(linkIndex, table int, destCIDRs ...string) {
	for _, destCIDR := range destCIDRs {
		_, expDest, _ := net.ParseCIDR(destCIDR)

		Eventually(func() []net.IPNet {
			return n.routeDestList(linkIndex, table)
		}, 5).ShouldNot(ContainElement(*expDest), "Route for %q exists", destCIDR)
	}
}

func (n *NetLink) routeGwList(linkIndex, table int) []net.IP {
	return routeList(n, linkIndex, table, func(r *netlink.Route) *net.IP {
		return &r.Gw
	})
}

func (n *NetLink) AwaitGwRoutes(linkIndex, table int, gwIPs ...string) {
	for _, ip := range gwIPs {
		Eventually(func() []net.IP {
			return n.routeGwList(linkIndex, table)
		}, 5).Should(ContainElement(net.ParseIP(ip)), "Route for %q not found", ip)
	}
}

func (n *NetLink) AwaitNoGwRoutes(linkIndex, table int, gwIPs ...string) {
	for _, ip := range gwIPs {
		Eventually(func() []net.IP {
			return n.routeGwList(linkIndex, table)
		}, 5).ShouldNot(ContainElement(net.ParseIP(ip)), "Route for %q exists", ip)
	}
}

func (n *NetLink) neighborIPList(linkIndex int) []net.IP {
	n.basic().mutex.Lock()
	defer n.basic().mutex.Unlock()

	ips := []net.IP{}
	for i := range n.basic().neighbors[linkIndex] {
		ips = append(ips, n.basic().neighbors[linkIndex][i].IP)
	}

	return ips
}

func (n *NetLink) AwaitNeighbors(linkIndex int, expIPs ...string) {
	for _, ip := range expIPs {
		expIP := net.ParseIP(ip)

		Eventually(func() []net.IP {
			return n.neighborIPList(linkIndex)
		}, 5).Should(ContainElement(expIP), "Neighbor for %q not found", expIP)
	}
}

func (n *NetLink) AwaitNoNeighbors(linkIndex int, expIPs ...string) {
	for _, ip := range expIPs {
		expIP := net.ParseIP(ip)

		Eventually(func() []net.IP {
			return n.neighborIPList(linkIndex)
		}, 5).ShouldNot(ContainElement(expIP), "Neighbor for %q exists", expIP)
	}
}

func (n *NetLink) getRule(table int, src, dst string) *netlink.Rule {
	n.basic().mutex.Lock()
	defer n.basic().mutex.Unlock()

	parse := func(s string) *net.IPNet {
		if s == "" {
			return nil
		}

		_, cidr, err := net.ParseCIDR(s)
		Expect(err).To(Succeed())

		return cidr
	}

	r := netlink.NewRule()
	r.Src = parse(src)
	r.Dst = parse(dst)

	i := slices.IndexOf(n.basic().rules[table], ruleKey(*r), ruleKey)
	if i < 0 {
		return nil
	}

	return &n.basic().rules[table][i]
}

func (n *NetLink) AwaitRule(table int, src, dst string) {
	Eventually(func() *netlink.Rule {
		return n.getRule(table, src, dst)
	}, 5).ShouldNot(BeNil(), "Rule for %v not found", table)
}

func (n *NetLink) AwaitNoRule(table int, src, dst string) {
	Eventually(func() *netlink.Rule {
		return n.getRule(table, src, dst)
	}, 5).Should(BeNil(), "Rule for %v exists", table)
}
