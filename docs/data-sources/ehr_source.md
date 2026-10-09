---
page_title: "tcm_ehr_source Data Source - TCM Provider"
description: |-
  Reads an EHR Source from the TCM service by ID.
---

# tcm_ehr_source (Data Source)

Reads an existing EHR Source from TCM by its UUID.

## Example Usage

```hcl
data "tcm_ehr_source" "acme" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "source_name" {
  value = data.tcm_ehr_source.acme.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the EHR source to read. |

## Attribute Reference

| Attribute     | Type   | Description |
|---------------|--------|-------------|
| `tenant_id`   | string | UUID of the associated tenant. |
| `source_id`   | string | UUID of the source. |
| `name`        | string | Name of the EHR source. |
| `description` | string | Description of the EHR source. |
| `is_disabled` | bool   | Whether this source is disabled. |
