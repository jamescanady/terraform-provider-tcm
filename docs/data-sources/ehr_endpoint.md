---
page_title: "tcm_ehr_endpoint Data Source - TCM Provider"
description: |-
  Reads an EHR Endpoint from the TCM service by ID.
---

# tcm_ehr_endpoint (Data Source)

Reads an existing EHR Endpoint from TCM by its UUID.

## Example Usage

```hcl
data "tcm_ehr_endpoint" "adt" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "endpoint_name" {
  value = data.tcm_ehr_endpoint.adt.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the EHR endpoint to read. |

## Attribute Reference

| Attribute     | Type   | Description |
|---------------|--------|-------------|
| `name`        | string | Name of the EHR endpoint. |
| `description` | string | Description of the EHR endpoint. |
| `is_disabled` | bool   | Whether this endpoint is disabled. |
