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

**Data Sources**

| Data Source | Description |
|-------------|-------------|
| [`tcm_product_environment`](docs/data-sources/product_environment.md) | Reads a Product Environment by ID. |
| [`tcm_namespace`](docs/data-sources/namespace.md) | Reads a Namespace by ID. |
| [`tcm_tenant`](docs/data-sources/tenant.md) | Reads a Tenant by ID. |
| [`tcm_tenant_product`](docs/data-sources/tenant_product.md) | Reads a TenantProduct mapping by ID. |
| [`tcm_tenant_product_environment`](docs/data-sources/tenant_product_environment.md) | Reads a TenantProductEnvironment by ID. |
| [`tcm_system_info`](docs/data-sources/system_info.md) | Reads a SystemInfo entry by ID. |
