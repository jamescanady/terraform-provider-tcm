---
page_title: "tcm_event_consumer Resource - TCM Provider"
description: |-
  Manages an EventConsumer (webhook subscriber) in the TCM service.
---

# tcm_event_consumer (Resource)

Manages an EventConsumer (webhook subscriber) in the TCM service.

## Example Usage

```hcl
resource "tcm_event_consumer" "acme_webhook" {
  name                     = "acme-webhook"
  endpoint                 = "https://acme.example.com/events"
  authorization_type       = "API_KEY"
  authorization_parameters = jsonencode({ apiKey = "secret" })
  is_disabled              = false
}
```

## Argument Reference

| Argument                    | Required | Type   | Description |
|-----------------------------|----------|--------|-------------|
| `name`                      | yes      | string | Name of the event consumer. |
| `endpoint`                  | yes      | string | URL to deliver events to. |
| `tenant_id`                 | no       | string | UUID of the tenant. |
| `description`               | no       | string | Description of the event consumer. |
| `authorization_type`        | no       | string | Authorization type. One of `API_KEY`, `BASIC`, or `OAUTH_CLIENT_CREDENTIALS`. |
| `authorization_parameters`  | no       | string | JSON-encoded object containing authorization credentials. |
| `is_disabled`               | no       | bool   | Whether this event consumer is disabled. Defaults to `false`. |

## Attribute Reference

| Attribute          | Type   | Description |
|--------------------|--------|-------------|
| `id`               | string | UUID assigned by TCM on creation. |
| `last_sync_date`   | string | Timestamp of the last synchronization attempt. |
| `last_sync_message`| string | Message from the last synchronization attempt. |
| `created_by`       | string | Identity that created this event consumer. |
| `created`          | string | Timestamp when this event consumer was created. |
| `last_modified_by` | string | Identity that last modified this event consumer. |
| `last_modified`    | string | Timestamp when this event consumer was last modified. |
| `version`          | number | Version number of this event consumer. |

## Import

Import an existing event consumer by its TCM UUID:

```bash
terraform import tcm_event_consumer.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_event_consumer.acme_webhook 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
