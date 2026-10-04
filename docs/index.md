---
page_title: "TCM Provider"
description: |-
  The TCM provider manages resources in the symplr Tenant Configuration Management service.
---

# TCM Provider

The TCM provider manages resources in the symplr **Tenant Configuration Management (TCM)** service. It is an internal provider — not published to the Terraform Registry — and is used via local dev overrides.

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

## Argument Reference

| Argument   | Required | Description |
|------------|----------|-------------|
| `base_url` | yes | TCM service base URL. |
| `token`    | yes | OAuth Bearer token. Requires the `tenant:configuration:api:write` scope (read operations also need `tenant:configuration:api:read`). |

## Environment URLs

| Environment | URL |
|-------------|-----|
| stable      | `https://stable-platform.symplr.com/ce-platform-tenant-configuration-service` |
| staging     | `https://stg-platform.symplr.com/ce-platform-tenant-configuration-service` |
| production  | `https://platform.symplr.com/ce-platform-tenant-configuration-service` |

## Resources

- [tcm_product](resources/product.md)
- [tcm_namespace](resources/namespace.md)
- [tcm_tenant](resources/tenant.md)
- [tcm_tenant_product](resources/tenant_product.md)
- [tcm_tenant_product_environment](resources/tenant_product_environment.md)
- [tcm_system_info](resources/system_info.md)

## Data Sources

- [tcm_namespace](data-sources/namespace.md)
- [tcm_tenant](data-sources/tenant.md)
- [tcm_tenant_product](data-sources/tenant_product.md)
- [tcm_tenant_product_environment](data-sources/tenant_product_environment.md)
- [tcm_system_info](data-sources/system_info.md)
