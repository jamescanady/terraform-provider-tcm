package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNamespace_toPayload_defaults(t *testing.T) {
	r := &NamespaceResource{}
	m := &NamespaceResourceModel{
		ID:          types.StringNull(),
		Name:        types.StringValue("my-namespace"),
		Description: types.StringNull(),
		IsDisabled:  types.BoolValue(false),
	}

	p := r.toPayload(m)

	if p.Name != "my-namespace" {
		t.Errorf("expected Name my-namespace, got %s", p.Name)
	}
	if p.IsDisabled != false {
		t.Error("expected IsDisabled false")
	}
	if p.Description != nil {
		t.Error("expected Description nil when model Description is null")
	}
	if p.ID != "" {
		t.Errorf("expected ID empty string when model ID is null, got %s", p.ID)
	}
}

func TestNamespace_toPayload_withAllFields(t *testing.T) {
	r := &NamespaceResource{}
	desc := "a description"
	m := &NamespaceResourceModel{
		ID:          types.StringValue("ns-001"),
		Name:        types.StringValue("full-namespace"),
		Description: types.StringValue(desc),
		IsDisabled:  types.BoolValue(true),
	}

	p := r.toPayload(m)

	if p.ID != "ns-001" {
		t.Errorf("expected ID ns-001, got %s", p.ID)
	}
	if p.Name != "full-namespace" {
		t.Errorf("expected Name full-namespace, got %s", p.Name)
	}
	if p.Description == nil || *p.Description != desc {
		t.Errorf("expected Description %q, got %v", desc, p.Description)
	}
	if !p.IsDisabled {
		t.Error("expected IsDisabled true")
	}
}

func TestNamespace_toPayload_unknownIDSkipped(t *testing.T) {
	r := &NamespaceResource{}
	m := &NamespaceResourceModel{
		ID:          types.StringUnknown(),
		Name:        types.StringValue("ns-unknown"),
		Description: types.StringNull(),
		IsDisabled:  types.BoolValue(false),
	}

	p := r.toPayload(m)

	if p.ID != "" {
		t.Errorf("expected ID to be empty for unknown model ID, got %s", p.ID)
	}
}

func TestNamespace_applyNamespaceResult_allFields(t *testing.T) {
	createdBy := "user-a"
	lastModBy := "user-b"
	name := "response-ns"
	desc := "response desc"
	resp := &namespaceResponse{
		ID:             "ns-resp-001",
		Name:           &name,
		Description:    &desc,
		IsDefault:      true,
		IsDisabled:     false,
		CreatedDate:    "2024-01-01T00:00:00Z",
		CreatedBy:      &createdBy,
		LastModified:   "2024-06-01T00:00:00Z",
		LastModifiedBy: &lastModBy,
	}

	var m NamespaceResourceModel
	applyNamespaceResult(&m, resp)

	if m.ID.ValueString() != "ns-resp-001" {
		t.Errorf("expected ID ns-resp-001, got %s", m.ID.ValueString())
	}
	if m.Name.ValueString() != "response-ns" {
		t.Errorf("expected Name response-ns, got %s", m.Name.ValueString())
	}
	if m.Description.ValueString() != "response desc" {
		t.Errorf("expected Description response desc, got %s", m.Description.ValueString())
	}
	if !m.IsDefault.ValueBool() {
		t.Error("expected IsDefault true")
	}
	if m.IsDisabled.ValueBool() {
		t.Error("expected IsDisabled false")
	}
	if m.CreatedDate.ValueString() != "2024-01-01T00:00:00Z" {
		t.Errorf("expected CreatedDate 2024-01-01T00:00:00Z, got %s", m.CreatedDate.ValueString())
	}
	if m.CreatedBy.ValueString() != "user-a" {
		t.Errorf("expected CreatedBy user-a, got %s", m.CreatedBy.ValueString())
	}
	if m.LastModified.ValueString() != "2024-06-01T00:00:00Z" {
		t.Errorf("expected LastModified 2024-06-01T00:00:00Z, got %s", m.LastModified.ValueString())
	}
	if m.LastModifiedBy.ValueString() != "user-b" {
		t.Errorf("expected LastModifiedBy user-b, got %s", m.LastModifiedBy.ValueString())
	}
}

func TestNamespace_applyNamespaceResult_nullableNils(t *testing.T) {
	resp := &namespaceResponse{
		ID:             "ns-nil-001",
		Name:           nil,
		Description:    nil,
		IsDefault:      false,
		IsDisabled:     false,
		CreatedDate:    "2024-01-01T00:00:00Z",
		CreatedBy:      nil,
		LastModified:   "2024-01-01T00:00:00Z",
		LastModifiedBy: nil,
	}

	var m NamespaceResourceModel
	applyNamespaceResult(&m, resp)

	if !m.Description.IsNull() {
		t.Error("expected Description to be null when response Description is nil")
	}
	if !m.CreatedBy.IsNull() {
		t.Error("expected CreatedBy to be null when response CreatedBy is nil")
	}
	if !m.LastModifiedBy.IsNull() {
		t.Error("expected LastModifiedBy to be null when response LastModifiedBy is nil")
	}
}
