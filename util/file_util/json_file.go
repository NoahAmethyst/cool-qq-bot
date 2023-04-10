package file_util

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func WriteJsonFile(data interface{}, targetPath, fileName string) (string, error) {
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
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(data); err != nil {
		log.Error().Msgf("Error encoding data:%s", err)
		return filePath, err
	}
	return filePath, nil
}
