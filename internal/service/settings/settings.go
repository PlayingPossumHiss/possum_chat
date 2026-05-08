package settings

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type Service struct {
	settings entity.Config
	mx       *sync.Mutex
}

func New() (*Service, error) {
	settings, err := getSettingsFromFile()
	if err != nil {
		return nil, err
	}

	return &Service{
		settings: settings,
		mx:       &sync.Mutex{},
	}, nil
}

func (s *Service) UpdateConfig(opts []entity.ConfigUpdateOption) error {
	logger.Info("update config")
	s.mx.Lock()
	defer s.mx.Unlock()
	config := s.settings
	for _, option := range opts {
		option(&config)
	}

	err := saveSettingToFile(config)
	if err != nil {
		return err
	}
	s.settings = config

	return nil
}

func (s *Service) Config() entity.Config {
	return s.settings
}

func getSettingsFromFile() (entity.Config, error) {
	_, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = saveSettingToFile(defaultConfig)
			if err != nil {
				return entity.Config{}, err
			}
		} else {
			return entity.Config{}, err
		}
	}

	jsonFile, err := os.Open(configPath)
	if err != nil {
		return entity.Config{}, err
	}
	defer func() {
		jsonFile.Close()
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

	config, version, err := configFromJson(settings)
	if err != nil {
		return entity.Config{}, err
	}
	if version != currentVersion {
		config = upgradeConfig(config, version)
	}

	return config, nil
}

// upgradeConfig пока тут ничего, но потом появится
func upgradeConfig(src entity.Config, _ string) entity.Config {
	return src
}

func saveSettingToFile(src entity.Config) error {
	parsedConfig := configToJson(src)
	configBytes, err := json.MarshalIndent(parsedConfig, "", "\t")
	if err != nil {
		return err
	}

	err = createConfigIfNotExist()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC, 0o644) //nolint
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(configBytes)
	if err != nil {
		return err
	}

	return nil
}

func createConfigIfNotExist() error {
	_, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(configPath)
			if err != nil {
				return err
			}

			defer file.Close()
		}
	}

	return nil
}
