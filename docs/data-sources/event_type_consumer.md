---
page_title: "tcm_event_type_consumer Data Source - TCM Provider"
description: |-
  Reads an EventType-to-EventConsumer mapping from the TCM service by ID.
---

# tcm_event_type_consumer (Data Source)

Reads an existing EventType-to-EventConsumer mapping from TCM by its UUID.

## Example Usage

```hcl
data "tcm_event_type_consumer" "lookup" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "consumer_endpoint" {
  value = data.tcm_event_type_consumer.lookup.consumer_endpoint
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the event type consumer mapping to read. |

## Attribute Reference

| Attribute                      | Type   | Description |
|--------------------------------|--------|-------------|
| `event_type_id`                | string | UUID of the associated event type. |
| `event_consumer_id`            | string | UUID of the associated event consumer. |
| `tenant_product_environment_id`| string | UUID of the associated tenant product environment. |
| `is_disabled`                  | bool   | Whether this subscription is disabled. |
| `consumer_name`                | string | Name of the associated event consumer. |
| `consumer_description`         | string | Description of the associated event consumer. |
| `consumer_endpoint`            | string | Endpoint URL of the associated event consumer. |
| `consumer_authorization_type`  | string | Authorization type of the associated event consumer. |
| `event_type_name`              | string | Name of the associated event type. |
| `tenant_id`                    | string | UUID of the associated tenant. |
| `tenant_name`                  | string | Name of the associated tenant. |
| `product_id`                   | string | UUID of the associated product. |
| `product_name`                 | string | Name of the associated product. |
| `product_string`               | string | String identifier of the associated product. |
| `environment_name`             | string | Name of the associated environment. |
| `created_by`                   | string | Identity that created this subscription. |
| `created`                      | string | Timestamp when this subscription was created. |
| `last_modified_by`             | string | Identity that last modified this subscription. |
| `last_modified`                | string | Timestamp when this subscription was last modified. |
