
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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	// "io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/satori/go.uuid"

	"github.com/MorpheoOrg/morpheo-go-packages/client"
	"github.com/MorpheoOrg/morpheo-go-packages/common"
)

// Worker describes a worker (where it stores its data, which container runtime it uses...).
// Most importantly, it carefully implements all the steps of our learning/testing/prediction
// workflow.
//
// For an in-detail understanding of what these different steps do and how, check out Camille's
// awesome example: https://github.com/MorpheoOrg/hypnogram-wf
// The doc also gets there in detail: https://morpheoorg.github.io/morpheo/modules/learning.html
type Worker struct {
	ID uuid.UUID
	// Worker configuration variables
	dataFolder           string
	trainFolder          string
	testFolder           string
	untargetedTestFolder string
	modelFolder          string
	predFolder           string
	perfFolder           string
	problemImagePrefix   string
	algoImagePrefix      string

	// ContainerRuntime abstractions
	containerRuntime common.ContainerRuntime

	// Morpheo API clients
	storage client.Storage
	peer    client.Peer
}

// Perfuplet describes the performance.json file, an output of learning tasks
type Perfuplet struct {
	Perf      float64            `json:"perf"`
	TrainPerf map[string]float64 `json:"train_perf"`
	TestPerf  map[string]float64 `json:"test_perf"`
}

// NewWorker creates a Worker instance
func NewWorker(dataFolder, trainFolder, testFolder, untargetedTestFolder, predFolder, perfFolder, modelFolder, problemImagePrefix, algoImagePrefix string, containerRuntime common.ContainerRuntime, storage client.Storage, peer client.Peer) *Worker {
	return &Worker{
		ID: uuid.NewV4(),

		dataFolder:           dataFolder,
		trainFolder:          trainFolder,
		testFolder:           testFolder,
		predFolder:           predFolder,
		perfFolder:           perfFolder,
		untargetedTestFolder: untargetedTestFolder,
		modelFolder:          modelFolder,

		problemImagePrefix: problemImagePrefix,
		algoImagePrefix:    algoImagePrefix,
		containerRuntime:   containerRuntime,

		storage: storage,
		peer:    peer,
	}
}

// HandleLearn manages a learning task (peer status updates, etc...)
func (w *Worker) HandleLearn(message []byte) (err error) {
	log.Println("[DEBUG][learn] Starting learning task")

	// Unmarshal the learn-uplet
	var task common.Learnuplet
	err = json.NewDecoder(bytes.NewReader(message)).Decode(&task)
	if err != nil {
		return fmt.Errorf("Error un-marshaling learn-uplet: %s -- Body: %s", err, message)
	}

	if err = task.Check(); err != nil {
		return fmt.Errorf("Error in train task: %s -- Body: %s", err, message)
	}

	// Update its status to pending on the peer
	_, _, err = w.peer.SetUpletWorker(task.Key, w.ID.String())
	if err != nil {
		return fmt.Errorf("Error setting uplet worker: %s", err)
	}

	err = w.LearnWorkflow(task)
	if err != nil {
		// TODO: handle fatal and non-fatal errors differently and set learnuplet status to failed only
		// if the error was fatal
		var m map[string]float64
		var f float64
		_, _, err2 := w.peer.ReportLearn(task.Key, common.TaskStatusFailed, f, m, m)