<p align="center">
  <a href="https://www.kowabunga.cloud/?utm_source=github&utm_medium=logo" target="_blank">
    <picture>
      <source srcset="https://raw.githubusercontent.com/kowabunga-cloud/infographics/master/art/kowabunga-title-white.png" media="(prefers-color-scheme: dark)" />
      <source srcset="https://raw.githubusercontent.com/kowabunga-cloud/infographics/master/art/kowabunga-title-black.png" media="(prefers-color-scheme: light), (prefers-color-scheme: no-preference)" />
      <img src="https://raw.githubusercontent.com/kowabunga-cloud/infographics/master/art/kowabunga-title-black.png" alt="Kowabunga" width="800">
    </picture>
  </a>
</p>

# Terraform provider for Kowabunga

This is official Terraform/OpenTofu provider for Kowabunga.

[![License: Apache License, Version 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://spdx.org/licenses/Apache-2.0.html)
[![go.dev](https://img.shields.io/badge/go.dev-pkg-007d9c.svg?style=flat)](https://pkg.go.dev/github.com/kowabunga-cloud/terraform-provider-kowabunga)

It allows you to:

- setup and administrate a Kowabunga cluster (admin)
- provision resources on a Kowabunga cluster (user)

## Current Releases

| Project            | Release Badge                                                                                       |
|--------------------|-----------------------------------------------------------------------------------------------------|
| **Kowabunga**           | [![Kowabunga Release](https://img.shields.io/github/v/release/kowabunga-cloud/kowabunga)](https://github.com/kowabunga-cloud/kowabunga/releases) |
| **Kowabunga Terraform Provider**     | [![Kowabunga Terraform Provider](https://img.shields.io/github/v/release/kowabunga-cloud/terraform-provider-kowabunga)](https://github.com/kowabunga-cloud/terraform-provider-kowabunga/releases) |

## Getting Started

In your `main.tf` file, specify the version you want to use:

```hcl
terraform {
  required_providers {
    kowabunga = {
      source = "kowabunga-cloud/kowabunga"
    }
  }
}

provider "kowabunga" {
  # Configuration options
}
```

And now run terraform init:

```sh
$ terraform init
```

### Provider configuration

```hcl
provider "kowabunga" {
  uri      = "http://kowabunga:port"
  token    = "kowabunga_api_token"
}
```

## Authors

The Kowabunga Project (https://www.kowabunga.cloud/)

## License

Licensed under [Apache License, Version 2.0](https://opensource.org/license/apache-2-0), see [`LICENSE`](LICENSE).
