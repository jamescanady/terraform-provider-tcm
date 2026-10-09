---
page_title: "tcm_event_type_consumer Resource - TCM Provider"
description: |-
  Manages the mapping between an EventType and an EventConsumer (subscription) in the TCM service.
---

# tcm_event_type_consumer (Resource)

Manages the mapping between an EventType and an EventConsumer (subscription) in the TCM service.

## Example Usage

```hcl
resource "tcm_event_type_consumer" "admitted_webhook" {
  event_type_id                  = tcm_event_type.admitted.id
  event_consumer_id              = tcm_event_consumer.acme_webhook.id
  tenant_product_environment_id  = tcm_tenant_product_environment.acme_prod.id
}
```

## Argument Reference

| Argument                         | Required | Type   | Description |
|----------------------------------|----------|--------|-------------|
| `event_type_id`                  | no       | string | UUID of the event type. |
| `event_consumer_id`              | no       | string | UUID of the event consumer. |
| `tenant_product_environment_id`  | no       | string | UUID of the tenant product environment. |
| `is_disabled`                    | no       | bool   | Whether this subscription is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute                      | Type   | Description |
|--------------------------------|--------|-------------|
| `id`                           | string | UUID assigned by TCM on creation. |
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

## Import

Import an existing event type consumer mapping by its TCM UUID:

```bash
terraform import tcm_event_type_consumer.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_event_type_consumer.admitted_webhook 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
