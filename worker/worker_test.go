
package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	. "github.com/MorpheoOrg/morpheo-compute/worker"
	"github.com/MorpheoOrg/morpheo-go-packages/client"
	"github.com/MorpheoOrg/morpheo-go-packages/common"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

var (
	worker      *Worker
	fixtures    *common.DataParser
	tmpPathData string
	// preduplet   = &common.Preduplet{
	// 	ID:                  uuid.NewV4(),
	// 	Problem:             uuid.NewV4(),
	// 	Workflow:            uuid.NewV4(),
	// 	Model:               uuid.NewV4(),
	// 	Data:                uuid.NewV4(),
	// 	WorkerID:            uuid.NewV4(),
	// 	Status:              "todo",
	// 	RequestDate:         22,
	// 	CompletionDate:      22,
	// 	PredictionStorageID: uuid.NewV4(),
	// }
	learnuplet = &common.Learnuplet{
		Key:            "learnuplet" + uuid.NewV4().String(),
		Problem:        uuid.NewV4(),
		TrainData:      []uuid.UUID{uuid.NewV4(), uuid.NewV4()},
		TestData:       []uuid.UUID{uuid.NewV4(), uuid.NewV4()},
		Algo:           uuid.NewV4(),
		ModelStart:     uuid.NewV4(),
		ModelEnd:       uuid.NewV4(),
		Rank:           0,
		Worker:         uuid.NewV4(),
		Status:         "todo",
		RequestDate:    22,
		CompletionDate: 22,
	}
)

const (
	perfString = "{\"perf\":0.5,\"train_perf\":{\"p\":0.5},\"test_perf\":{\"p\":0.5}}"
)

func TestMain(m *testing.M) {
	// Let's hook to our container mock
	containerRuntime := common.NewMockRuntime()

	// Create storage Mock
	storageMock, err := client.NewStorageAPIMock()
	if err != nil {
		log.Panicln("Error loading Storage Mock: ", err)
	}

	// Let's finally create our worker
	tmpPathData = filepath.Join(os.TempDir(), "morpheo_tmp_data")
	worker = NewWorker(
		tmpPathData, "train", "test", "untargeted_test", "pred", "perf", "model",
		"problem", "algo", containerRuntime,
		storageMock, &client.PeerMock{},
	)

	// Run the tests
	exitcode := m.Run()

	// Wipe out everything once the tests are done/failed
	if err := os.RemoveAll(tmpPathData); err != nil {
		log.Println(err)
	}

	os.Exit(exitcode)
}

func TestHandleLearn(t *testing.T) {
	// t.Parallel()

	// Pre-setup directory structure for Learn to avoid permission issues
	taskDataFolder := filepath.Join(tmpPathData, learnuplet.Algo.String())
	assert.Nil(t, worker.SetupDirectories(taskDataFolder, 0777))

	// Create performance.json in tmp
	tmpPathPerfFile := filepath.Join(taskDataFolder, "perf/performance.json")
	f, err := os.Create(tmpPathPerfFile)
	assert.Nil(t, err)
	_, err = f.Write([]byte(perfString))
	assert.Nil(t, err)
	f.Close()

	// Test the whole pipeline works...
	msg, _ := json.Marshal(learnuplet)
	assert.Nil(t, worker.HandleLearn(msg))
}

// func TestHandlePred(t *testing.T) {
// 	// t.Parallel()

// 	// Pre-setup directory structure for Pred to avoid permission issues
// 	taskDataFolder := filepath.Join(tmpPathData, preduplet.Model.String())
// 	assert.Nil(t, worker.SetupDirectories(taskDataFolder, 0777))

// 	// Create a tar-gzed model mock in tmp
// 	modelID := preduplet.Data.String()