
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
		if err2 != nil {
			return fmt.Errorf("Error in LearnWorkflow: %s. Error setting learnuplet status to failed on the peer: %s", err, err2)
		}
		return fmt.Errorf("Error in LearnWorkflow: %s", err)
	}
	return nil
}

// HandlePred manages a prediction task (peer status updates, etc...)
func (w *Worker) HandlePred(message []byte) (err error) {
	// log.Println("[DEBUG][pred] Starting predicting task")

	// // Unmarshal the learn-uplet
	// var task common.Preduplet
	// err = json.NewDecoder(bytes.NewReader(message)).Decode(&task)
	// if err != nil {
	// 	return fmt.Errorf("Error un-marshaling preduplet: %s -- Body: %s", err, message)
	// }

	// if err = task.Check(); err != nil {
	// 	return fmt.Errorf("Error in pred task: %s -- Body: %s", err, message)
	// }

	// // Update its status to pending on the peer
	// err = w.peer.UpdateUpletStatus(common.TypePredUplet, common.TaskStatusPending, task.Key, task.Worker)
	// if err != nil {
	// 	return fmt.Errorf("Error setting preduplet status to pending on the peer: %s", err)
	// }

	// err = w.PredWorkflow(task)
	// if err != nil {
	// 	// TODO: handle fatal and non-fatal errors differently and set preduplet status to failed only
	// 	// if the error was fatal
	// 	err2 := w.peer.UpdateUpletStatus(common.TypePredUplet, common.TaskStatusFailed, task.Key, task.Worker)
	// 	if err2 != nil {
	// 		return fmt.Errorf("2 Errors: Error in PredWorkflow: %s. Error setting preduplet status to failed on the peer: %s", err, err2)
	// 	}
	// 	return fmt.Errorf("Error in PredWorkflow: %s", err)
	// }
	return nil
}

// LearnWorkflow implements our learning workflow
func (w *Worker) LearnWorkflow(task common.Learnuplet) (err error) {
	log.Printf("[DEBUG][learn] Starting learning workflow for %s", task.Key)

	// Setup directory structure
	taskDataFolder := filepath.Join(w.dataFolder, task.Algo.String())
	trainFolder := filepath.Join(taskDataFolder, w.trainFolder)
	testFolder := filepath.Join(taskDataFolder, w.testFolder)
	untargetedTestFolder := filepath.Join(taskDataFolder, w.untargetedTestFolder)
	modelFolder := filepath.Join(taskDataFolder, w.modelFolder)
	perfFolder := filepath.Join(taskDataFolder, w.perfFolder)

	pathList := []string{taskDataFolder, trainFolder, testFolder, untargetedTestFolder, modelFolder, perfFolder}
	for _, path := range pathList {
		err = os.MkdirAll(path, os.ModeDir)
		if err != nil {
			return fmt.Errorf("Error creating folder under %s: %s", path, err)
		}
	}

	// Let's make sure these folders are wiped out once the task is done/failed
	defer os.RemoveAll(taskDataFolder)

	// Load problem workflow
	problemWorkflow, err := w.storage.GetProblemWorkflowBlob(task.Problem)
	if err != nil {
		return fmt.Errorf("Error pulling problem workflow %s from storage: %s", task.Problem, err)
	}
	problemImageName := fmt.Sprintf("%s-%s", w.problemImagePrefix, task.Problem)
	err = w.ImageLoad(problemImageName, problemWorkflow)
	if err != nil {
		return fmt.Errorf("Error loading problem workflow image %s in Docker daemon: %s", task.Problem, err)
	}
	problemWorkflow.Close()
	defer w.containerRuntime.ImageUnload(problemImageName)

	log.Println("[DEBUG][learn] 1st Image loaded")
	// Load algo
	algo, err := w.storage.GetAlgoBlob(task.Algo)
	if err != nil {
		return fmt.Errorf("Error pulling algo %s from storage: %s", task.Algo, err)
	}

	algoImageName := fmt.Sprintf("%s-%s", w.algoImagePrefix, task.Algo)
	err = w.ImageLoad(algoImageName, algo)
	if err != nil {
		return fmt.Errorf("Error loading algo image %s in Docker daemon: %s", algoImageName, err)
	}
	algo.Close()
	defer w.containerRuntime.ImageUnload(algoImageName)

	// Pull model if a model_start parameter was given in the learn-uplet
	if task.Rank > 0 {
		// Check that modelStart is set
		if uuid.Equal(uuid.Nil, task.ModelStart) {
			return fmt.Errorf("Error in learnuplet: ModelStart is a Nil uuid, although Rank is set to %d", task.Rank)
		}
		// Pull model from storage
		model, err := w.storage.GetModelBlob(task.ModelStart)
		if err != nil {
			return fmt.Errorf("Error pulling start model %s from storage: %s", task.ModelStart, err)
		}
		err = w.UntargzInFolder(modelFolder, model)
		if err != nil {
			return fmt.Errorf("Error un-tar-gz-ing model: %s", err)
		}
		model.Close()
	}

	// Pulling train dataset
	for _, dataID := range task.TrainData {
		data, err := w.storage.GetDataBlob(dataID)
		if err != nil {
			return fmt.Errorf("Error pulling train dataset %s from storage: %s", dataID, err)
		}
		path := fmt.Sprintf("%s/%s", trainFolder, dataID)
		dataFile, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("Error creating file %s: %s", path, err)
		}
		n, err := io.Copy(dataFile, data)
		if err != nil {
			return fmt.Errorf("Error copying train data file %s (%d bytes written): %s", path, n, err)
		}
		dataFile.Close()
		data.Close()
	}

	// And the test data
	for _, dataID := range task.TestData {
		data, err := w.storage.GetDataBlob(dataID)
		if err != nil {
			return fmt.Errorf("Error pulling test dataset %s from storage: %s", dataID, err)
		}
		path := fmt.Sprintf("%s/%s", testFolder, dataID)
		dataFile, err := os.Create(path)
		n, err := io.Copy(dataFile, data)
		if err != nil {
			return fmt.Errorf("Error copying test data file %s (%d bytes written): %s", path, n, err)
		}
		dataFile.Close()
		data.Close()
	}

	// Let's copy test data into untargetedTestFolder and remove targets
	_, err = w.UntargetTestingVolume(problemImageName, testFolder, untargetedTestFolder)
	if err != nil {
		return fmt.Errorf("Error preparing problem %s for model %s: %s", task.Problem, task.ModelStart, err)
	}

	// Let's pass the task to our execution backend, now that everything should be in place
	_, err = w.Train(algoImageName, trainFolder, untargetedTestFolder, modelFolder)
	if err != nil {
		return fmt.Errorf("Error in train task: %s -- Body: %s", err, task)
	}

	// Let's compute the performance !
	_, err = w.ComputePerf(problemImageName, trainFolder, testFolder, untargetedTestFolder, perfFolder)
	if err != nil {
		// FIXME: do not return here
		return fmt.Errorf("Error computing perf for problem %s and model (new) %s: %s", task.Problem, task.ModelEnd, err)
	}

	// Let's create a new model and post it to storage
	algoInfo, err := w.storage.GetAlgo(task.Algo)
	if err != nil {
		return fmt.Errorf("Error retrieving algorithm %s metadata: %s", task.Algo, err)
	}
	newModel := common.NewModel(task.ModelEnd, algoInfo)
	newModel.ID = task.ModelEnd

	// Let's compress our model in a separate goroutine while writing it on disk on the fly
	path := fmt.Sprintf("%s/model.tar.gz", taskDataFolder)
	modelArchiveWriter, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error creating new model archive file %s: %s", path, err)
	}