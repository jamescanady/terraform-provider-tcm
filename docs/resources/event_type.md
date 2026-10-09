---
page_title: "tcm_event_type Resource - TCM Provider"
description: |-
  Manages an EventType in the TCM service.
---

# tcm_event_type (Resource)

Manages an EventType in the TCM service.

## Example Usage

```hcl
resource "tcm_event_type" "patient_admitted" {
  product_id  = tcm_product.my_product.id
  name        = "patient.admitted"
  description = "Fired when a patient is admitted"
  is_disabled = false
}
```

## Argument Reference

| Argument      | Required | Type   | Description |
|---------------|----------|--------|-------------|
| `product_id`  | yes      | string | UUID of the product. |
| `name`        | yes      | string | Unique event type name. |
| `description` | no       | string | Description of the event type. |
| `is_disabled` | no       | bool   | Whether this event type is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute          | Type   | Description |
|--------------------|--------|-------------|
| `id`               | string | UUID assigned by TCM on creation. |
| `created_by`       | string | Identity that created this event type. |
| `created`          | string | Timestamp when this event type was created. |
| `last_modified_by` | string | Identity that last modified this event type. |
| `last_modified`    | string | Timestamp when this event type was last modified. |

## Import

Import an existing event type by its TCM UUID:

```bash
terraform import tcm_event_type.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_event_type.patient_admitted 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
