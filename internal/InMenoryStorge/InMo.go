package inmemorystorage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Savefile struct {
	FileID   uuid.UUID `json:"file_id"`
	FileName string    `json:"file_name"`
	Path     string    `json:"path"`
	Time     time.Time `json:"time_at_created"`
}

var storage = make(map[uuid.UUID]Savefile)

func (s *Savefile) New() (*Savefile, error) {
	if s.FileName == "" || s.Path == "" {
		return nil, errors.New("file name or path cannot be empty")
	}

	temp := Savefile{
		FileID:   uuid.New(),
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
