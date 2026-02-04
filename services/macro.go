package services

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"eman-backend/config"
)

type MacroService struct {
	cfg *config.Config
}

func NewMacroService(cfg *config.Config) *MacroService {
	return &MacroService{cfg: cfg}
}

func (s *MacroService) generateToken(timestamp int64) string {
	data := fmt.Sprintf("%s%d%s", s.cfg.Domain, timestamp, s.cfg.AppSecret)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *MacroService) buildURL(endpoint string, params map[string]string) string {
	timestamp := time.Now().Unix()
	token := s.generateToken(timestamp)

	url := fmt.Sprintf("%s%s?domain=%s&time=%d&token=%s",
		s.cfg.MacroAPI, endpoint, s.cfg.Domain, timestamp, token)

	for key, value := range params {
		url += fmt.Sprintf("&%s=%s", key, value)
	}

	return url
}

func (s *MacroService) fetch(url string) (json.RawMessage, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	return json.RawMessage(body), nil
}

func (s *MacroService) GetComplexes() (json.RawMessage, error) {
	url := s.buildURL("/estate/group/getComplexes/", nil)
	return s.fetch(url)
}

func (s *MacroService) GetEstates(params map[string]string) (json.RawMessage, error) {
	url := s.buildURL("/estate/get/", params)
	return s.fetch(url)
}
