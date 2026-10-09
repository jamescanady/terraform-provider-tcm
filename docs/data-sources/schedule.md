---
page_title: "tcm_schedule Data Source - TCM Provider"
description: |-
  Reads a Schedule from the TCM service by ID.
---

# tcm_schedule (Data Source)

Reads an existing Schedule from TCM by its UUID.

## Example Usage

```hcl
data "tcm_schedule" "lookup" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "schedule_name" {
  value = data.tcm_schedule.lookup.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the schedule to read. |

## Attribute Reference

| Attribute                        | Type   | Description |
|----------------------------------|--------|-------------|
| `name`                           | string | Name of the schedule. |
| `description`                    | string | Description of the schedule. |
| `tenant_product_environment_id`  | string | UUID of the associated tenant product environment. |
| `schedule_category_id`           | string | UUID of the associated schedule category. |
| `enabled`                        | bool   | Whether this schedule is enabled. |
| `schedule_details`               | string | JSON-encoded object containing schedule configuration. |
