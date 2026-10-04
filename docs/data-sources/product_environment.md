---
page_title: "tcm_product_environment Data Source - TCM Provider"
description: |-
  Reads a ProductEnvironment from the TCM service by ID.
---

# tcm_product_environment (Data Source)

Reads an existing ProductEnvironment from TCM by its UUID.

## Example Usage

```hcl
data "tcm_product_environment" "prod" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "product_environment_name" {
  value = data.tcm_product_environment.prod.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the product environment to read. |

## Attribute Reference

| Attribute    | Type   | Description |
|--------------|--------|-------------|
| `product_id` | string | UUID of the associated product. |
| `name`       | string | Environment name. |
| `is_disabled`| bool   | Whether the product environment is disabled. |
