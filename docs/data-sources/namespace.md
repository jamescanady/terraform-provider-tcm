---
page_title: "tcm_namespace Data Source - TCM Provider"
description: |-
  Reads a Namespace from the TCM service by ID.
---

# tcm_namespace (Data Source)

Reads an existing Namespace from TCM by its UUID.

## Example Usage

```hcl
data "tcm_namespace" "production" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "namespace_name" {
  value = data.tcm_namespace.production.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the namespace to read. |

## Attribute Reference

| Attribute        | Type   | Description |
|------------------|--------|-------------|
| `name`           | string | Namespace name. |
| `description`    | string | Namespace description. |
| `is_disabled`    | bool   | Whether the namespace is disabled. |
| `is_default`     | bool   | Whether this is the default namespace. |
| `created_date`   | string | UTC timestamp of when the namespace was created. |
| `created_by`     | string | Identity that created the namespace. |
| `last_modified`  | string | UTC timestamp of the last modification. |
| `last_modified_by` | string | Identity that last modified the namespace. |
