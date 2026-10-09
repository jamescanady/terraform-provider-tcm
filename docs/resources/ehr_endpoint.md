---
page_title: "tcm_ehr_endpoint Resource - TCM Provider"
description: |-
  Manages an EHR Endpoint in the TCM service.
---

# tcm_ehr_endpoint (Resource)

Manages an EHR Endpoint in the TCM service.

## Example Usage

```hcl
resource "tcm_ehr_endpoint" "adt" {
  name        = "ADT-Feed"
  description = "ADT event feed endpoint"
}
```

## Argument Reference

| Argument      | Required | Type   | Description |
|---------------|----------|--------|-------------|
| `name`        | yes      | string | Name of the EHR endpoint. |
| `description` | no       | string | Description of the EHR endpoint. |
| `is_disabled` | no       | bool   | Whether this endpoint is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing EHR endpoint by its TCM UUID:

```bash
terraform import tcm_ehr_endpoint.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_ehr_endpoint.adt 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
