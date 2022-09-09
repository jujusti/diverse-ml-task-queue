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
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/middleware/logger"

	"github.com/MorpheoOrg/morpheo-go-packages/client"
	"github.com/MorpheoOrg/morpheo-go-packages/common"
)

// TODO: write tests for the two main views

// Available HTTP Routes
const (
	RootRoute   = "/"
	HealthRoute = "/health"
)

type apiServer struct {
	conf     *ProducerConfig
	producer common.Producer
	peer     client.Peer
}

func (s *apiServer) configureRoutes(app *iris.Framework) {
	app.Get(RootRoute, s.index)
	app.Get(HealthRoute, s.health)
	app.Get("/query", s.query)   // For test purposes
	app.Get("/invoke", s.invoke) // For test purposes
}

// SetIrisApp sets the base for the Iris App
func (s *apiServer) SetIrisApp() *iris.Framework {
	// Iris setup
	app := iris.New()
	app.Adapt(iris.DevLogger())
	app.Adapt(httprouter.New())

	// Logging middleware configuration
	customLogger := logger.New(logger.Config{
		Status: true,
		IP:     true,
		Method: true,
		Path:   true,
	})
	app.Use(customLogger)

	s.configureRoutes(app)
	return app
}

func main() {
	// App-specific config (parses CLI flags)
	conf := NewProducerConfig()

	// Let's dependency inject the producer for the chosen Broker
	var producer common.Producer
	switch conf.Broker {
	case common.BrokerNSQ:
		var err error
		producer, err = common.NewNSQProducer(conf.BrokerHost, conf.BrokerPort)
		defer producer.Stop()
		if err != nil {
			log.Panicln(err)
		}
	case common.BrokerMOCK:
		producer = &common.ProducerMOCK{}
	default:
		log.Panicf("Unsupported broker (%s). Available brokers: 'nsq', 'mock'", conf.Broker)
	}

	// Let's create our peer client to request the blockchain
	// TODO: WITH ADMIN/USER ID INSTEAD
	peer, err := client.NewPeerAPI(
		"secrets/config.yaml",
		"Aphp",
		"mychannel"