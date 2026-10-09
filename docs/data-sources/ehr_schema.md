---
page_title: "tcm_ehr_schema Data Source - TCM Provider"
description: |-
  Reads an EHR Schema definition from the TCM service by ID.
---

# tcm_ehr_schema (Data Source)

Reads an existing EHR Schema definition from TCM by its UUID.

## Example Usage

```hcl
data "tcm_ehr_schema" "acme" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "required_fields" {
  value = data.tcm_ehr_schema.acme.required_fields
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the EHR schema to read. |

## Attribute Reference

| Attribute          | Type         | Description |
|--------------------|--------------|-------------|
| `tenant_id`        | string       | UUID of the associated tenant. |
| `endpoint_id`      | string       | UUID of the associated EHR endpoint. |
| `required_fields`  | list(string) | List of field names required in the schema. |
| `is_disabled`      | bool         | Whether this schema is disabled. |
