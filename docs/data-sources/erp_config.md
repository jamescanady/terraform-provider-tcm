---
page_title: "tcm_erp_config Data Source - TCM Provider"
description: |-
  Reads an ERPConfig (ERP integration configuration) from the TCM service by ID.
---

# tcm_erp_config (Data Source)

Reads an existing ERPConfig (ERP integration configuration) from TCM by its UUID.

## Example Usage

```hcl
data "tcm_erp_config" "acme_workday" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "erp_name" {
  value = data.tcm_erp_config.acme_workday.erp_name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the ERP config to read. |

## Attribute Reference

| Attribute                      | Type         | Description |
|--------------------------------|--------------|-------------|
| `erp_id`                       | string       | UUID of the associated ERP. |
| `tenant_product_environment_id` | string      | UUID of the tenant product environment. |
| `url`                          | string       | URL for the ERP integration endpoint. |
| `client_id`                    | string       | Client ID for authentication. |
| `client_secret`                | string       | Client secret for authentication. Sensitive — value is not shown in output. |
| `lookup_id`                    | string       | Lookup identifier. |
| `host_name`                    | string       | Host name for the integration. |
| `tenant_slug`                  | string       | Tenant slug used by the ERP. |
| `api_version`                  | string       | API version string. |
| `ontology_type_ids`            | list(string) | List of ontology type UUIDs. Not populated — the API does not return this value. |
| `global_tenant_code`           | string       | Global tenant code. |
| `erp_name`                     | string       | Name of the associated ERP. |
| `is_deleted`                   | bool         | Whether this ERP config has been deleted. |
