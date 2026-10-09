---
page_title: "tcm_ehr_destination Resource - TCM Provider"
description: |-
  Manages an EHR Destination in the TCM service.
---

# tcm_ehr_destination (Resource)

Manages an EHR Destination in the TCM service.

## Example Usage

```hcl
resource "tcm_ehr_destination" "acme" {
  json_metadata_id = tcm_ehr_json_metadata.acme.id
  destination_id   = "dest-uuid"
  name             = "Acme Destination"
  description      = "Primary EHR destination for Acme"
}
```

## Argument Reference

| Argument          | Required | Type   | Description |
|-------------------|----------|--------|-------------|
| `json_metadata_id` | no      | string | UUID of the associated EHR JSON metadata. |
| `destination_id`  | no       | string | UUID of the destination. |
| `name`            | no       | string | Name of the EHR destination. |
| `description`     | no       | string | Description of the EHR destination. |
| `is_disabled`     | no       | bool   | Whether this destination is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing EHR destination by its TCM UUID:

```bash
terraform import tcm_ehr_destination.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_ehr_destination.acme 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
