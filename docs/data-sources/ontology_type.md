---
page_title: "tcm_ontology_type Data Source - TCM Provider"
description: |-
  Reads an OntologyType from the TCM service by ID.
---

# tcm_ontology_type (Data Source)

Reads an existing OntologyType from TCM by its UUID.

## Example Usage

```hcl
data "tcm_ontology_type" "lookup" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "ontology_type_name" {
  value = data.tcm_ontology_type.lookup.name
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the ontology type to read. |

## Attribute Reference

| Attribute     | Type   | Description |
|---------------|--------|-------------|
| `name`        | string | Name of the ontology type. |
| `description` | string | Description of the ontology type. |
| `is_disabled` | bool   | Whether this ontology type is disabled. |
