// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package datastore

import (
	"errors"
	"time"

	"github.com/supernomad/quantum/common"
)

// Type represents the datastore backend to use for synchronizing mapping objects over the quantum network.
type Type int

const (
	// ETCDDatastore will tell quantum to use etcd as the backend datastore.
	ETCDDatastore Type = iota

	// MOCKDatastore will tell quantum to use a moked out backend datastore for testing.
	MOCKDatastore

	lockTTL = 10 * time.Second
)

// Datastore interface for quantum to use for retrieving mapping data from the backend datastore.
type Datastore interface {
	// Init should handle setting up the datastore connections, and initializing the mappings/local mapping.
	Init() error

	// Mapping should return the mapping and true if it exists, if not the mapping should be nil and false should be returned along with it.
	Mapping(ip uint32) (*common.Mapping, bool)

	// GatewayMapping should retun the mapping and true if it exists specifically for destinations outside of the quantum network, if the mapping doesn't exist it will return nil and false.
	GatewayMapping(ip uint32) (*common.Mapping, bool)

	// Start should kick off any routines that need to run in the background to groom the mappings and manage the datastore state.
	Start()

	// Stop should fully shutdown all operation and ensure that all connections are terminated gracefully.
	Stop()
}

// New generates a datastore object based on the passed in Type and user configuration.
func New(datastoreType Type, cfg *common.Config) (Datastore, error) {
	switch datastoreType {
	case ETCDDatastore:
		return newEtcd(cfg)
	case MOCKDatastore:
		return newMock(cfg)
	default:
		return nil, errors.New("specified backend doesn't exist")
	}
}
