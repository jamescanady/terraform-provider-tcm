---
page_title: "tcm_tenant_product Data Source - TCM Provider"
description: |-
  Reads a TenantProduct mapping from the TCM service by ID.
---

# tcm_tenant_product (Data Source)

Reads an existing TenantProduct mapping from TCM by its UUID.

## Example Usage

```hcl
data "tcm_tenant_product" "acme_audit" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "tenant_product_code" {
  value = data.tcm_tenant_product.acme_audit.tenant_product_code
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the tenant product mapping to read. |

## Attribute Reference

| Attribute             | Type   | Description |
|-----------------------|--------|-------------|
| `tenant_id`           | string | UUID of the associated tenant. |
| `product_id`          | string | UUID of the associated product. |
| `tenant_product_code` | string | Tenant-specific product code. |
| `is_disabled`         | bool   | Whether the mapping is disabled. |
