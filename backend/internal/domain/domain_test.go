package domain_test

import (
	"encoding/json"
	"testing"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
)

func TestIDMarshalJSON(t *testing.T) {
	id := domain.NewID()

	payload := struct {
		ID domain.ID `json:"id"`
	}{ID: id}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	got, ok := decoded["id"]
	if !ok {
		t.Fatalf("expected id key in marshaled output, got %v", decoded)
	}

	if got != id.String() {
		t.Fatalf("expected id %q, got %q", id.String(), got)
	}
}

func TestIDUnmarshalJSON(t *testing.T) {
	id := domain.NewID()

	input := []byte(`{"id":"` + id.String() + `"}`)

	var payload struct {
		ID domain.ID `json:"id"`
	}

	if err := json.Unmarshal(input, &payload); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if payload.ID != id {
		t.Fatalf("expected id %q, got %q", id.String(), payload.ID.String())
	}
}

func TestDTOMuscleGroupMarshalingUsesStringIDs(t *testing.T) {
	id := domain.NewID()
	dtoPayload := dto.MuscleGroupDTO{ID: id, Name: "Chest"}

	data, err := json.Marshal(struct {
		Groups []dto.MuscleGroupDTO `json:"muscle_groups"`
	}{Groups: []dto.MuscleGroupDTO{dtoPayload}})
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded struct {
		Groups []map[string]any `json:"muscle_groups"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(decoded.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(decoded.Groups))
	}

	rawID, ok := decoded.Groups[0]["ID"].(string)
	if !ok {
		t.Fatalf("expected ID as string, got %T", decoded.Groups[0]["ID"])
	}

	if rawID != id.String() {
		t.Fatalf("expected ID %q, got %q", id.String(), rawID)
	}
}
