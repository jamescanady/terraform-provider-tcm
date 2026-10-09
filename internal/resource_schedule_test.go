package internal

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchedule_toPayload_required(t *testing.T) {
	r := &ScheduleResource{}
	m := &ScheduleResourceModel{
		ID:                         types.StringNull(),
		TenantProductEnvironmentID: types.StringValue("tpe-001"),
		ScheduleCategoryID:         types.StringValue("cat-001"),
		Enabled:                    types.BoolValue(false),
	}
	p := r.toPayload(m)
	if p.TenantProductEnvironmentID != "tpe-001" {
		t.Errorf("expected TenantProductEnvironmentID tpe-001, got %q", p.TenantProductEnvironmentID)
	}
	if p.ScheduleCategoryID != "cat-001" {
		t.Errorf("expected ScheduleCategoryID cat-001, got %q", p.ScheduleCategoryID)
	}
	if p.Enabled {
		t.Errorf("expected Enabled false")
	}
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null")
	}
	if p.ScheduleDetails != nil {
		t.Errorf("expected ScheduleDetails nil when not set")
	}
}

func TestSchedule_toPayload_scheduleDetails(t *testing.T) {
	r := &ScheduleResource{}
	jsonStr := `{"cron":"0 * * * *"}`
	m := &ScheduleResourceModel{
		ID:                         types.StringNull(),
		TenantProductEnvironmentID: types.StringValue("tpe-001"),
		ScheduleCategoryID:         types.StringValue("cat-001"),
		Enabled:                    types.BoolValue(true),
		ScheduleDetails:            types.StringValue(jsonStr),
	}
	p := r.toPayload(m)
	if p.ScheduleDetails == nil {
		t.Fatal("expected ScheduleDetails to be non-nil")
	}
	if string(*p.ScheduleDetails) != jsonStr {
		t.Errorf("expected ScheduleDetails %q, got %q", jsonStr, string(*p.ScheduleDetails))
	}
}

func TestSchedule_applyResult(t *testing.T) {
	id := "sched-result"
	name := "My Schedule"
	desc := "schedule desc"
	jsonStr := `{"interval":60}`
	rawDetails := json.RawMessage(jsonStr)

	p := &schedulePayload{
		ID:                         &id,
		Name:                       &name,
		Description:                &desc,
		TenantProductEnvironmentID: "tpe-002",
		ScheduleCategoryID:         "cat-002",
		Enabled:                    true,
		ScheduleDetails:            &rawDetails,
	}
	m := &ScheduleResourceModel{}
	applyScheduleResult(m, p)

	if m.ID.ValueString() != "sched-result" {
		t.Errorf("expected ID sched-result, got %q", m.ID.ValueString())
	}
	if m.Name.ValueString() != "My Schedule" {
		t.Errorf("expected Name 'My Schedule', got %q", m.Name.ValueString())
	}
	if m.Description.ValueString() != "schedule desc" {
		t.Errorf("expected Description 'schedule desc', got %q", m.Description.ValueString())
	}
	if m.TenantProductEnvironmentID.ValueString() != "tpe-002" {
		t.Errorf("expected TenantProductEnvironmentID tpe-002, got %q", m.TenantProductEnvironmentID.ValueString())
	}
	if m.ScheduleCategoryID.ValueString() != "cat-002" {
		t.Errorf("expected ScheduleCategoryID cat-002, got %q", m.ScheduleCategoryID.ValueString())
	}
	if !m.Enabled.ValueBool() {
		t.Errorf("expected Enabled true")
	}
	if m.ScheduleDetails.ValueString() != jsonStr {
		t.Errorf("expected ScheduleDetails %q, got %q", jsonStr, m.ScheduleDetails.ValueString())
	}
}

func TestSchedule_applyResult_nilScheduleDetails(t *testing.T) {
	p := &schedulePayload{
		TenantProductEnvironmentID: "tpe-003",
		ScheduleCategoryID:         "cat-003",
	}
	m := &ScheduleResourceModel{}
	applyScheduleResult(m, p)
	if !m.ScheduleDetails.IsNull() {
		t.Errorf("expected ScheduleDetails null when payload field is nil")
	}
}
