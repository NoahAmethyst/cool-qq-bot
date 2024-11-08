package file_util

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func WriteJsonFile(data interface{}, targetPath, fileName string, format bool) (string, error) {
	fileName = strings.ReplaceAll(fileName, ".json", "")
	filePath := fmt.Sprintf("%s/%s.%s", targetPath, fileName, "json")
	file, err := os.Create(filePath)
	if err != nil {
		log.Errorf("Creating file [%s] failed %s", fileName, err.Error())
		return filePath, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	encoder := json.NewEncoder(file)
	if format {
		encoder.SetIndent("", "    ")
	}
	if err := encoder.Encode(data); err != nil {
		log.Errorf("Encoding data failed %s", err.Error())
		return filePath, err
	}
	return filePath, nil
}

func LoadJsonFile(file string, data interface{}) error {
	bytes, err := os.ReadFile(file)
	if len(bytes) == 0 {
		log.Warnf("Json file [%s] is empty", file)
		return nil
	} else if err != nil {
		log.Errorf("Read json file [%s] failed %s", file, err.Error())
		return err
	}

	err = json.Unmarshal(bytes, data)
	if err != nil {
		log.Errorf("Unmarshal json from json file [%s] failed %s", file, err.Error())
	}
	return err
}

func ClearFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// clear file
	err = file.Truncate(0)
	return err
}
