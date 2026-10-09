---
page_title: "tcm_ontology_type Resource - TCM Provider"
description: |-
  Manages an OntologyType in the TCM service. Note: is_disabled is read-only and managed by the TCM service.
---

# tcm_ontology_type (Resource)

Manages an OntologyType in the TCM service.

`is_disabled` is read-only and managed by the TCM service. It cannot be set via Terraform configuration.

## Example Usage

```hcl
resource "tcm_ontology_type" "employee" {
  name        = "Employee"
  description = "Maps ERP employee records"
}
```

## Argument Reference

| Argument      | Required | Type   | Description |
|---------------|----------|--------|-------------|
| `name`        | no       | string | Name of the ontology type. |
| `description` | no       | string | Description of the ontology type. |

## Attribute Reference

| Attribute     | Type   | Description |
|---------------|--------|-------------|
| `id`          | string | UUID assigned by TCM on creation. |
| `is_disabled` | bool   | Whether this ontology type is disabled. Read-only; managed by the TCM service. |

## Import

Import an existing ontology type by its TCM UUID:

```bash
terraform import tcm_ontology_type.<name> <uuid>
```

**Example:**

```bash
terraform import tcm_ontology_type.employee 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```
