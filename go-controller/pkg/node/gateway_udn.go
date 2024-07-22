package node

import (
	"fmt"
	"net"
	"strings"

	v1 "k8s.io/api/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"

	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/config"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/generator/udn"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/kube"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/node/vrfmanager"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/types"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/util"
	"github.com/vishvananda/netlink"
)

const (
	// ctMarkUDNBase is the conntrack mark base value for user defined networks to use
	// Each network gets its own mark == base + network-id
	ctMarkUDNBase = 3
)

// UserDefinedNetworkGateway contains information
// required to program a UDN at each node's
// gateway.
// NOTE: Currently invoked only for primary networks.
type UserDefinedNetworkGateway struct {
	// network information
	util.NetInfo
	// stores the networkID of this network
	networkID int
	// node that its programming things on
	node          *v1.Node
	nodeLister    listers.NodeLister
	kubeInterface kube.Interface
	// vrf manager that creates and manages vrfs for all UDNs
	// used with a lock since its shared between all network controllers
	vrfManager *vrfmanager.Controller
	// masqCTMark holds the mark value for this network
	// which is used for egress traffic in shared gateway mode
	masqCTMark uint
	// v4MasqIP holds the IPv4 masquerade IP for this network
	v4MasqIP *net.IPNet
	// v6MasqIP holds the IPv6 masquerade IP for this network
	v6MasqIP *net.IPNet
}

// UTILS Needed for UDN (also leveraged for default netInfo) in bridgeConfiguration

// getBridgePortConfigurations returns a slice of Network port configurations along with the
// uplinkName and physical port's ofport value
func (b *bridgeConfiguration) getBridgePortConfigurations() ([]bridgeUDNConfiguration, string, string) {
	b.Lock()
	defer b.Unlock()
	netConfigs := make([]bridgeUDNConfiguration, len(b.netConfig))
	for _, netConfig := range b.netConfig {
		netConfigs = append(netConfigs, *netConfig)
	}
	return netConfigs, b.uplinkName, b.ofPortPhys
}

// addNetworkBridgeConfig adds the patchport and ctMark value for the provided netInfo into the bridge configuration cache
func (b *bridgeConfiguration) addNetworkBridgeConfig(nInfo util.NetInfo, masqCTMark uint, v4MasqIP, v6MasqIP *net.IPNet) {
	b.Lock()
	defer b.Unlock()

	netName := nInfo.GetNetworkName()
	patchPort := nInfo.GetNetworkScopedPatchPortName(b.bridgeName, b.nodeName)

	_, found := b.netConfig[netName]
	if !found {
		netConfig := &bridgeUDNConfiguration{
			patchPort:  patchPort,
			masqCTMark: fmt.Sprintf("0x%x", masqCTMark),
			v4MasqIP:   v4MasqIP,
			v6MasqIP:   v6MasqIP,
		}

		b.netConfig[netName] = netConfig
	} else {
		klog.Warningf("Trying to update bridge config for network %s which already"+
			"exists in cache...networks are not mutable...ignoring update", nInfo.GetNetworkName())
	}
}

// delNetworkBridgeConfig deletes the provided netInfo from the bridge configuration cache
func (b *bridgeConfiguration) delNetworkBridgeConfig(nInfo util.NetInfo) {
	b.Lock()
	defer b.Unlock()

	delete(b.netConfig, nInfo.GetNetworkName())
}

func (b *bridgeConfiguration) patchedNetConfigs() []*bridgeUDNConfiguration {
	result := make([]*bridgeUDNConfiguration, 0, len(b.netConfig))
	for _, netConfig := range b.netConfig {
		if netConfig.ofPortPatch == "" {
			continue
		}
		result = append(result, netConfig)
	}
	return result
}

// END UDN UTILs for bridgeConfiguration

// bridgeUDNConfiguration holds the patchport and ctMark
// information for a given network
type bridgeUDNConfiguration struct {
	patchPort   string
	ofPortPatch string
	masqCTMark  string
	v4MasqIP    *net.IPNet
	v6MasqIP    *net.IPNet
}

func (netConfig *bridgeUDNConfiguration) setBridgeNetworkOfPortsInternal() error {
	ofportPatch, stderr, err := util.GetOVSOfPort("get", "Interface", netConfig.patchPort, "ofport")
	if err != nil {
		return fmt.Errorf("failed while waiting on patch port %q to be created by ovn-controller and "+
			"while getting ofport. stderr: %q, error: %v", netConfig.patchPort, stderr, err)
	}
	netConfig.ofPortPatch = ofportPatch
	return nil
}

