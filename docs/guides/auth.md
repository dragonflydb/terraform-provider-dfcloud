# Authentication Guide

Visit [dragonflydb.cloud](https://dragonflydb.cloud) and generate an API key with the required permissions. Then set the following environment variable before using the Terraform provider:

```bash
export DFCLOUD_API_KEY=<YOUR_API_KEY>
```

The same API key can also be set in the provider configuration block in the Terraform configuration file:

```hcl
provider "dfcloud" {
  api_key = "<YOUR_API_KEY>"
}
```
