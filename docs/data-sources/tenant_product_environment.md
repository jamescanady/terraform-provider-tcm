---
page_title: "tcm_tenant_product_environment Data Source - TCM Provider"
description: |-
  Reads a TenantProductEnvironment from the TCM service by ID.
---

# tcm_tenant_product_environment (Data Source)

Reads an existing TenantProductEnvironment from TCM by its UUID.

## Example Usage

```hcl
data "tcm_tenant_product_environment" "acme_prod" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "namespace_id" {
  value = data.tcm_tenant_product_environment.acme_prod.namespace_id
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the tenant product environment to read. |

## Attribute Reference

| Attribute                | Type   | Description |
|--------------------------|--------|-------------|
| `tenant_id`              | string | UUID of the associated tenant. |
| `product_environment_id` | string | UUID of the associated product environment. |
| `namespace_id`           | string | UUID of the associated namespace. |
| `product_tenant_code`    | string | Product tenant code. |
| `product_alias`          | string | Product alias in this environment. |
| `is_disabled`            | bool   | Whether this mapping is disabled. |