func setBridgeNetworkOfPorts(bridge *bridgeConfiguration, netName string) error {
	bridge.Lock()
	defer bridge.Unlock()

	netConfig, found := bridge.netConfig[netName]
	if !found {
		return fmt.Errorf("failed to find network %s configuration on bridge %s", netName, bridge.bridgeName)
	}
	return netConfig.setBridgeNetworkOfPortsInternal()
}

func NewUserDefinedNetworkGateway(netInfo util.NetInfo, networkID int, node *v1.Node, nodeLister listers.NodeLister,
	kubeInterface kube.Interface, vrfManager *vrfmanager.Controller) (*UserDefinedNetworkGateway, error) {
	// Generate a per network conntrack mark and masquerade IPs to be used for egress traffic.
	var (
		v4MasqIP *net.IPNet
		v6MasqIP *net.IPNet
	)
	masqCTMark := ctMarkUDNBase + uint(networkID)
	if config.IPv4Mode {
		v4MasqIPs, err := udn.AllocateV4MasqueradeIPs(networkID)
		if err != nil {
			return nil, fmt.Errorf("failed to get v4 masquerade IP, network %s (%d): %v", netInfo.GetNetworkName(), networkID, err)
		}
		v4MasqIP = v4MasqIPs.GatewayRouter
	}
	if config.IPv6Mode {
		v6MasqIPs, err := udn.AllocateV6MasqueradeIPs(networkID)
		if err != nil {
			return nil, fmt.Errorf("failed to get v6 masquerade IP, network %s (%d): %v", netInfo.GetNetworkName(), networkID, err)
		}
		v6MasqIP = v6MasqIPs.GatewayRouter
	}
	return &UserDefinedNetworkGateway{
		NetInfo:       netInfo,
		networkID:     networkID,
		node:          node,
		nodeLister:    nodeLister,
		kubeInterface: kubeInterface,
		vrfManager:    vrfManager,
		masqCTMark:    masqCTMark,
		v4MasqIP:      v4MasqIP,
		v6MasqIP:      v6MasqIP,
	}, nil
}

// AddNetwork will be responsible to create all plumbings
// required by this UDN on the gateway side
func (udng *UserDefinedNetworkGateway) AddNetwork() error {
	mplink, err := udng.addUDNManagementPort()
	if err != nil {
		return fmt.Errorf("could not create management port netdevice for network %s: %w", udng.GetNetworkName(), err)
	}
	vrfDeviceName := util.GetVRFDeviceNameForUDN(udng.networkID)
	vrfTableId := util.CalculateRouteTableID(mplink.Attrs().Index)
	err = udng.vrfManager.AddVRF(vrfDeviceName, mplink.Attrs().Name, uint32(vrfTableId))
	if err != nil {
		return fmt.Errorf("could not add VRF %d for network %s, err: %v", vrfTableId, udng.GetNetworkName(), err)
	}
	return nil
}

// DelNetwork will be responsible to remove all plumbings
// used by this UDN on the gateway side
func (udng *UserDefinedNetworkGateway) DelNetwork() error {
	vrfDeviceName := util.GetVRFDeviceNameForUDN(udng.networkID)
	err := udng.vrfManager.DeleteVRF(vrfDeviceName)
	if err != nil {
		return err
	}
	return udng.deleteUDNManagementPort()
}

