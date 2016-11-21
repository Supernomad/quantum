package common

import (
	"encoding/json"
	"net"
	"time"
)

// DefaultNetworkConfig to use when the NetworkConfig is not specified in the backend datastore.
var DefaultNetworkConfig *NetworkConfig

// NetworkConfig object to represent the current network.
type NetworkConfig struct {
	Network   string        `json:"network"`
	LeaseTime time.Duration `json:"leaseTime"`
	BaseIP    net.IP        `json:"-"`
	IPNet     *net.IPNet    `json:"-"`
}

// ParseNetworkConfig from the return of the backend datastore
func ParseNetworkConfig(data []byte) (*NetworkConfig, error) {
	var networkCfg NetworkConfig
	json.Unmarshal(data, &networkCfg)

	baseIP, ipnet, err := net.ParseCIDR(networkCfg.Network)
	if err != nil {
		return nil, err
	}

	networkCfg.BaseIP = baseIP
	networkCfg.IPNet = ipnet

	return &networkCfg, nil
}

// Bytes representation of a NetworkConfig object
func (networkCfg *NetworkConfig) Bytes() []byte {
	buf, _ := json.Marshal(networkCfg)
	return buf
}

// String representation of a NetworkConfig object
func (networkCfg *NetworkConfig) String() string {
	return string(networkCfg.Bytes())
}

func init() {
	defaultLeaseTime, _ := time.ParseDuration("48h")
	DefaultNetworkConfig = &NetworkConfig{
		Network:   "10.10.0.0/16",
		LeaseTime: defaultLeaseTime,
	}

	baseIP, ipnet, _ := net.ParseCIDR(DefaultNetworkConfig.Network)
	DefaultNetworkConfig.BaseIP = baseIP
	DefaultNetworkConfig.IPNet = ipnet
}
