---
page_title: "tcm_ehr_source Resource - TCM Provider"
description: |-
  Manages an EHR Source in the TCM service.
---

# tcm_ehr_source (Resource)

Manages an EHR Source in the TCM service.

## Example Usage

```hcl
resource "tcm_ehr_source" "acme" {
  tenant_id   = tcm_tenant.acme.id
  source_id   = "src-uuid"
  name        = "Acme EHR Source"
  description = "Primary EHR source for Acme"
}
```

## Argument Reference

| Argument      | Required | Type   | Description |
|---------------|----------|--------|-------------|
| `tenant_id`   | no       | string | UUID of the associated tenant. |
| `source_id`   | no       | string | UUID of the source. |
| `name`        | no       | string | Name of the EHR source. |
| `description` | no       | string | Description of the EHR source. |
| `is_disabled` | no       | bool   | Whether this source is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing EHR source by its TCM UUID:

```bash
terraform import tcm_ehr_source.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_ehr_source.acme 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
