package conf

import (
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/xtls/xray-core/proxy/wireguard"
	"google.golang.org/protobuf/proto"
)

type WireGuardPeerConfig struct {
	PublicKey    string   `json:"publicKey,omitempty"`
	PreSharedKey string   `json:"preSharedKey,omitempty"`
	Endpoint     string   `json:"endpoint,omitempty"`
	KeepAlive    uint32   `json:"keepAlive,omitempty"`
	AllowedIPs   []string `json:"allowedIPs,omitempty"`
}

func (c *WireGuardPeerConfig) Build() (proto.Message, error) {
	var err error
	config := new(wireguard.PeerConfig)

	if c.PublicKey != "" {
		config.PublicKey, err = parseWireGuardKey(c.PublicKey)
		if err != nil {
			return nil, err
		}
	}

	if c.PreSharedKey != "" {
		config.PreSharedKey, err = parseWireGuardKey(c.PreSharedKey)
		if err != nil {
			return nil, err
		}
	}

	config.Endpoint = c.Endpoint
	// default 0
	config.KeepAlive = c.KeepAlive
	if c.AllowedIPs == nil {
		config.AllowedIps = []string{"0.0.0.0/0", "::0/0"}
	} else {
		config.AllowedIps = c.AllowedIPs
	}

	return config, nil
}

type WireGuardConfig struct {
	IsClient bool `json:",omitempty"`

	KernelMode     *bool                  `json:"kernelMode,omitempty"`
	SecretKey      string                 `json:"secretKey,omitempty"`
	Address        []string               `json:"address,omitempty"`
	Peers          []*WireGuardPeerConfig `json:"peers,omitempty"`
	MTU            int32                  `json:"mtu,omitempty"`
	NumWorkers     int32                  `json:"workers,omitempty"`
	Reserved       []byte                 `json:"reserved,omitempty"`
	DomainStrategy string                 `json:"domainStrategy,omitempty"`
}

func (c *WireGuardConfig) Build() (proto.Message, error) {
	config := new(wireguard.DeviceConfig)

	var err error
	config.SecretKey, err = parseWireGuardKey(c.SecretKey)
	if err != nil {
		return nil, err
	}

	if c.Address == nil {
		// bogon ips
		config.Endpoint = []string{"10.0.0.1", "fd59:7153:2388:b5fd:0000:0000:0000:0001"}
	} else {
		config.Endpoint = c.Address
	}

	if c.Peers != nil {
		config.Peers = make([]*wireguard.PeerConfig, len(c.Peers))
		for i, p := range c.Peers {
			msg, err := p.Build()
			if err != nil {
				return nil, err
			}
			config.Peers[i] = msg.(*wireguard.PeerConfig)
		}
	}

	if c.MTU == 0 {
		config.Mtu = 1420
	} else {
		config.Mtu = c.MTU
	}
	// these a fallback code exists in wireguard-go code,
	// we don't need to process fallback manually
	config.NumWorkers = c.NumWorkers

	if len(c.Reserved) != 0 && len(c.Reserved) != 3 {
		return nil, newError(`"reserved" should be empty or 3 bytes`)
	}
	config.Reserved = c.Reserved

	switch strings.ToLower(c.DomainStrategy) {
	case "forceip", "":
		config.DomainStrategy = wireguard.DeviceConfig_FORCE_IP
	case "forceipv4":
		config.DomainStrategy = wireguard.DeviceConfig_FORCE_IP4
	case "forceipv6":
		config.DomainStrategy = wireguard.DeviceConfig_FORCE_IP6
	case "forceipv4v6":
		config.DomainStrategy = wireguard.DeviceConfig_FORCE_IP46
	case "forceipv6v4":
		config.DomainStrategy = wireguard.DeviceConfig_FORCE_IP64
	default:
		return nil, newError("unsupported domain strategy: ", c.DomainStrategy)
	}

	config.IsClient = c.IsClient
	if c.KernelMode != nil {
		config.KernelMode = *c.KernelMode
		if config.KernelMode && !wireguard.KernelTunSupported() {
			newError("kernel mode is not supported on your OS or permission is insufficient").AtWarning().WriteToLog()
		}
	} else {
		config.KernelMode = wireguard.KernelTunSupported()
		if config.KernelMode {
			newError("kernel mode is enabled as it's supported and permission is sufficient").AtDebug().WriteToLog()
		}
	}

	return config, nil
}

func parseWireGuardKey(str string) (string, error) {
	var err error

	if len(str)%2 == 0 {
		_, err = hex.DecodeString(str)
		if err == nil {
			return str, nil
		}
	}

	var dat []byte
	str = strings.TrimSuffix(str, "=")
	if strings.ContainsRune(str, '+') || strings.ContainsRune(str, '/') {
		dat, err = base64.RawStdEncoding.DecodeString(str)
	} else {
		dat, err = base64.RawURLEncoding.DecodeString(str)
	}
	if err == nil {
		return hex.EncodeToString(dat), nil
	}

	return "", newError("failed to deserialize key").Base(err)
}
