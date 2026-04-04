package settings

import (
	"encoding/json"
	"io"
	"os"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type Service struct {
	settings entity.Config
}

func New() (*Service, error) {
	settings, err := getSettingsFromFile()
	if err != nil {
		return nil, err
	}

	return &Service{
		settings: settings,
	}, nil
}

func (s *Service) Config() entity.Config {
	return s.settings
}

func getSettingsFromFile() (entity.Config, error) {
	jsonFile, err := os.Open(configPath)
	if err != nil {
		return entity.Config{}, err
	}
	defer func() {
		dErr := jsonFile.Close()
		if dErr != nil {
			logger.Error(dErr.Error())
		}
	}()

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return entity.Config{}, err
	}

	var settings config

	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		return entity.Config{}, err
	}

	config, err := configFromJson(settings)
	if err != nil {
		return entity.Config{}, err
	}

	return config, nil
}
