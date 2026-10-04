---
page_title: "tcm_system_info Data Source - TCM Provider"
description: |-
  Reads a SystemInfo entry from the TCM service by ID.
---

# tcm_system_info (Data Source)

Reads an existing SystemInfo entry from TCM by its UUID.

## Example Usage

```hcl
data "tcm_system_info" "acme_api" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "host" {
  value = data.tcm_system_info.acme_api.host
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the system info to read. |

## Attribute Reference

| Attribute                | Type   | Description |
|--------------------------|--------|-------------|
| `product_id`             | string | UUID of the associated product. |
| `tenant_id`              | string | UUID of the associated tenant. |
| `namespace_id`           | string | UUID of the associated namespace. |
| `product_environment_id` | string | UUID of the associated product environment. |
| `flow_id`                | string | Flow identifier. |
| `connection_type`        | string | Connection type (e.g. `Http`). |
| `description`            | string | Description. |
| `host`                   | string | Host URL. |
| `flow_version`           | string | Flow version. |
| `base_api_path`          | string | Base API path. |
| `endpoint`               | string | Specific endpoint path. |
| `o_auth_scope`           | string | OAuth scope string. |
| `is_disabled`            | bool   | Whether this system info is disabled. |
| `namespace_name`         | string | Name of the associated namespace. |
| `tenant_name`            | string | Name of the associated tenant. |
| `product_name`           | string | Name of the associated product. |
