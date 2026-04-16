package services

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"eman-backend/config"
)

type MacroService struct {
	cfg             *config.Config
	estatesMu       sync.RWMutex
	estatesSnapshot []map[string]any
	lastEstatesSync time.Time
}

func NewMacroService(cfg *config.Config) *MacroService {
	service := &MacroService{cfg: cfg}
	service.startEstatesCacheRefresher()
	return service
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

// MacroRequestPayload represents the body for POST /estate/request/
type MacroRequestPayload struct {
	Domain   string `json:"domain"`
	Time     int64  `json:"time"`
	Token    string `json:"token"`
	Action   string `json:"action"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email,omitempty"`
	Message  string `json:"message,omitempty"`
	EstateID *int   `json:"id,omitempty"`
}

// MacroRequestResponse represents the response from POST /estate/request/
type MacroRequestResponse struct {
	Success  bool   `json:"success"`
	EstateID int    `json:"estate_id,omitempty"`
	Error    bool   `json:"error"`
	Message  string `json:"message,omitempty"`
}

// SendRequest sends a lead/submission to MacroCRM
func (s *MacroService) SendRequest(action, name, phone, email, message string, estateID *int) (*MacroRequestResponse, error) {
	timestamp := time.Now().Unix()
	token := s.generateToken(timestamp)

	payload := MacroRequestPayload{
		Domain:   s.cfg.Domain,
		Time:     timestamp,
		Token:    token,
		Action:   action,
		Name:     name,
		Phone:    phone,
		Email:    email,
		Message:  message,
		EstateID: estateID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload failed: %w", err)
	}

	url := fmt.Sprintf("%s/estate/request/", s.cfg.MacroAPI)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("macro request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result MacroRequestResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w (body: %s)", err, string(body))
	}

	if result.Error {
		return &result, fmt.Errorf("macro API error: %s", result.Message)
	}

	return &result, nil
}

func (s *MacroService) GetComplexes() (json.RawMessage, error) {
	url := s.buildURL("/estate/group/getComplexes/", nil)
	return s.fetch(url)
}

func (s *MacroService) startEstatesCacheRefresher() {
	// Initial warmup.
	go func() {
		if err := s.refreshEstatesSnapshot(); err != nil {
			log.Printf("[MacroCache] initial estates sync failed: %v", err)
		}
	}()

	// Periodic refresh every 30 minutes.
	go func() {
		ticker := time.NewTicker(s.cfg.MacroEstateSyncInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := s.refreshEstatesSnapshot(); err != nil {
				log.Printf("[MacroCache] periodic estates sync failed, keeping old snapshot: %v", err)
			}
		}
	}()
}

func (s *MacroService) refreshEstatesSnapshot() error {
	url := s.buildURL("/estate/get/", nil)
	raw, err := s.fetch(url)
	if err != nil {
		return err
	}

	var estates []map[string]any
	if err := json.Unmarshal(raw, &estates); err != nil {
		return fmt.Errorf("decode estates payload failed: %w", err)
	}

	s.estatesMu.Lock()
	s.estatesSnapshot = estates
	s.lastEstatesSync = time.Now()
	s.estatesMu.Unlock()

	log.Printf("[MacroCache] estates snapshot updated: %d items", len(estates))
	return nil
}

func (s *MacroService) getEstatesSnapshot() []map[string]any {
	s.estatesMu.RLock()
	defer s.estatesMu.RUnlock()

	if len(s.estatesSnapshot) == 0 {
		return nil
	}

	// Shallow copy is enough; maps are treated as read-only.
	out := make([]map[string]any, len(s.estatesSnapshot))
	copy(out, s.estatesSnapshot)
	return out
}

func findEstateByID(estates []map[string]any, id int) map[string]any {
	for _, item := range estates {
		itemID, ok := getNumberField(item, "id")
		if ok && int(itemID) == id {
			return item
		}
	}
	return nil
}

func estateHumanTitle(item map[string]any) string {
	title := strings.TrimSpace(getStringField(item, "title"))
	if title != "" {
		return title
	}
	address := strings.TrimSpace(getStringField(item, "address"))
	if address != "" {
		return address
	}
	return ""
}

func (s *MacroService) fetchEstateByID(id int) map[string]any {
	url := s.buildURL("/estate/get/", map[string]string{
		"id": strconv.Itoa(id),
	})
	raw, err := s.fetch(url)
	if err != nil {
		return nil
	}

	var estates []map[string]any
	if err := json.Unmarshal(raw, &estates); err != nil {
		return nil
	}
	return findEstateByID(estates, id)
}

