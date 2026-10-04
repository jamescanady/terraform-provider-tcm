---
page_title: "tcm_tenant Data Source - TCM Provider"
description: |-
  Reads a Tenant from the TCM service by ID.
---

# tcm_tenant (Data Source)

Reads an existing Tenant from TCM by its UUID.

## Example Usage

```hcl
data "tcm_tenant" "acme" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "tenant_name" {
  value = data.tcm_tenant.acme.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the tenant to read. |

## Attribute Reference

| Attribute            | Type   | Description |
|----------------------|--------|-------------|
| `name`               | string | Tenant name. |
| `description`        | string | Tenant description. |
| `global_tenant_code` | string | Global tenant code. |
| `tenant_short_code`  | string | Tenant short code. |
| `is_disabled`        | bool   | Whether the tenant is disabled. |
