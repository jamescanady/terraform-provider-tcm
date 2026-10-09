---
page_title: "tcm_ehr_tenant Resource - TCM Provider"
description: |-
  Manages an EHR Tenant configuration in the TCM service.
---

# tcm_ehr_tenant (Resource)

Manages an EHR Tenant configuration in the TCM service.

## Example Usage

```hcl
resource "tcm_ehr_tenant" "acme" {
  tenant_id            = tcm_tenant.acme.id
  name                 = "Acme Health System"
  redox_ehr_identifier = "acme-ehr"
  symplr_url_slug      = "acme"
  redox_environment    = "production"
  description          = "Acme production EHR tenant"
}
```

## Argument Reference

| Argument               | Required | Type   | Description |
|------------------------|----------|--------|-------------|
| `name`                 | yes      | string | Name of the EHR tenant. |
| `redox_ehr_identifier` | yes      | string | Redox EHR identifier for this tenant. |
| `symplr_url_slug`      | yes      | string | symplr URL slug for this tenant. |
| `redox_environment`    | yes      | string | Redox environment (e.g. `production`). |
| `tenant_id`            | no       | string | UUID of the associated tenant. |
| `description`          | no       | string | Description of the EHR tenant. |
| `is_disabled`          | no       | bool   | Whether this EHR tenant is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing EHR tenant by its TCM UUID:

```bash
terraform import tcm_ehr_tenant.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_ehr_tenant.acme 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
