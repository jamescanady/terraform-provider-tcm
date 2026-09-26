# terraform-provider-tcm

Internal Terraform provider for the symplr **Tenant Configuration Management (TCM)** service. Not published to the Terraform Registry — used via local dev overrides only.

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

## Resources

### `tcm_product`

Manages a Product entry in TCM.

> **Note:** TCM has no hard-delete endpoint for products. Running `terraform destroy` will **soft-delete** the product by setting `isDisabled = true` via `PUT /v1/Product/{id}`. The record remains in TCM.

#### Example

```hcl
resource "tcm_product" "event_engine_audit" {
  name        = "EventEngineAudit"
  description = "Event Engine Audit product"
}
```

#### Arguments

| Argument      | Required | Type   | Description |
|---------------|----------|--------|-------------|
| `name`        | yes      | string | Product name. |
| `description` | yes      | string | Product description. Maximum 110 characters. |
| `is_disabled` | no       | bool   | Whether the product is disabled. Defaults to `false`. |

#### Attributes (read-only)

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | string | UUID assigned by TCM on creation. |

---

## Importing Existing Products

If a product already exists in TCM and you want to bring it under Terraform management, import it using its TCM UUID:

```bash
terraform import tcm_product.<resource_name> <uuid>
```

**Example:**
```bash
terraform import tcm_product.event_engine_audit 3f2504e0-4f89-11d3-9a0c-0305e82c3301
```

Terraform will populate the rest of the state from TCM via a `Read` call after the import.

### Finding the UUID

If you only know the product name, use the TCM `find` endpoint to look up the UUID:

```bash
curl -s \
  -H "Authorization: Bearer $TCM_TOKEN" \
  "https://stable-platform.symplr.com/ce-platform-tenant-configuration-service/v1/Product/find/EventEngineAudit" \
  | jq '.[0].id'
```

Use the returned UUID in the `terraform import` command above.

> **Important:** If you skip the import and run `terraform apply` against a product that already exists, TCM will return a `409 Conflict` and the apply will fail. Always import first when the product is pre-existing.