// GetEstateTitleByID returns the human-readable estate title.
// Priority: cached snapshot -> one forced refresh -> direct API by id.
func (s *MacroService) GetEstateTitleByID(id int) string {
	if id <= 0 {
		return ""
	}

	if item := findEstateByID(s.getEstatesSnapshot(), id); item != nil {
		return estateHumanTitle(item)
	}

	// One opportunistic refresh for better hit chance.
	if err := s.refreshEstatesSnapshot(); err == nil {
		if item := findEstateByID(s.getEstatesSnapshot(), id); item != nil {
			return estateHumanTitle(item)
		}
	}

	if item := s.fetchEstateByID(id); item != nil {
		return estateHumanTitle(item)
	}

	return ""
}

func getNumberField(item map[string]any, key string) (float64, bool) {
	v, ok := item[key]
	if !ok || v == nil {
		return 0, false
	}

	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case json.Number:
		parsed, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func getStringField(item map[string]any, key string) string {
	v, ok := item[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func parseNumberParam(values []string) (float64, bool) {
	if len(values) == 0 {
		return 0, false
	}
	last := strings.TrimSpace(values[len(values)-1])
	if last == "" {
		return 0, false
	}
	n, err := strconv.ParseFloat(last, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func parseIntParam(values []string, defaultValue int) int {
	if len(values) == 0 {
		return defaultValue
	}
	last := strings.TrimSpace(values[len(values)-1])
	if last == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(last)
	if err != nil {
		return defaultValue
	}
	return n
}

func anyValueMatches(raw string, candidates []string) bool {
	if len(candidates) == 0 {
		return true
	}
	for _, candidate := range candidates {
		if strings.EqualFold(strings.TrimSpace(raw), strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func anyNumberMatches(raw float64, candidates []string) bool {
	if len(candidates) == 0 {
		return true
	}
	for _, candidate := range candidates {
		n, err := strconv.ParseFloat(strings.TrimSpace(candidate), 64)
		if err != nil {
			continue
		}
		if raw == n {
			return true
		}
	}
	return false
}

func (s *MacroService) applyEstateFilters(estates []map[string]any, params url.Values) []map[string]any {
	filtered := make([]map[string]any, 0, len(estates))

	typeValues := params["type"]
	activityValues := params["activity"]
	categoryValues := params["category"]
	roomValues := params["rooms"]
	floorValues := params["floor"]

	priceFrom, hasPriceFrom := parseNumberParam(params["price_from"])
	priceTo, hasPriceTo := parseNumberParam(params["price_to"])
	areaFrom, hasAreaFrom := parseNumberParam(params["area_from"])
	areaTo, hasAreaTo := parseNumberParam(params["area_to"])
	floorFrom, hasFloorFrom := parseNumberParam(params["floor_from"])
	floorTo, hasFloorTo := parseNumberParam(params["floor_to"])

	for _, item := range estates {
		if !anyValueMatches(getStringField(item, "type"), typeValues) {
			continue
		}
		if !anyValueMatches(getStringField(item, "activity"), activityValues) {
			continue
		}
		if !anyValueMatches(getStringField(item, "category"), categoryValues) {
			continue
		}

		if rooms, ok := getNumberField(item, "estate_rooms"); ok && !anyNumberMatches(rooms, roomValues) {
			continue
		}
		if floor, ok := getNumberField(item, "estate_floor"); ok && !anyNumberMatches(floor, floorValues) {
			continue
		}

		if hasPriceFrom {
			if price, ok := getNumberField(item, "estate_price"); !ok || price < priceFrom {
				continue
			}
		}
		if hasPriceTo {
			if price, ok := getNumberField(item, "estate_price"); !ok || price > priceTo {
				continue
			}
		}
		if hasAreaFrom {
			if area, ok := getNumberField(item, "estate_area"); !ok || area < areaFrom {
				continue
			}
		}
		if hasAreaTo {
			if area, ok := getNumberField(item, "estate_area"); !ok || area > areaTo {
				continue
			}
		}
		if hasFloorFrom {
			if floor, ok := getNumberField(item, "estate_floor"); !ok || floor < floorFrom {
				continue
			}
		}
		if hasFloorTo {
			if floor, ok := getNumberField(item, "estate_floor"); !ok || floor > floorTo {
				continue
			}
		}

		filtered = append(filtered, item)
	}

	offset := parseIntParam(params["offset"], 0)
	if offset < 0 {
		offset = 0
	}
	if offset >= len(filtered) {
		return []map[string]any{}
	}

	limit := parseIntParam(params["limit"], 0)
	if limit <= 0 {
		return filtered[offset:]
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end]
}

func (s *MacroService) GetEstates(params url.Values) (json.RawMessage, error) {
	estates := s.getEstatesSnapshot()
	if len(estates) == 0 {
		// Try immediate refresh on cold start / empty cache.
		if err := s.refreshEstatesSnapshot(); err != nil {
			return nil, fmt.Errorf("estates cache unavailable and refresh failed: %w", err)
		}
		estates = s.getEstatesSnapshot()
	}

	filtered := s.applyEstateFilters(estates, params)
	out, err := json.Marshal(filtered)
	if err != nil {
		return nil, fmt.Errorf("encode estates response failed: %w", err)
	}
	return out, nil
}
