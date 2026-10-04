---
page_title: "tcm_tenant_product Resource - TCM Provider"
description: |-
  Manages a TenantProduct mapping in the TCM service.
---

# tcm_tenant_product (Resource)

Manages a TenantProduct mapping in TCM, associating a tenant with a product.

## Example Usage

```hcl
resource "tcm_tenant_product" "acme_audit" {
  tenant_id           = tcm_tenant.acme.id
  product_id          = tcm_product.event_engine_audit.id
  tenant_product_code = "ACME_AUDIT"
}
```

## Argument Reference

| Argument             | Required | Type   | Description |
|----------------------|----------|--------|-------------|
| `tenant_id`          | yes      | string | UUID of the tenant. **Cannot be changed after creation** — any change forces a destroy and recreate. |
| `product_id`         | yes      | string | UUID of the product. **Cannot be changed after creation** — any change forces a destroy and recreate. |
| `tenant_product_code` | no      | string | Optional tenant-specific product code. |
| `is_disabled`        | no       | bool   | Whether the mapping is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing tenant product mapping by its TCM UUID:

```bash
terraform import tcm_tenant_product.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_tenant_product.acme_audit 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
