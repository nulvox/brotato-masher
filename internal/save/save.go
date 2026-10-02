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
	"challenges_completed",
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
	if _, ok := doc["read_announcements"].([]any); !ok {
		return nil, errors.New("invalid save: read_announcements must be an array")
	}
	if _, ok := doc["difficulties_unlocked"].([]any); !ok {
		return nil, errors.New("invalid save: difficulties_unlocked must be an array")
	}
	if err := validateDifficulties(doc["difficulties_unlocked"]); err != nil {
		return nil, err
	}
	data, ok := doc["data"].(map[string]any)
	if !ok || data == nil {
		return nil, errors.New("invalid save: data must be an object")
	}
	for _, key := range stats {
		if value, present := data[key]; present {
			if key == "chal_hourglass_quit_wave" {
				if _, ok := value.(bool); !ok {
					if _, ok := integerValue(value); !ok {
						return nil, fmt.Errorf("invalid save: data.%s must be boolean or integer", key)
					}
				}
				continue
			}
			if _, ok := integerValue(value); !ok {
				return nil, fmt.Errorf("invalid save: data.%s must be a non-negative integer", key)
			}
		}
	}
	for _, key := range []string{"items_bought", "killed_enemies", "killed_by_enemies"} {
		if value, ok := doc[key].(map[string]any); !ok || !numericMap(value) {
			return nil, fmt.Errorf("invalid save: %s must map IDs to non-negative integers", key)
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
		"version": 3,
		"counts":  counts,
		"stats":   statValues(doc),
		"counters": map[string]any{
			"items_bought":      len(doc["items_bought"].(map[string]any)),
			"killed_enemies":    len(doc["killed_enemies"].(map[string]any)),
			"killed_by_enemies": len(doc["killed_by_enemies"].(map[string]any)),
		},
		"difficulties": len(doc["difficulties_unlocked"].([]any)),
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

// Apply mutates only explicitly supported paths on a decoded copy. Unknown fields remain intact.
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
	if value, present := edits["read_announcements"]; present {
		if _, ok := value.([]any); !ok {
			return nil, errors.New("read_announcements must be an array")
		}
		doc["read_announcements"] = value
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
			if key == "chal_hourglass_quit_wave" {
				if flag, ok := rawValue.(bool); ok {
					data[key] = flag
					continue
				}
			}
			n, ok := integerValue(rawValue)
			if !ok {
				return nil, fmt.Errorf("%s must be a non-negative integer", key)
			}
			data[key] = n
		}
	}
	for _, key := range []string{"items_bought", "killed_enemies", "killed_by_enemies"} {
		if value, present := edits[key]; present {
			counters, ok := value.(map[string]any)
			if !ok || !numericMap(counters) {
				return nil, fmt.Errorf("%s must map IDs to non-negative integers", key)
			}
			doc[key] = counters
		}
	}
	if value, present := edits["difficulties_unlocked"]; present {
		if err := validateDifficulties(value); err != nil {
			return nil, err
		}
		doc["difficulties_unlocked"] = value
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

func numericMap(value map[string]any) bool {
	for key, item := range value {
		if key == "" {
			return false
		}
		if _, ok := integerValue(item); !ok {
			return false
		}
	}
	return true
}

func validateDifficulties(value any) error {
	rows, ok := value.([]any)
	if !ok {
		return errors.New("difficulties_unlocked must be an array")
	}
	for i, rowValue := range rows {
		row, ok := rowValue.(map[string]any)
		if !ok {
			return fmt.Errorf("difficulty row %d must be an object", i)
		}
		if _, ok := row["character_id"].(string); !ok {
			return fmt.Errorf("difficulty row %d has invalid character_id", i)
		}
		zones, ok := row["zones_difficulty_info"].([]any)
		if !ok {
			return fmt.Errorf("difficulty row %d has invalid zones_difficulty_info", i)
		}
		for j, zoneValue := range zones {
			zone, ok := zoneValue.(map[string]any)
			if !ok {
				return fmt.Errorf("difficulty row %d zone %d must be an object", i, j)
			}
			for _, key := range []string{"difficulty_selected_value", "max_selectable_difficulty", "zone_id"} {
				if _, ok := integerValue(zone[key]); !ok {
					return fmt.Errorf("difficulty row %d zone %d has invalid %s", i, j, key)
				}
			}
			for _, key := range []string{"max_difficulty_beaten", "max_endless_wave_beaten"} {
				best, ok := zone[key].(map[string]any)
				if !ok {
					return fmt.Errorf("difficulty row %d zone %d has invalid %s", i, j, key)
				}
				for _, nested := range []string{"difficulty_value", "wave_number"} {
					if _, ok := signedIntegerValue(best[nested]); !ok {
						return fmt.Errorf("difficulty row %d zone %d has invalid %s.%s", i, j, key, nested)
					}
				}
				for _, nested := range []string{"enemy_damage", "enemy_health", "enemy_speed", "nightmare_proj", "retries", "used_ban_count"} {
					if _, ok := integerValue(best[nested]); !ok {
						return fmt.Errorf("difficulty row %d zone %d has invalid %s.%s", i, j, key, nested)
					}
				}
				if flag, ok := best["is_coop"].(bool); !ok || !flag && best["is_coop"] != false {
					return fmt.Errorf("difficulty row %d zone %d has invalid %s.is_coop", i, j, key)
				}
			}
		}
	}
	return nil
}

func signedIntegerValue(value any) (float64, bool) {
	switch n := value.(type) {
	case float64:
		return n, n == float64(int64(n))
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	default:
		return 0, false
	}
}

func integerValue(value any) (float64, bool) {
	n, ok := signedIntegerValue(value)
	return n, ok && n >= 0
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
