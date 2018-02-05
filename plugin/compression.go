// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package plugin

import (
	"github.com/golang/snappy"
	"github.com/supernomad/quantum/common"
)

// Compression plugin struct to use for compressing outgoing packets or decompressing incoming packets.
type Compression struct {
	cfg *common.Config
}

func compress(raw []byte) ([]byte, int) {
	buf := snappy.Encode(nil, raw)
	return buf, len(buf)
}

func decompress(raw []byte) ([]byte, int) {
	buf, err := snappy.Decode(nil, raw)
	if err != nil {
		return nil, -1
	}
	return buf, len(buf)
}

// Apply returns the payload/mapping compressed if the direction is Outgoing and decompressed if the direction is Incoming.
func (comp *Compression) Apply(direction Direction, payload *common.Payload, mapping *common.Mapping) (*common.Payload, *common.Mapping, bool) {
	if !common.StringInSlice(CompressionPlugin, mapping.SupportedPlugins) {
		return payload, mapping, true
	}

	switch direction {
	case Incoming:
		decompressed, length := decompress(payload.Packet)
		if decompressed == nil {
			return payload, mapping, false
		}

		copy(payload.Raw[common.PacketStart:], decompressed)
		payload.Packet = payload.Raw[common.PacketStart : common.PacketStart+length]
		payload.Length = common.HeaderSize + length
	case Outgoing:
		compressed, length := compress(payload.Packet)
		if compressed == nil {
			return payload, mapping, false
		}

		copy(payload.Raw[common.PacketStart:], compressed)
		payload.Packet = payload.Raw[common.PacketStart : common.PacketStart+length]
		payload.Length = common.HeaderSize + length
	}
	return payload, mapping, true
}

// Close which is a noop.
func (comp *Compression) Close() error {
	return nil
}

// Name returns 'compression'.
func (comp *Compression) Name() string {
	return CompressionPlugin
}

// Order returns the CompressionPluginOrder value.
func (comp *Compression) Order() int {
	return CompressionPluginOrder
}

func newCompression(cfg *common.Config) (Plugin, error) {
	return &Compression{
		cfg: cfg,
	}, nil
}
