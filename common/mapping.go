package common

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"net"
	"syscall"
)

// Mapping represents the relationship between a public/private address and encryption metadata
type Mapping struct {
	PrivateIP  string
	PublicKey  []byte
	PublicIP   string
	PublicPort int
	Addr       net.IP           `json:"-"`
	Cipher     cipher.AEAD      `json:"-"`
	Sockaddr   syscall.Sockaddr `json:"-"`
}

// Bytes returns the mapping as a byte slice
func (m *Mapping) Bytes() []byte {
	buf, _ := json.Marshal(m)
	return buf
}

// ParseMapping creates a new mapping based on the output of Mapping.Bytes
func ParseMapping(data, privkey []byte) (*Mapping, error) {
	var mapping Mapping
	json.Unmarshal(data, &mapping)

	secret := GenerateSharedSecret(mapping.PublicKey, privkey)

	block, err := aes.NewCipher(secret[0:16])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	mapping.Cipher = aesgcm

	mapping.Addr = net.ParseIP(mapping.PublicIP)

	if addr := mapping.Addr.To4(); addr != nil {
		sa := &syscall.SockaddrInet4{Port: mapping.PublicPort}
		copy(sa.Addr[:], addr[:])
		mapping.Sockaddr = sa
	} else if addr := mapping.Addr.To16(); addr != nil {
		sa := &syscall.SockaddrInet6{Port: mapping.PublicPort}
		copy(sa.Addr[:], addr[:])
		mapping.Sockaddr = sa
	}

	return &mapping, nil
}

// NewMapping generates a new basic Mapping
func NewMapping(privateIP, publicIP string, publicPort int, pubkey []byte) *Mapping {
	return &Mapping{
		PublicIP:   publicIP,
		PublicPort: publicPort,
		PrivateIP:  privateIP,
		PublicKey:  pubkey,
	}
}
