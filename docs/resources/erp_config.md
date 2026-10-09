---
page_title: "tcm_erp_config Resource - TCM Provider"
description: |-
  Manages an ERPConfig (ERP integration configuration) in the TCM service.
---

# tcm_erp_config (Resource)

Manages an ERPConfig (ERP integration configuration) in the TCM service.

## Example Usage

```hcl
resource "tcm_erp_config" "acme_workday" {
  erp_id                        = tcm_erp.workday.id
  tenant_product_environment_id = tcm_tenant_product_environment.acme_prod.id
  url                           = "https://workday.acme.example.com"
  client_id                     = "acme-client"
  client_secret                 = var.workday_secret
  api_version                   = "v38"
  ontology_type_ids             = [tcm_ontology_type.employee.id]
}
```

## Argument Reference

| Argument                       | Required | Type         | Description |
|--------------------------------|----------|--------------|-------------|
| `erp_id`                       | no       | string       | UUID of the associated ERP. |
| `tenant_product_environment_id` | no      | string       | UUID of the tenant product environment. |
| `url`                          | no       | string       | URL for the ERP integration endpoint. |
| `client_id`                    | no       | string       | Client ID for authentication. |
| `client_secret`                | no       | string       | Client secret for authentication. Sensitive — will not be shown in plan output. |
| `lookup_id`                    | no       | string       | Lookup identifier. |
| `host_name`                    | no       | string       | Host name for the integration. |
| `tenant_slug`                  | no       | string       | Tenant slug used by the ERP. |
| `api_version`                  | no       | string       | API version string. |
| `ontology_type_ids`            | no       | list(string) | List of ontology type UUIDs to associate. Write-only — the API does not return this value, so it will not be populated on import. |

> **Note:** `client_secret` is sensitive and will not be shown in plan output.

> **Note:** `ontology_type_ids` is write-only. The API does not return it, so the value cannot be recovered via `terraform import`.

## Attribute Reference

| Attribute            | Type   | Description |
|----------------------|--------|-------------|
| `id`                 | string | UUID assigned by TCM on creation. |
| `global_tenant_code` | string | Global tenant code (read-only, populated by TCM). |
| `erp_name`           | string | Name of the associated ERP (read-only, populated by TCM). |
| `is_deleted`         | bool   | Whether this ERP config has been deleted (read-only, populated by TCM). |

## Import

Import an existing ERP config by its TCM UUID:

```bash
terraform import tcm_erp_config.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_erp_config.acme_workday 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
