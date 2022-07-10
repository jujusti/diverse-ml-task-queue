/*
 * Copyright Morpheo Org. 2017
 *
 * contact@morpheo.co
 *
 * This software is part of the Morpheo project, an open-source machine
 * learning platform.
 *
 * This software is governed by the CeCILL license, compatible with the
 * GNU GPL, under French law and abiding by the rules of distribution of
 * free software. You can  use, modify and/ or redistribute the software
 * under the terms of the CeCILL license as circulated by CEA, CNRS and
 * INRIA at the following URL "http://www.cecill.info".
 *
 * As a counterpart to the access to the source code and  rights to copy,
 * modify and redistribute granted by the license, users are provided only
 * with a limited warranty  and the software's author,  the holder of the
 * economic rights,  and the successive licensors  have only  limited
 * liability.
 *
 * In this respect, the user's attention is drawn to the risks associated
 * with loading,  using,  modifying and/or developing or reproducing the
 * software by the user in light of its specific status of free software,
 * that may mean  that it is complicated to manipulate,  and  that  also
 * therefore means  that it is reserved for developers  and  experienced
 * professionals having in-depth computer knowledge. Users are therefore
 * encouraged to load and test the software's suitability as regards their
 * requirements in conditions enabling the security of their systems and/or
 * data to be ensured and,  more generally, to use and operate it in the
 * same conditions as regards security.
 *
 * The fact that you are presently reading this means that you have had
 * knowledge of the CeCILL license and that you accept its terms.
 */

package main

import (
	"flag"
	"sync"

	"github.com/MorpheoOrg/morpheo-go-packages/common"
)

// ProducerConfig Compute API configuration, subject to dynamic changes for the addresses of
// storage & orchestrator endpoints, and any RESTFul HTTP API added in the future.
type ProducerConfig struct {
	Hostname             string
	Port                 int
	OchestratorEndpoints []string
	StorageEndpoints     []string
	Broker               string
	BrokerHost           string
	BrokerPort           int
	CertFile             string
	KeyFile              string

	lock sync.Mutex
}

// TLSOn returns true if TLS credentials have been provided
func (c *ProducerConfig) TLSOn() bool {
	return c.CertFile != "" && c.KeyFile != ""
}

// Lock locks the config store
func (c *ProducerConfig) Lock() {
	c.lock.Lock()
}

// Unlock unlocks the config store to be written to
func (c *ProducerConfig) Unlock() {
	c.lock.Unlock()
}

// NewProducerConfig computes the configuration object. Note that a pointer is returned not to avoid
// copy but rather to allow the configuration to be dynamically changed.  If this isn't possible
// with a flags or env. variables, we may later make it possible to get the config from a K/V store
// such as etcd or consul to allow dynamic conf updates without requiring a restart.
//
// When using the config, please keep in mind that it can therefore be changed at any time. If you
// don't want this to happen, please use the object's Lock()/Unlock() features.
func NewProducerConfig() (conf *ProducerConfig) {
	var (
		hostname      string
		port          int
		orchestrators common.MultiStringFlag
		storages      common.MultiStringFlag
		broker        string
		brokerHost    string
		brokerPort    int
		certFile      string
		keyFile       string
	)

	// CLI Flags
	flag.StringVar(&hostname, "host",