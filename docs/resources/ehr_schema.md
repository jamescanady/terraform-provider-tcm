---
page_title: "tcm_ehr_schema Resource - TCM Provider"
description: |-
  Manages an EHR Schema definition in the TCM service.
---

# tcm_ehr_schema (Resource)

Manages an EHR Schema definition in the TCM service.

## Example Usage

```hcl
resource "tcm_ehr_schema" "acme" {
  tenant_id       = tcm_tenant.acme.id
  endpoint_id     = tcm_ehr_endpoint.adt.id
  required_fields = ["patientId", "facilityCode", "eventType"]
}
```

## Argument Reference

| Argument          | Required | Type         | Description |
|-------------------|----------|--------------|-------------|
| `tenant_id`       | no       | string       | UUID of the associated tenant. |
| `endpoint_id`     | no       | string       | UUID of the associated EHR endpoint. |
| `required_fields` | no       | list(string) | List of field names required in the schema. |
| `is_disabled`     | no       | bool         | Whether this schema is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing EHR schema by its TCM UUID:

```bash
terraform import tcm_ehr_schema.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_ehr_schema.acme 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
