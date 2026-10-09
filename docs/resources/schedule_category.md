---
page_title: "tcm_schedule_category Resource - TCM Provider"
description: |-
  Manages a ScheduleCategory in the TCM service.
---

# tcm_schedule_category (Resource)

Manages a ScheduleCategory in the TCM service.

## Example Usage

```hcl
resource "tcm_schedule_category" "nightly_sync" {
  code        = "NIGHTLY_SYNC"
  description = "Nightly synchronization jobs"
  is_disabled = false
}
```

## Argument Reference

| Argument      | Required | Type   | Description |
|---------------|----------|--------|-------------|
| `code`        | no       | string | Code identifier for the schedule category. |
| `description` | no       | string | Description of the schedule category. |
| `is_disabled` | no       | bool   | Whether this schedule category is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing schedule category by its TCM UUID:

```bash
terraform import tcm_schedule_category.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_schedule_category.nightly_sync 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
