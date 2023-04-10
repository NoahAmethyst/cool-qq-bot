package file_util

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func WriteJsonFile(data interface{}, targetPath, fileName string, format bool) (string, error) {
	fileName = strings.ReplaceAll(fileName, ".json", "")
	filePath := fmt.Sprintf("%s/%s.%s", targetPath, fileName, "json")
	file, err := os.Create(filePath)
	if err != nil {
		log.Error().Msgf("Error creating file:%s", err)
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
		log.Error().Msgf("Error encoding data:%s", err)
		return filePath, err
	}
	return filePath, nil
}

func LoadJsonFile(file string, data interface{}) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "read json file",
			"error":  err,
		}).Send()
		return err
	}

	err = json.Unmarshal(bytes, data)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": fmt.Sprintf("unmarshal json from json file %s", file),
			"error":  err,
		}).Send()
	}
	return err

}