// addUDNManagementPort does the following:
// STEP1: creates the (netdevice) OVS interface on br-int for the UDN's management port
// STEP2: It saves the MAC address generated on the 1st go as an option on the OVS interface
// so that it persists on reboots
// STEP3: sets up the management port link on the host
// STEP4: adds the management port IP .2 to the mplink
// STEP5: adds the mac address to the node management port annotation
func (udng *UserDefinedNetworkGateway) addUDNManagementPort() (netlink.Link, error) {
	var err error
	interfaceName := util.GetNetworkScopedK8sMgmtHostIntfName(uint(udng.networkID))
	var networkLocalSubnets []*net.IPNet
	if udng.TopologyType() == types.Layer3Topology {
		networkLocalSubnets, err = util.ParseNodeHostSubnetAnnotation(udng.node, udng.GetNetworkName())
		if err != nil {
			return nil, fmt.Errorf("waiting for node %s to start, no annotation found on node for network %s: %w",
				udng.node.Name, udng.GetNetworkName(), err)
		}
	} else if udng.TopologyType() == types.Layer2Topology {
		// NOTE: We don't support L2 networks without subnets as primary UDNs
		globalFlatL2Networks := udng.Subnets()
		for _, globalFlatL2Network := range globalFlatL2Networks {
			networkLocalSubnets = append(networkLocalSubnets, globalFlatL2Network.CIDR)
		}
	}

	// STEP1
	stdout, stderr, err := util.RunOVSVsctl(
		"--", "--may-exist", "add-port", "br-int", interfaceName,
		"--", "set", "interface", interfaceName,
		"type=internal", "mtu_request="+fmt.Sprintf("%d", udng.NetInfo.MTU()),
		"external-ids:iface-id="+udng.GetNetworkScopedK8sMgmtIntfName(udng.node.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add port to br-int for network %s, stdout: %q, stderr: %q, error: %w",
			udng.GetNetworkName(), stdout, stderr, err)
	}
	klog.V(3).Infof("Added OVS management port interface %s for network %s", interfaceName, udng.GetNetworkName())

	// STEP2
	macAddress, err := util.GetOVSPortMACAddress(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get management port MAC address for network %s: %v", udng.GetNetworkName(), err)
	}
	// persist the MAC address so that upon node reboot we get back the same mac address.
	_, stderr, err = util.RunOVSVsctl("set", "interface", interfaceName,
		fmt.Sprintf("mac=%s", strings.ReplaceAll(macAddress.String(), ":", "\\:")))
	if err != nil {
		return nil, fmt.Errorf("failed to persist MAC address %q for %q while plumbing network %s: stderr:%s (%v)",
			macAddress.String(), interfaceName, udng.GetNetworkName(), stderr, err)
	}

	// STEP3
	mplink, err := util.LinkSetUp(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to set the link up for interface %s while plumbing network %s, err: %v",
			interfaceName, udng.GetNetworkName(), err)
	}
	klog.V(3).Infof("Setup management port link %s for network %s succeeded", interfaceName, udng.GetNetworkName())

	// STEP4
	for _, subnet := range networkLocalSubnets {
		if config.IPv6Mode && utilnet.IsIPv6CIDR(subnet) || config.IPv4Mode && utilnet.IsIPv4CIDR(subnet) {
			ip := util.GetNodeManagementIfAddr(subnet)
			var err error
			var exists bool
			if exists, err = util.LinkAddrExist(mplink, ip); err == nil && !exists {
				err = util.LinkAddrAdd(mplink, ip, 0, 0, 0)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to add management port IP from subnet %s to netdevice %s for network %s, err: %v",
					subnet, interfaceName, udng.GetNetworkName(), err)
			}
		}
	}

	// STEP5
	if err := util.UpdateNodeManagementPortMACAddressesWithRetry(udng.node, udng.nodeLister, udng.kubeInterface, macAddress, udng.GetNetworkName()); err != nil {
		return nil, fmt.Errorf("unable to update mac address annotation for node %s, for network %s, err: %v", udng.node.Name, udng.GetNetworkName(), err)
	}
	klog.V(3).Infof("Added management port mac address information of %s for network %s", interfaceName, udng.GetNetworkName())
	return mplink, nil
}

// deleteUDNManagementPort does the following:
// STEP1: deletes the OVS interface on br-int for the UDN's management port interface
// STEP2: deletes the mac address from the annotation
func (udng *UserDefinedNetworkGateway) deleteUDNManagementPort() error {
	var err error
	interfaceName := util.GetNetworkScopedK8sMgmtHostIntfName(uint(udng.networkID))
	// STEP1
	stdout, stderr, err := util.RunOVSVsctl(
		"--", "--if-exists", "del-port", "br-int", interfaceName,
	)
	if err != nil {
		return fmt.Errorf("failed to delete port from br-int for network %s, stdout: %q, stderr: %q, error: %v",
			udng.GetNetworkName(), stdout, stderr, err)
	}
	klog.V(3).Infof("Removed OVS management port interface %s for network %s", interfaceName, udng.GetNetworkName())
	// sending nil mac address will delete the network's annotation value
	if err := util.UpdateNodeManagementPortMACAddressesWithRetry(udng.node, udng.nodeLister, udng.kubeInterface, nil, udng.GetNetworkName()); err != nil {
		return fmt.Errorf("unable to remove mac address annotation for node %s, for network %s, err: %v", udng.node.Name, udng.GetNetworkName(), err)
	}
	klog.V(3).Infof("Removed management port mac address information of %s for network %s", interfaceName, udng.GetNetworkName())
	return nil
}
