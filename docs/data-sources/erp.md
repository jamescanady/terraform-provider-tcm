---
page_title: "tcm_erp Data Source - TCM Provider"
description: |-
  Reads an ERP record from the TCM service by ID.
---

# tcm_erp (Data Source)

Reads an existing ERP record from TCM by its UUID.

## Example Usage

```hcl
data "tcm_erp" "workday" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "erp_name" {
  value = data.tcm_erp.workday.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the ERP record to read. |

## Attribute Reference

| Attribute            | Type   | Description |
|----------------------|--------|-------------|
| `name`               | string | Name of the ERP. |
| `description`        | string | Description of the ERP. |
| `version`            | string | Version identifier. |
| `is_deleted`         | bool   | Whether this ERP has been deleted. |
| `created_by`         | string | Identity that created the record. |
| `created_date`       | string | Timestamp when the record was created. |
| `last_modified_by`   | string | Identity that last modified the record. |
| `last_modified_date` | string | Timestamp of the last modification. |
