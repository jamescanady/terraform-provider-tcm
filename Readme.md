# terraform-provider-tcm

Internal Terraform provider for the symplr **Tenant Configuration Management (TCM)** service. Not published to the Terraform Registry — used via local dev overrides only.

---

## Releasing

Releases are created by pushing a version tag. CI will build, package, and publish a GitHub Release automatically via GoReleaser.

```bash
git tag v1.2.3
git push origin v1.2.3
```

Follow [Semantic Versioning](https://semver.org): `MAJOR.MINOR.PATCH`.

| Change type | Version bump | Example |
|---|---|---|
| Breaking change | Major | `v1.0.0` → `v2.0.0` |
| New resource or argument | Minor | `v1.0.0` → `v1.1.0` |
| Bug fix | Patch | `v1.0.0` → `v1.0.1` |

The tag must start with `v` (lowercase). The release workflow triggers only on tag pushes — merging to `main` runs CI only.

---

## Building

Requires Go 1.26+.

```bash
cd provider-tcm
go mod download
go build -o terraform-provider-tcm .
```

The output binary (`terraform-provider-tcm`) must be installed where Terraform can find it (see [Local Installation](#local-installation) below).

---

## Local Installation

Create or update `~/.terraformrc` to point Terraform at the local binary:

```hcl
provider_installation {
  dev_overrides {
    "symplr/tcm" = "/absolute/path/to/ui-terraform/provider-tcm"
  }
  direct {}
}
```

With dev overrides active, `terraform init` is not required — Terraform uses the binary directly. Run `terraform plan` or `terraform apply` straight away.

---

## Provider Configuration

```hcl
terraform {
  required_providers {
    tcm = {
      source = "symplr/tcm"
    }
  }
}

provider "tcm" {
  base_url = "https://stable-platform.symplr.com/ce-platform-tenant-configuration-service"
  token    = var.tcm_token
}
```

| Argument   | Required | Description |
|------------|----------|-------------|
| `base_url` | yes | TCM service base URL. Use `stable` for non-production, `platform.symplr.com` for production. |
| `token`    | yes | OAuth Bearer token. Must have the `tenant:configuration:api:write` scope (read operations also require `tenant:configuration:api:read`). |

Available base URLs by environment:

| Environment | URL |
|-------------|-----|
| stable      | `https://stable-platform.symplr.com/ce-platform-tenant-configuration-service` |
| staging     | `https://stg-platform.symplr.com/ce-platform-tenant-configuration-service` |
| production  | `https://platform.symplr.com/ce-platform-tenant-configuration-service` |

---

## Resources and Data Sources

Full reference documentation is in the [`docs/`](docs/) folder.

**Resources**

| Resource | Description |
|----------|-------------|
| [`tcm_product`](docs/resources/product.md) | Manages a Product. |
| [`tcm_product_environment`](docs/resources/product_environment.md) | Manages a Product Environment. |
| [`tcm_namespace`](docs/resources/namespace.md) | Manages a Namespace. |
| [`tcm_tenant`](docs/resources/tenant.md) | Manages a Tenant. |
| [`tcm_tenant_product`](docs/resources/tenant_product.md) | Manages a Tenant ↔ Product mapping. |
| [`tcm_tenant_product_environment`](docs/resources/tenant_product_environment.md) | Manages a Tenant ↔ Product Environment mapping. |
| [`tcm_system_info`](docs/resources/system_info.md) | Manages a SystemInfo connection entry. |
| [`tcm_event_type`](docs/resources/event_type.md) | Manages an EventType. |
| [`tcm_event_consumer`](docs/resources/event_consumer.md) | Manages an EventConsumer (webhook subscriber). |
| [`tcm_event_type_consumer`](docs/resources/event_type_consumer.md) | Manages an EventType ↔ EventConsumer subscription. |
| [`tcm_schedule`](docs/resources/schedule.md) | Manages a Schedule. |
| [`tcm_schedule_category`](docs/resources/schedule_category.md) | Manages a ScheduleCategory. |
| [`tcm_ontology_type`](docs/resources/ontology_type.md) | Manages an OntologyType. |
| [`tcm_erp_config_ontology_type`](docs/resources/erp_config_ontology_type.md) | Manages an ERPConfig ↔ OntologyType mapping. |
| [`tcm_erp`](docs/resources/erp.md) | Manages an ERP record. |
| [`tcm_erp_config`](docs/resources/erp_config.md) | Manages an ERP integration configuration. |
| [`tcm_ehr_endpoint`](docs/resources/ehr_endpoint.md) | Manages an EHR Endpoint. |
| [`tcm_ehr_destination`](docs/resources/ehr_destination.md) | Manages an EHR Destination. |
| [`tcm_ehr_source`](docs/resources/ehr_source.md) | Manages an EHR Source. |
| [`tcm_ehr_tenant`](docs/resources/ehr_tenant.md) | Manages an EHR Tenant configuration. |
| [`tcm_ehr_schema`](docs/resources/ehr_schema.md) | Manages an EHR Schema definition. |
| [`tcm_ehr_json_metadata`](docs/resources/ehr_json_metadata.md) | Manages EHR JSON Metadata. |

**Data Sources**

| Data Source | Description |
|-------------|-------------|
| [`tcm_product_environment`](docs/data-sources/product_environment.md) | Reads a Product Environment by ID. |
| [`tcm_namespace`](docs/data-sources/namespace.md) | Reads a Namespace by ID. |
| [`tcm_tenant`](docs/data-sources/tenant.md) | Reads a Tenant by ID. |
| [`tcm_tenant_product`](docs/data-sources/tenant_product.md) | Reads a TenantProduct mapping by ID. |
| [`tcm_tenant_product_environment`](docs/data-sources/tenant_product_environment.md) | Reads a TenantProductEnvironment by ID. |
| [`tcm_system_info`](docs/data-sources/system_info.md) | Reads a SystemInfo entry by ID. |
| [`tcm_event_type`](docs/data-sources/event_type.md) | Reads an EventType by ID. |
| [`tcm_event_consumer`](docs/data-sources/event_consumer.md) | Reads an EventConsumer by ID. |
| [`tcm_event_type_consumer`](docs/data-sources/event_type_consumer.md) | Reads an EventTypeConsumer mapping by ID. |
| [`tcm_schedule`](docs/data-sources/schedule.md) | Reads a Schedule by ID. |
| [`tcm_schedule_category`](docs/data-sources/schedule_category.md) | Reads a ScheduleCategory by ID. |
| [`tcm_ontology_type`](docs/data-sources/ontology_type.md) | Reads an OntologyType by ID. |
| [`tcm_erp_config_ontology_type`](docs/data-sources/erp_config_ontology_type.md) | Reads an ERPConfigOntologyType mapping by ID. |
| [`tcm_erp`](docs/data-sources/erp.md) | Reads an ERP record by ID. |
| [`tcm_erp_config`](docs/data-sources/erp_config.md) | Reads an ERPConfig by ID. |
| [`tcm_ehr_endpoint`](docs/data-sources/ehr_endpoint.md) | Reads an EHR Endpoint by ID. |
| [`tcm_ehr_destination`](docs/data-sources/ehr_destination.md) | Reads an EHR Destination by ID. |
| [`tcm_ehr_source`](docs/data-sources/ehr_source.md) | Reads an EHR Source by ID. |
| [`tcm_ehr_tenant`](docs/data-sources/ehr_tenant.md) | Reads an EHR Tenant configuration by ID. |
| [`tcm_ehr_schema`](docs/data-sources/ehr_schema.md) | Reads an EHR Schema by ID. |
| [`tcm_ehr_json_metadata`](docs/data-sources/ehr_json_metadata.md) | Reads EHR JSON Metadata by ID. |

---

## Testing

The test suite is split into two tiers:

### Fast unit tests

White-box tests in `package internal` that call `toPayload` and `applyXxxResult` directly — no HTTP, no Terraform engine. Each test completes in microseconds.

```bash
go test ./internal/...
```

### Integration tests (slow)

Full lifecycle tests that spin up an in-process mock HTTP server and drive Terraform plan/apply/destroy via `resource.UnitTest`. Each lifecycle takes ~13 seconds due to the Terraform gRPC subprocess. Gated by the `integration` build tag.

```bash
go test -tags integration ./internal/...
```

To run a single resource's lifecycle test:

```bash
go test -tags integration -run TestAccEHRSchemaResource_lifecycle ./internal/...
```

> **IDE note:** Files ending in `_lifecycle_test.go` carry `//go:build integration` and will show a gopls warning ("No packages found") in editors that don't pass `-tags=integration` to the language server. Add `"buildFlags": ["-tags=integration"]` to your gopls settings to suppress this.
