---
page_title: "tcm_product_environment Resource - TCM Provider"
description: |-
  Manages a ProductEnvironment in the TCM service.
---

# tcm_product_environment (Resource)

Manages a ProductEnvironment in TCM. A ProductEnvironment represents a named deployment environment (e.g. `dev`, `staging`, `production`) associated with a specific Product.

## Example Usage

```hcl
resource "tcm_product_environment" "prod" {
  product_id = tcm_product.my_product.id
  name       = "production"
}
```

## Argument Reference

| Argument     | Required | Type   | Description |
|--------------|----------|--------|-------------|
| `product_id` | yes      | string | UUID of the product this environment belongs to. **Cannot be changed after creation** — any change forces a destroy and recreate. |
| `name`       | no       | string | Environment name (e.g. `production`, `staging`). |
| `is_disabled`| no       | bool   | Whether the product environment is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing product environment by its TCM UUID:

```bash
terraform import tcm_product_environment.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_product_environment.prod 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
