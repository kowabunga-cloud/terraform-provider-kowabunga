# Terraform provider for Kowabunga

This is a Terraform provider that lets you:

- setup and administrate a Kowabunga cluster (admin)
- provision resources on a Kowabunga cluster (user)

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

```
$ terraform init
```

### Provider configuration

```hcl
provider "kowabunga" {
  uri      = "http://kowabunga:port"
  token    = "kowabunga_api_token"
}
```

```
## Authors

* The Kowabunga Project (https://www.kowabunga.cloud/)
