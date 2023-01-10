package network

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

const (
	networkMappingFileName = "network.json"
)

type HomeDirConfig struct {
	NetworkPathMapping map[string]string `json:"network_path_mapping"`
}

func NewHomeDirConfig() *HomeDirConfig {
	return &HomeDirConfig{NetworkPathMapping: map[string]string{}}
}

func (hdc *HomeDirConfig) GetNetworkPath(name string) string {
	return hdc.NetworkPathMapping[name]
}

func (hdc *HomeDirConfig) AddNetwork(name, dataDir string) error {
	if _, ok := hdc.NetworkPathMapping[name]; ok {
		return errors.New("network exists")
	}
	hdc.NetworkPathMapping[name] = dataDir
	return nil
}

func (hdc *HomeDirConfig) Store() error {
	homeDirPath := os.Getenv("HOME")
	configPath := filepath.Join(homeDirPath, ".fdt", networkMappingFileName)
	dirPath := filepath.Dir(configPath)
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}
	data, _ := json.Marshal(hdc)
	return os.WriteFile(configPath, data, 0755)
}

func Load() (*HomeDirConfig, error) {
	homeDirPath := os.Getenv("HOME")
	configPath := filepath.Join(homeDirPath, ".fdt", networkMappingFileName)
	hdc := NewHomeDirConfig()
	_, err := os.Stat(configPath)
	if err != nil {
		if err = hdc.Store(); err != nil {
			return nil, err
		}
		return hdc, nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "load network mapping config failed")
	}

	if err = json.Unmarshal(data, &hdc); err != nil {
		return nil, errors.Wrap(err, "json unmarshal networkPathMapping failed")
	}

	return hdc, nil
}
