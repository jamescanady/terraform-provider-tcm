---
page_title: "tcm_ehr_json_metadata Data Source - TCM Provider"
description: |-
  Reads EHR JSON Metadata from the TCM service by ID.
---

# tcm_ehr_json_metadata (Data Source)

Reads an existing EHR JSON Metadata record from TCM by its UUID.

## Example Usage

```hcl
data "tcm_ehr_json_metadata" "acme" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "facility_code" {
  value = data.tcm_ehr_json_metadata.acme.facility_code
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the EHR JSON metadata record to read. |

## Attribute Reference

| Attribute       | Type   | Description |
|-----------------|--------|-------------|
| `tenant_id`     | string | UUID of the associated tenant. |
| `endpoint_id`   | string | UUID of the associated EHR endpoint. |
| `facility_code` | string | Facility code for this JSON metadata record. |
| `is_disabled`   | bool   | Whether this JSON metadata record is disabled. |
