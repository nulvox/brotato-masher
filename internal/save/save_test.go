package save

import (
	"encoding/json"
	"testing"
)

const fixture = `{"version":3,"characters_unlocked":[1,2],"consumables_unlocked":[3],"items_unlocked":[],"upgrades_unlocked":[],"weapons_unlocked":[],"zones_unlocked":[0],"data":{"enemies_killed":4,"run_won":0},"unknown":{"keep":true}}`

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
		"stats":               map[string]any{"enemies_killed": float64(99)},
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
