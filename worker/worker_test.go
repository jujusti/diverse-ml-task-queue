
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