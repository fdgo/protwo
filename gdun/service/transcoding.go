package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"io"
	"os"
)

//----------------------------------------------------------------------------

type TranscodingTaskType int

const (
	PDF TranscodingTaskType = 1 + iota
	IMAGE
	AUDIO
	VIDEO
)

//----------------------------------------------------------------------------

type TranscodingTask struct {
	FileName   string
	StorageKey string
	Type       TranscodingTaskType
	UserID     int
	UserIP     string
}

type TranscodingResult struct {
	UserID     int
	StorageKey string
	Pages      int
	Status     int
	Info       string
}

//----------------------------------------------------------------------------

type TranscodingService struct {
	cwOss       *common.ObjectStorage
	cwKeyPrefix string
	ghostScript string
	tmpDir      string
	results     chan *TranscodingResult
	tasks       chan *TranscodingTask
}

func NewTranscodingService(cwOss *common.ObjectStorage, cwKeyPrefix string, ghostScript string, tmpDir string, results chan *TranscodingResult) *TranscodingService {
	ts := new(TranscodingService)

	ts.cwOss = cwOss
	ts.cwKeyPrefix = cwKeyPrefix

	ts.ghostScript = ghostScript

	ts.tmpDir = tmpDir

	ts.results = results // This channel might be null.
	ts.tasks = make(chan *TranscodingTask)

	go ts.start()

	return ts
}

//----------------------------------------------------------------------------

func (ts *TranscodingService) start() {
	for {
		select {
		case t, okay := <-ts.tasks:
			if okay {
				ts.transcode(t)
			}
		}
	}
}

//----------------------------------------------------------------------------

func (ts *TranscodingService) transcode(t *TranscodingTask) {

	switch t.Type {
	case PDF:
		status, size, err := ts.TrascodePDF(t)
		if ts.results != nil {
			r := new(TranscodingResult)
			r.UserID = t.UserID
			r.StorageKey = t.StorageKey
			r.Pages = size
			r.Status = status
			if err != nil {
				r.Info = err.Error()
			} else {
				r.Info = ""
			}
			ts.results <- r
		}
	case IMAGE:
	case AUDIO:
	case VIDEO:
	}
}

func (ts *TranscodingService) computeStorageKey(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, err
	}

	b := h.Sum(([]byte)(""))
	return b, nil
}

//----------------------------------------------------------------------------

func (ts *TranscodingService) AddFile(fileName string, taskType TranscodingTaskType, ip string, userID int) (string, error) {
	// Generate a storage key for this PDF file.
	b, err := ts.computeStorageKey(fileName)
	if err != nil {
		return "", err
	}

	// Initiate a task.
	task := new(TranscodingTask)
	task.FileName = fileName
	task.Type = taskType
	task.StorageKey = fmt.Sprintf("%x", b)
	task.UserID = userID
	task.UserIP = ip

	switch task.Type {
	case PDF:
		// Check whether the transcoding tool exists or not.
		if len(ts.ghostScript) == 0 {
			return "", errors.New("No GhostScript available.")
		}

		// Check whether this file exists or not.
		existing, err := ts.cwOss.Exist(ts.cwKeyPrefix + task.StorageKey)
		if err != nil {
			return "", err
		}
		if existing {
			// We prefer to covering the existing files yet.
			return task.StorageKey, nil
		}
	}

	// Put this task to task queue.
	ts.tasks <- task

	return task.StorageKey, nil
}

//----------------------------------------------------------------------------
