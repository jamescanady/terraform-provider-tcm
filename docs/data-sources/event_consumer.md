---
page_title: "tcm_event_consumer Data Source - TCM Provider"
description: |-
  Reads an EventConsumer from the TCM service by ID.
---

# tcm_event_consumer (Data Source)

Reads an existing EventConsumer from TCM by its UUID.

## Example Usage

```hcl
data "tcm_event_consumer" "lookup" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "consumer_endpoint" {
  value = data.tcm_event_consumer.lookup.endpoint
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the event consumer to read. |

## Attribute Reference

| Attribute                   | Type   | Description |
|-----------------------------|--------|-------------|
| `name`                      | string | Name of the event consumer. |
| `endpoint`                  | string | URL events are delivered to. |
| `tenant_id`                 | string | UUID of the associated tenant. |
| `description`               | string | Description of the event consumer. |
| `authorization_type`        | string | Authorization type (`API_KEY`, `BASIC`, or `OAUTH_CLIENT_CREDENTIALS`). |
| `authorization_parameters`  | string | JSON-encoded object containing authorization credentials. Sensitive credentials may be redacted. |
| `is_disabled`               | bool   | Whether this event consumer is disabled. |
| `last_sync_date`            | string | Timestamp of the last synchronization attempt. |
| `last_sync_message`         | string | Message from the last synchronization attempt. |
| `created_by`                | string | Identity that created this event consumer. |
| `created`                   | string | Timestamp when this event consumer was created. |
| `last_modified_by`          | string | Identity that last modified this event consumer. |
| `last_modified`             | string | Timestamp when this event consumer was last modified. |
| `version`                   | number | Version number of this event consumer. |
