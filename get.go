package graphqlfixture

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

// Get returns the raw data given the captorName
func (fs *Fixtures) Get(captorName string) (interface{}, error ) {
	if fs.captured == nil {
		return nil, errors.New("captured is empty")
	}
	capturedVal, found := fs.captured[captorName]
	if !found {
		return nil, errors.New("not found")
	}
	return capturedVal, nil
}

// Get parses the data retrieved by given captorName into the desired value type
func (fs *Fixtures) GetAndParse(captorName string, value interface{}) error {
	capturedVal, err := fs.Get(captorName)
	if err != nil {
		return err
	}
	jsonBytes, err := json.Marshal(capturedVal)
	if err != nil {
		return fmt.Errorf("fail to marshal into json: %w", err)
	}
	err = json.Unmarshal(jsonBytes, value)
	if err != nil {
		return fmt.Errorf("fail to unmarshal from json to desired value type: %w", err)
	}
	return nil
}

// Logs return logs
func (fs *Fixtures) Logs() []string{
	return fs.logs
}

// SetupUntil returns setupUntilIdx (can be nil)
func (fs *Fixtures) SetupUntil() *int{
	return fs.setupUntilIdx
}

// TeardownUntil returns teardownUntilIdx (can be nil)
func (fs *Fixtures) TeardownUntil() *int{
	return fs.teardownUntilIdx
}