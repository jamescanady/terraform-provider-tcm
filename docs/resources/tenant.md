---
page_title: "tcm_tenant Resource - TCM Provider"
description: |-
  Manages a Tenant in the TCM service.
---

# tcm_tenant (Resource)

Manages a Tenant in TCM.

## Example Usage

```hcl
resource "tcm_tenant" "acme" {
  name               = "Acme Corp"
  description        = "Acme Corporation tenant"
  global_tenant_code = "ACME_GLOBAL"
  tenant_short_code  = "acme"
}
```

## Argument Reference

| Argument             | Required | Type   | Description |
|----------------------|----------|--------|-------------|
| `name`               | yes      | string | Tenant name. |
| `description`        | yes      | string | Tenant description. Maximum 110 characters. |
| `global_tenant_code` | yes      | string | Global tenant code. |
| `tenant_short_code`  | yes      | string | Tenant short code (lowercase alphanumeric and underscore). **Cannot be changed after creation** — any change forces a destroy and recreate. |
| `is_disabled`        | no       | bool   | Whether the tenant is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing tenant by its TCM UUID:

```bash
terraform import tcm_tenant.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_tenant.acme 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
