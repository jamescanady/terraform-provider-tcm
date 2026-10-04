---
page_title: "tcm_system_info Resource - TCM Provider"
description: |-
  Manages a SystemInfo entry in the TCM service.
---

# tcm_system_info (Resource)

Manages a SystemInfo entry in TCM, defining connection details for a product's integration.

## Example Usage

```hcl
resource "tcm_system_info" "acme_api" {
  product_id      = tcm_product.event_engine_audit.id
  tenant_id       = tcm_tenant.acme.id
  namespace_id    = tcm_namespace.production.id
  host            = "https://api.acme.example.com"
  flow_version    = "2.0"
  base_api_path   = "/api/v2"
  o_auth_scope    = "openid profile email"
  connection_type = "Http"
  description     = "Acme production API connection"
}
```

## Argument Reference

| Argument                | Required | Type   | Description |
|-------------------------|----------|--------|-------------|
| `product_id`            | yes      | string | UUID of the product. |
| `host`                  | yes      | string | Host URL for the integration. |
| `flow_version`          | yes      | string | Flow version identifier. |
| `base_api_path`         | yes      | string | Base API path. |
| `o_auth_scope`          | yes      | string | OAuth scope string. |
| `tenant_id`             | no       | string | UUID of the tenant. |
| `namespace_id`          | no       | string | UUID of the namespace. |
| `product_environment_id` | no      | string | UUID of the product environment. |
| `flow_id`               | no       | string | Flow identifier. |
| `connection_type`       | no       | string | Connection type (e.g. `Http`). |
| `description`           | no       | string | Description. Maximum 110 characters. |
| `endpoint`              | no       | string | Specific endpoint path. |
| `is_disabled`           | no       | bool   | Whether this system info is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute        | Type   | Description |
|------------------|--------|-------------|
| `id`             | string | UUID assigned by TCM on creation. |
| `namespace_name` | string | Name of the associated namespace (read-only, populated by TCM). |
| `tenant_name`    | string | Name of the associated tenant (read-only, populated by TCM). |
| `product_name`   | string | Name of the associated product (read-only, populated by TCM). |

## Import

Import an existing system info by its TCM UUID:

```bash
terraform import tcm_system_info.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_system_info.acme_api 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
