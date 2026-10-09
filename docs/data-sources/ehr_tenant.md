---
page_title: "tcm_ehr_tenant Data Source - TCM Provider"
description: |-
  Reads an EHR Tenant configuration from the TCM service by ID.
---

# tcm_ehr_tenant (Data Source)

Reads an existing EHR Tenant configuration from TCM by its UUID.

## Example Usage

```hcl
data "tcm_ehr_tenant" "acme" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "ehr_tenant_name" {
  value = data.tcm_ehr_tenant.acme.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the EHR tenant to read. |

## Attribute Reference

| Attribute               | Type   | Description |
|-------------------------|--------|-------------|
| `tenant_id`             | string | UUID of the associated tenant. |
| `name`                  | string | Name of the EHR tenant. |
| `redox_ehr_identifier`  | string | Redox EHR identifier for this tenant. |
| `symplr_url_slug`       | string | symplr URL slug for this tenant. |
| `redox_environment`     | string | Redox environment. |
| `description`           | string | Description of the EHR tenant. |
| `is_disabled`           | bool   | Whether this EHR tenant is disabled. |
