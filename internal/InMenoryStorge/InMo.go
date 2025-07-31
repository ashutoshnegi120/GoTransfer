/*
this file is make for test perpos and 
we use the sqlite for product ready project not inmemorystorage
*/
package inmemorystorage

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Savefile struct {
	FileID   uuid.UUID `json:"file_id"`
	FileName string    `json:"file_name"`
	Path     string    `json:"path"`
	Time     time.Time `json:"time_at_created"`
}

type FileStatus string

const (
	StatusUploaded  FileStatus = "uploaded"
	StatusProcessing FileStatus = "processing"
	StatusAvailable  FileStatus = "available"
	StatusFailed     FileStatus = "failed"
)

type FileTrack struct {
	FileID uuid.UUID
	Status FileStatus
}

var FileStatusStore = make(map[uuid.UUID]FileStatus)
var MU sync.RWMutex
var storage = make(map[uuid.UUID]Savefile)

func (s *Savefile) New() (*Savefile, error) {
	if s.FileName == "" || s.Path == "" {
		return nil, errors.New("file name or path cannot be empty")
	}

	temp := Savefile{
		FileID:   s.FileID,
		FileName: s.FileName,
		Path:     s.Path,
		Time:     time.Now(),
	}

	storage[temp.FileID] = temp
	return &temp, nil
}

func GetPath(fileID uuid.UUID) (Savefile, error) {
	file, ok := storage[fileID]
	if !ok {
		return Savefile{}, errors.New("file ID not found in storage")
	}
	return file, nil
}
