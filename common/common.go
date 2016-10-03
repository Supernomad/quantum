package common

import (
	"encoding/binary"
	"net"
)

const (
	// RealInterfaceNameEnv is the environment variable the the real interface name is stored for reloads.
	RealInterfaceNameEnv = "_QUANTUM_REAL_INTERFACE_NAME"
)

const (
	// MTU - The max size packet to recieve from the TUN device
	MTU = 65400 - HeaderSize - FooterSize - IPHeaderSize

	// HeaderSize - The size of the perpended data
	HeaderSize = 16

	// FooterSize - The size of the appended data
	FooterSize = 16

	// IPHeaderSize - The size of the injected ip header
	IPHeaderSize = 20

	// MaxPacketLength - The maximum packet size to send via the UDP device
	MaxPacketLength = 65536
)

const (
	// IPStart - The ip start position
	IPStart = 0

	// IPEnd - The ip end position
	IPEnd = 4

	// NonceStart - The nonce start position
	NonceStart = 4

	// NonceEnd - The nonce end postion
	NonceEnd = 16

	// PacketStart - The packet start position
	PacketStart = 16
)

// IPtoInt takes a string ip in the form '0.0.0.0' and returns a uint32 that represents that ipaddress
func IPtoInt(IP string) uint32 {
	buf := net.ParseIP(IP).To4()
	return binary.LittleEndian.Uint32(buf)
}

// IncrementIP will increment the given ip in place.
func IncrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}
