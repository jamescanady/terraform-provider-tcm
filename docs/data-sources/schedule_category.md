---
page_title: "tcm_schedule_category Data Source - TCM Provider"
description: |-
  Reads a ScheduleCategory from the TCM service by ID.
---

# tcm_schedule_category (Data Source)

Reads an existing ScheduleCategory from TCM by its UUID.

## Example Usage

```hcl
data "tcm_schedule_category" "lookup" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "category_code" {
  value = data.tcm_schedule_category.lookup.code
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the schedule category to read. |

## Attribute Reference

| Attribute     | Type   | Description |
|---------------|--------|-------------|
| `code`        | string | Code identifier for the schedule category. |
| `description` | string | Description of the schedule category. |
| `is_disabled` | bool   | Whether this schedule category is disabled. |
