---
page_title: "tcm_erp_config_ontology_type Resource - TCM Provider"
description: |-
  Manages the mapping between an ERPConfig and an OntologyType in the TCM service.
---

# tcm_erp_config_ontology_type (Resource)

Manages the mapping between an ERPConfig and an OntologyType in the TCM service.

## Example Usage

```hcl
resource "tcm_erp_config_ontology_type" "acme_employee" {
  erp_config_id     = tcm_erp_config.acme.id
  ontology_type_id  = tcm_ontology_type.employee.id
}
```

## Argument Reference

| Argument           | Required | Type   | Description |
|--------------------|----------|--------|-------------|
| `erp_config_id`    | no       | string | UUID of the ERP config. |
| `ontology_type_id` | no       | string | UUID of the ontology type. |

## Attribute Reference

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

## Import

Import an existing ERP config ontology type mapping by its TCM UUID:

```bash
terraform import tcm_erp_config_ontology_type.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_erp_config_ontology_type.acme_employee 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
