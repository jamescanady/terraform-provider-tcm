---
page_title: "tcm_schedule Resource - TCM Provider"
description: |-
  Manages a Schedule in the TCM service.
---

# tcm_schedule (Resource)

Manages a Schedule in the TCM service.

## Example Usage

```hcl
resource "tcm_schedule" "nightly_sync" {
  name                          = "nightly-sync"
  tenant_product_environment_id = tcm_tenant_product_environment.acme_prod.id
  schedule_category_id          = tcm_schedule_category.sync.id
  enabled                       = true
  schedule_details              = jsonencode({ cron = "0 2 * * *" })
}
```

## Argument Reference

| Argument                         | Required | Type   | Description |
|----------------------------------|----------|--------|-------------|
| `tenant_product_environment_id`  | yes      | string | UUID of the tenant product environment. |
| `schedule_category_id`           | yes      | string | UUID of the schedule category. |
| `name`                           | no       | string | Name of the schedule. |
| `description`                    | no       | string | Description of the schedule. |
| `enabled`                        | no       | bool   | Whether this schedule is enabled. Defaults to `false`. |
| `schedule_details`               | no       | string | JSON-encoded object containing schedule configuration (e.g. cron expression). |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing schedule by its TCM UUID:

```bash
terraform import tcm_schedule.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_schedule.nightly_sync 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
