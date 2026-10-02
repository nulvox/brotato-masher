package save

import (
	"encoding/json"
	"testing"
)

const fixture = `{"version":3,"characters_unlocked":[1,2],"consumables_unlocked":[3],"items_unlocked":[],"upgrades_unlocked":[],"weapons_unlocked":[],"zones_unlocked":[0],"challenges_completed":[10],"data":{"enemies_killed":4,"run_won":0,"chal_hourglass_quit_wave":false},"difficulties_unlocked":[{"character_id":"character_test","zones_difficulty_info":[{"difficulty_selected_value":0,"max_selectable_difficulty":6,"zone_id":0,"max_difficulty_beaten":{"difficulty_value":-1,"enemy_damage":1,"enemy_health":1,"enemy_speed":1,"is_coop":false,"nightmare_proj":1,"retries":0,"used_ban_count":0,"wave_number":-1},"max_endless_wave_beaten":{"difficulty_value":-1,"enemy_damage":1,"enemy_health":1,"enemy_speed":1,"is_coop":false,"nightmare_proj":1,"retries":0,"used_ban_count":0,"wave_number":-1}}]}],"inactive_mods":[],"items_bought":{"1":2},"killed_by_enemies":{"2":1},"killed_enemies":{"2":4},"read_announcements":[],"unknown":{"keep":true}}`

func TestSummary(t *testing.T) {
	got, err := Summary([]byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	if got["version"] != 3 {
		t.Fatalf("version = %#v", got["version"])
	}
}

func TestApplyPreservesUnknownFields(t *testing.T) {
	out, err := Apply([]byte(fixture), map[string]any{
		"characters_unlocked": []any{9},
		"stats":               map[string]any{"enemies_killed": float64(99), "chal_hourglass_quit_wave": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["unknown"].(map[string]any)["keep"] != true {
		t.Fatal("unknown field was not preserved")
	}
	if len(doc["characters_unlocked"].([]any)) != 1 {
		t.Fatal("collection was not replaced")
	}
}

func TestRejectsUnsupportedVersion(t *testing.T) {
	if _, err := Parse([]byte(`{"version":2}`)); err == nil {
		t.Fatal("expected version error")
	}
}

func TestRejectsNonNumericCollectionIDs(t *testing.T) {
	bad := []byte(`{"version":3,"characters_unlocked":["unsafe"],"consumables_unlocked":[],"items_unlocked":[],"upgrades_unlocked":[],"weapons_unlocked":[],"zones_unlocked":[],"data":{}}`)
	if _, err := Parse(bad); err == nil {
		t.Fatal("expected collection ID error")
	}
}
