---
page_title: "tcm_event_type Data Source - TCM Provider"
description: |-
  Reads an EventType from the TCM service by ID.
---

# tcm_event_type (Data Source)

Reads an existing EventType from TCM by its UUID.

## Example Usage

```hcl
data "tcm_event_type" "lookup" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "event_type_name" {
  value = data.tcm_event_type.lookup.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the event type to read. |

## Attribute Reference

| Attribute          | Type   | Description |
|--------------------|--------|-------------|
| `product_id`       | string | UUID of the associated product. |
| `name`             | string | Unique event type name. |
| `description`      | string | Description of the event type. |
| `is_disabled`      | bool   | Whether this event type is disabled. |
| `created_by`       | string | Identity that created this event type. |
| `created`          | string | Timestamp when this event type was created. |
| `last_modified_by` | string | Identity that last modified this event type. |
| `last_modified`    | string | Timestamp when this event type was last modified. |
