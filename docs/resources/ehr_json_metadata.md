---
page_title: "tcm_ehr_json_metadata Resource - TCM Provider"
description: |-
  Manages EHR JSON Metadata in the TCM service.
---

# tcm_ehr_json_metadata (Resource)

Manages EHR JSON Metadata in the TCM service.

## Example Usage

```hcl
resource "tcm_ehr_json_metadata" "acme" {
  tenant_id     = tcm_tenant.acme.id
  endpoint_id   = tcm_ehr_endpoint.adt.id
  facility_code = "ACM001"
}
```

## Argument Reference

| Argument        | Required | Type   | Description |
|-----------------|----------|--------|-------------|
| `tenant_id`     | no       | string | UUID of the associated tenant. |
| `endpoint_id`   | no       | string | UUID of the associated EHR endpoint. |
| `facility_code` | no       | string | Facility code for this JSON metadata record. |
| `is_disabled`   | no       | bool   | Whether this JSON metadata record is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing EHR JSON metadata record by its TCM UUID:

```bash
terraform import tcm_ehr_json_metadata.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_ehr_json_metadata.acme 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
