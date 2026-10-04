---
page_title: "tcm_tenant_product_environment Resource - TCM Provider"
description: |-
  Manages a TenantProductEnvironment in the TCM service.
---

# tcm_tenant_product_environment (Resource)

Manages a TenantProductEnvironment in TCM, linking a tenant to a specific product environment within an optional namespace.

## Example Usage

```hcl
resource "tcm_tenant_product_environment" "acme_prod" {
  tenant_id              = tcm_tenant.acme.id
  product_environment_id = "a1b2c3d4-0000-0000-0000-000000000001"
  namespace_id           = tcm_namespace.production.id
  product_tenant_code    = "ACME_PROD"
  product_alias          = "acme-production"
}
```

## Argument Reference

| Argument                | Required | Type   | Description |
|-------------------------|----------|--------|-------------|
| `tenant_id`             | yes      | string | UUID of the tenant. **Cannot be changed after creation** — any change forces a destroy and recreate. |
| `product_environment_id` | yes     | string | UUID of the product environment. **Cannot be changed after creation** — any change forces a destroy and recreate. |
| `namespace_id`          | no       | string | UUID of the namespace to associate. |
| `product_tenant_code`   | no       | string | Product tenant code. |
| `product_alias`         | no       | string | Alias for the product in this environment. |
| `is_disabled`           | no       | bool   | Whether this mapping is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing tenant product environment by its TCM UUID:

```bash
terraform import tcm_tenant_product_environment.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_tenant_product_environment.acme_prod 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
