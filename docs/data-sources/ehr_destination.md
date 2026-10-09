---
page_title: "tcm_ehr_destination Data Source - TCM Provider"
description: |-
  Reads an EHR Destination from the TCM service by ID.
---

# tcm_ehr_destination (Data Source)

Reads an existing EHR Destination from TCM by its UUID.

## Example Usage

```hcl
data "tcm_ehr_destination" "acme" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "destination_name" {
  value = data.tcm_ehr_destination.acme.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the EHR destination to read. |

## Attribute Reference

| Attribute          | Type   | Description |
|--------------------|--------|-------------|
| `json_metadata_id` | string | UUID of the associated EHR JSON metadata. |
| `destination_id`   | string | UUID of the destination. |
| `name`             | string | Name of the EHR destination. |
| `description`      | string | Description of the EHR destination. |
| `is_disabled`      | bool   | Whether this destination is disabled. |
