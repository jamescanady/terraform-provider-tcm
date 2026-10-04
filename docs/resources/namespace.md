---
page_title: "tcm_namespace Resource - TCM Provider"
description: |-
  Manages a Namespace in the TCM service.
---

# tcm_namespace (Resource)

Manages a Namespace in TCM.

## Example Usage

```hcl
resource "tcm_namespace" "production" {
  name        = "Production"
  description = "Production namespace"
}
```

## Argument Reference

| Argument      | Required | Type   | Description |
|---------------|----------|--------|-------------|
| `name`        | yes      | string | Namespace name. |
| `description` | no       | string | Namespace description. |
| `is_disabled` | no       | bool   | Whether the namespace is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute        | Type   | Description |
|------------------|--------|-------------|
| `id`             | string | UUID assigned by TCM on creation. |
| `is_default`     | bool   | Whether this is the default namespace. |
| `created_date`   | string | UTC timestamp of when the namespace was created. |
| `created_by`     | string | Identity that created the namespace. |
| `last_modified`  | string | UTC timestamp of the last modification. |
| `last_modified_by` | string | Identity that last modified the namespace. |

## Import

Import an existing namespace by its TCM UUID:

```bash
terraform import tcm_namespace.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_namespace.production 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
