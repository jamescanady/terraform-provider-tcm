---
page_title: "tcm_erp_config_ontology_type Data Source - TCM Provider"
description: |-
  Reads an ERPConfig-to-OntologyType mapping from the TCM service by ID.
---

# tcm_erp_config_ontology_type (Data Source)

Reads an existing ERPConfig-to-OntologyType mapping from TCM by its UUID.

## Example Usage

```hcl
data "tcm_erp_config_ontology_type" "lookup" {
  id = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
}

output "erp_config_id" {
  value = data.tcm_erp_config_ontology_type.lookup.erp_config_id
}
```

## Argument Reference

| Argument | Required | Type   | Description |
|----------|----------|--------|-------------|
| `id`     | yes      | string | UUID of the ERP config ontology type mapping to read. |

## Attribute Reference

| Attribute          | Type   | Description |
|--------------------|--------|-------------|
| `erp_config_id`    | string | UUID of the associated ERP config. |
| `ontology_type_id` | string | UUID of the associated ontology type. |
