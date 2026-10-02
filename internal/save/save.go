package save

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

var collections = []string{
	"characters_unlocked", "consumables_unlocked", "items_unlocked",
	"upgrades_unlocked", "weapons_unlocked", "zones_unlocked",
}

var stats = []string{
	"chal_hourglass_quit_wave", "enemies_killed", "enemies_killed_far_away",
	"evil_mob_killed", "evil_mob_killed_by", "fruit_eaten_full_hp",
	"is_unlock_all_save", "materials_collected", "run_started", "run_won",
	"steps_taken", "trees_killed",
}

func Parse(raw []byte) (map[string]any, error) {
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if doc == nil {
		return nil, errors.New("save must be a JSON object")
	}
	version, ok := doc["version"].(float64)
	if !ok || version != 3 {
		return nil, errors.New("unsupported save: expected Brotato save version 3")
	}
	for _, key := range collections {
		value, ok := doc[key].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid save: %s must be an array", key)
		}
		if _, err := numericArray(value); err != nil {
			return nil, fmt.Errorf("invalid save: %s: %w", key, err)
		}
	}
	data, ok := doc["data"].(map[string]any)
	if !ok || data == nil {
		return nil, errors.New("invalid save: data must be an object")
	}
	for _, key := range stats {
		if value, present := data[key]; present {
			if _, ok := integerValue(value); !ok {
				return nil, fmt.Errorf("invalid save: data.%s must be a non-negative integer", key)
			}
		}
	}
	return doc, nil
}

func Summary(raw []byte) (map[string]any, error) {
	doc, err := Parse(raw)
	if err != nil {
		return nil, err
	}
	counts := map[string]any{}
	for _, key := range collections {
		counts[key] = len(doc[key].([]any))
	}
	return map[string]any{
		"version": 3, "counts": counts, "stats": statValues(doc),
	}, nil
}

func statValues(doc map[string]any) map[string]any {
	data := doc["data"].(map[string]any)
	out := map[string]any{}
	for _, key := range stats {
		if value, ok := data[key]; ok {
			out[key] = value
		}
	}
	return out
}

// Apply mutates only the explicitly supported paths on a decoded copy.
// Unknown top-level keys and nested values are retained by json.Unmarshal/Marshal.
func Apply(raw []byte, edits map[string]any) ([]byte, error) {
	doc, err := Parse(raw)
	if err != nil {
		return nil, err
	}
	for _, key := range collections {
		if value, present := edits[key]; present {
			ids, err := numericArray(value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			doc[key] = ids
		}
	}
	if value, present := edits["stats"]; present {
		statEdits, ok := value.(map[string]any)
		if !ok {
			return nil, errors.New("stats edits must be an object")
		}
		data := doc["data"].(map[string]any)
		for key, rawValue := range statEdits {
			if !contains(stats, key) {
				return nil, fmt.Errorf("unsupported statistic: %s", key)
			}
			n, ok := integerValue(rawValue)
			if !ok {
				return nil, fmt.Errorf("%s must be a non-negative integer", key)
			}
			data[key] = n
		}
	}
	return marshalPretty(doc)
}

func numericArray(value any) ([]any, error) {
	array, ok := value.([]any)
	if !ok {
		return nil, errors.New("must be an array")
	}
	out := make([]any, len(array))
	for i, item := range array {
		n, ok := integerValue(item)
		if !ok {
			return nil, fmt.Errorf("entry %d must be a non-negative integer", i)
		}
		out[i] = n
	}
	return out, nil
}

func integerValue(value any) (float64, bool) {
	switch n := value.(type) {
	case float64:
		return n, n >= 0 && n == float64(uint64(n))
	case int:
		return float64(n), n >= 0
	case int64:
		return float64(n), n >= 0
	case uint64:
		return float64(n), true
	default:
		return 0, false
	}
}

func marshalPretty(doc map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func CollectionNames() []string {
	out := append([]string(nil), collections...)
	sort.Strings(out)
	return out
}
func StatNames() []string { out := append([]string(nil), stats...); sort.Strings(out); return out }
