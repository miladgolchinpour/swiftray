package services

import (
	"encoding/json"
	"os"
	"path/filepath"

	"SwiftRay/app/models"
	"SwiftRay/app/utils"
)

type StorageService struct {
	dataDir string
}

func NewStorageService() *StorageService {
	dataDir := utils.AppDataDir()
	os.MkdirAll(dataDir, 0o755)
	return &StorageService{dataDir: dataDir}
}

func (s *StorageService) filePath(name string) string {
	return filepath.Join(s.dataDir, name+".json")
}

func (s *StorageService) loadJSON(name string, v interface{}) error {
	data, err := os.ReadFile(s.filePath(name))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (s *StorageService) saveJSON(name string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath(name), data, 0o644)
}

func (s *StorageService) LoadSettings() models.AppSettings {
	settings := models.DefaultSettings()
	if err := s.loadJSON("appSettings", &settings); err != nil {
		return models.DefaultSettings()
	}
	return settings
}

func (s *StorageService) SaveSettings(settings models.AppSettings) error {
	return s.saveJSON("appSettings", settings)
}

func (s *StorageService) LoadSubscriptions() []models.Subscription {
	var subs []models.Subscription
	if err := s.loadJSON("subscriptions", &subs); err != nil {
		return []models.Subscription{}
	}
	return subs
}

func (s *StorageService) SaveSubscriptions(subs []models.Subscription) error {
	return s.saveJSON("subscriptions", subs)
}

func (s *StorageService) LoadLocalNodes() []models.Node {
	var nodes []models.Node
	if err := s.loadJSON("localNodes", &nodes); err != nil {
		return []models.Node{}
	}
	return nodes
}

func (s *StorageService) SaveLocalNodes(nodes []models.Node) error {
	return s.saveJSON("localNodes", nodes)
}

func (s *StorageService) LoadSelectedNodeID() string {
	var id string
	if err := s.loadJSON("selectedNodeID", &id); err != nil {
		return ""
	}
	return id
}

func (s *StorageService) SaveSelectedNodeID(id string) error {
	return s.saveJSON("selectedNodeID", id)
}

func (s *StorageService) LoadSelectedSubID() string {
	var id string
	if err := s.loadJSON("selectedSubID", &id); err != nil {
		return ""
	}
	return id
}

func (s *StorageService) SaveSelectedSubID(id string) error {
	return s.saveJSON("selectedSubID", id)
}
