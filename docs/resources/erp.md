---
page_title: "tcm_erp Resource - TCM Provider"
description: |-
  Manages an ERP record in the TCM service.
---

# tcm_erp (Resource)

Manages an ERP record in the TCM service.

## Example Usage

```hcl
resource "tcm_erp" "workday" {
  name        = "Workday"
  description = "Workday HR integration"
  version     = "38.0"
}
```

## Argument Reference

| Argument      | Required | Type   | Description |
|---------------|----------|--------|-------------|
| `name`        | no       | string | Name of the ERP. Maximum 50 characters. |
| `description` | no       | string | Description of the ERP. Maximum 110 characters. |
| `version`     | no       | string | Version identifier. Maximum 10 characters. |

## Attribute Reference

| Attribute            | Type   | Description |
|----------------------|--------|-------------|
| `id`                 | string | UUID assigned by TCM on creation. |
| `is_deleted`         | bool   | Whether this ERP has been deleted (read-only, populated by TCM). |
| `created_by`         | string | Identity that created the record (read-only, populated by TCM). |
| `created_date`       | string | Timestamp when the record was created (read-only, populated by TCM). |
| `last_modified_by`   | string | Identity that last modified the record (read-only, populated by TCM). |
| `last_modified_date` | string | Timestamp of the last modification (read-only, populated by TCM). |

## Import

Import an existing ERP by its TCM UUID:

```bash
terraform import tcm_erp.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_erp.workday 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
