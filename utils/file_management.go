package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"

	"zenoforge.com/goLiveNotif/log"
)

func SaveDataToFile(data interface{}, dirPath string, filename string) error {
	if err := os.MkdirAll(dirPath, 0770); err != nil {
		return err
	}

	filePath := filepath.Join(dirPath, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

// data input must be a pointer do decode the file into
func LoadDataFromFile(data interface{}, dirPath string, filename string) error {
	if reflect.ValueOf(data).Kind() != reflect.Ptr {
		return errors.New("data parameter must be a pointer")
	}

	fullPath := filepath.Join(dirPath, filename)

	log.Info(fullPath)

	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}

	return nil
}
