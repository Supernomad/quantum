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
	// IPStart - The ip start position
	IPStart = 0

	// IPEnd - The ip end position
	IPEnd = 4

	// IPLength - The length of the private ip header
	IPLength = 4

	// NonceStart - The nonce start position
	NonceStart = 4

	// NonceEnd - The nonce end postion
	NonceEnd = 16

	// NonceLength - The nonce length
	NonceLength = 12

	// TagLength - The crypto tag length
	TagLength = 16

	// PacketStart - The packet start position
	PacketStart = 16

	// MaxPacketLength - The maximum packet size to send via the UDP device
	MaxPacketLength = 65500

	// HeaderSize - The size of the perpended data
	HeaderSize = IPLength + NonceLength

	// FooterSize - The size of the appended data
	FooterSize = TagLength

	// MTU - The max size packet to recieve from the TUN device
	MTU = MaxPacketLength - HeaderSize - FooterSize
)

const ()

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
