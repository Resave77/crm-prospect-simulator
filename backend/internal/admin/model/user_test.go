package model

import (
	"encoding/json"
	"testing"
)

func TestUpdateUserInputManagerIDDecoding(t *testing.T) {
	uuidStr := "3f2a4b6c-1e5d-4f8a-9c7b-2d3e4f5a6b7c"

	var omitted UpdateUserInput
	if err := json.Unmarshal([]byte(`{"name":"Updated"}`), &omitted); err != nil {
		t.Fatalf("decode omitted: %v", err)
	}
	if omitted.ManagerID.Present {
		t.Fatal("omitted managerId must not be present")
	}

	var nullCase UpdateUserInput
	if err := json.Unmarshal([]byte(`{"managerId":null}`), &nullCase); err != nil {
		t.Fatalf("decode null: %v", err)
	}
	if !nullCase.ManagerID.Present || nullCase.ManagerID.Value != nil {
		t.Fatalf("explicit null managerId must be present with nil value, got %+v", nullCase.ManagerID)
	}

	var uuidCase UpdateUserInput
	if err := json.Unmarshal([]byte(`{"managerId":"`+uuidStr+`"}`), &uuidCase); err != nil {
		t.Fatalf("decode uuid: %v", err)
	}
	if !uuidCase.ManagerID.Present || uuidCase.ManagerID.Value == nil || uuidCase.ManagerID.Value.String() != uuidStr {
		t.Fatalf("uuid managerId must be present with value, got %+v", uuidCase.ManagerID)
	}

	var frontendPayload UpdateUserInput
	body := `{"employeeId":"YF-0001","name":"Budi","email":"budi@yummy.test","phone":"0812","role":"SALES_MANAGER","managerId":null}`
	if err := json.Unmarshal([]byte(body), &frontendPayload); err != nil {
		t.Fatalf("decode frontend payload: %v", err)
	}
	if frontendPayload.Role == nil || string(*frontendPayload.Role) != "SALES_MANAGER" {
		t.Fatalf("role not decoded, got %v", frontendPayload.Role)
	}
	if !frontendPayload.ManagerID.Present || frontendPayload.ManagerID.Value != nil {
		t.Fatalf("frontend null managerId must be present with nil value, got %+v", frontendPayload.ManagerID)
	}
}

func TestUpdateUserInputInvalidManagerUUIDRejected(t *testing.T) {
	var input UpdateUserInput
	if err := json.Unmarshal([]byte(`{"managerId":"not-a-uuid"}`), &input); err == nil {
		t.Fatal("expected error for invalid managerId UUID")
	}
}
