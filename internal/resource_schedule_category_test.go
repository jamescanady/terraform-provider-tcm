package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestScheduleCategory_toPayload_defaults(t *testing.T) {
	r := &ScheduleCategoryResource{}
	m := &ScheduleCategoryResourceModel{
		ID:         types.StringNull(),
		IsDisabled: types.BoolValue(false),
	}
	p := r.toPayload(m)
	if p.IsDisabled {
		t.Errorf("expected IsDisabled false")
	}
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null")
	}
	if p.Code != nil {
		t.Errorf("expected Code nil when not set")
	}
	if p.Description != nil {
		t.Errorf("expected Description nil when not set")
	}
}

func TestScheduleCategory_toPayload_withFields(t *testing.T) {
	r := &ScheduleCategoryResource{}
	m := &ScheduleCategoryResourceModel{
		ID:          types.StringValue("sc-001"),
		Code:        types.StringValue("CAT001"),
		Description: types.StringValue("My category"),
		IsDisabled:  types.BoolValue(true),
	}
	p := r.toPayload(m)
	if p.ID == nil || *p.ID != "sc-001" {
		t.Errorf("expected ID sc-001")
	}
	if p.Code == nil || *p.Code != "CAT001" {
		t.Errorf("expected Code CAT001")
	}
	if p.Description == nil || *p.Description != "My category" {
		t.Errorf("expected Description 'My category'")
	}
	if !p.IsDisabled {
		t.Errorf("expected IsDisabled true")
	}
}

func TestScheduleCategory_applyResult(t *testing.T) {
	id := "sc-result"
	code := "RESULT-CAT"
	desc := "result description"
	p := &scheduleCategoryPayload{
		ID:          &id,
		Code:        &code,
		Description: &desc,
		IsDisabled:  true,
	}
	m := &ScheduleCategoryResourceModel{}
	applyScheduleCategoryResult(m, p)

	if m.ID.ValueString() != "sc-result" {
		t.Errorf("expected ID sc-result, got %q", m.ID.ValueString())
	}
	if m.Code.ValueString() != "RESULT-CAT" {
		t.Errorf("expected Code RESULT-CAT, got %q", m.Code.ValueString())
	}
	if m.Description.ValueString() != "result description" {
		t.Errorf("expected Description 'result description', got %q", m.Description.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
}
