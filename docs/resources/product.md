---
page_title: "tcm_product Resource - TCM Provider"
description: |-
  Manages a Product entry in the TCM service.
---

# tcm_product (Resource)

Manages a Product entry in TCM.

> **Note:** TCM has no hard-delete endpoint for products. Running `terraform destroy` will **soft-delete** the product by setting `isDisabled = true` via `PUT /v1/Product/{id}`. The record remains in TCM.

## Example Usage

```hcl
resource "tcm_product" "event_engine_audit" {
  name        = "EventEngineAudit"
  description = "Event Engine Audit product"
}
```

## Argument Reference

| Argument      | Required | Type   | Description |
|---------------|----------|--------|-------------|
| `name`        | yes      | string | Product name. |
| `description` | yes      | string | Product description. Maximum 110 characters. |
| `is_disabled` | no       | bool   | Whether the product is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing product by its TCM UUID:

```bash
terraform import tcm_product.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_product.event_engine_audit 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```

### Finding the UUID

If you only know the product name, use the TCM `find` endpoint:

```bash
curl -s \
  -H "Authorization: Bearer $TCM_TOKEN" \
  "https://stable-platform.symplr.com/ce-platform-tenant-configuration-service/v1/Product/find/EventEngineAudit" \
  | jq '.[0].id'
```

> **Important:** If you skip the import and run `terraform apply` against a product that already exists, TCM will return a `409 Conflict`. Always import first when the product is pre-existing.
