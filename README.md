# Unleash Terraform Provider

A Terraform provider for managing [Unleash](https://unleash.github.io/) feature toggles. This provider allows you to
create, read, update, and delete feature toggles in Unleash which
the [Official Unleash Plugin](https://github.com/Unleash/terraform-provider-unleash) does not support.

## Terraform Usage

To define a provider configuration:-

```
terraform {
  required_providers {
    unleash = {
      source  = "LINEMANWongnai/unleash"
      version = "1.0.0"
    }
  }
}

provider "unleash" {
    base_url = "https://myunleash-host"
    authorization = "admin-api-key"
}
```

Then you can add feature resources:-

```
resource "unleash_feature" "default_feature_1" {
    project         = "default"
    name            = "feature_1"
    type            = "release"
    impression_data = true
    environments = {
        development = {
            enabled = true
            strategies = [
                {
                    disabled    = false
                    name        = "flexibleRollout"
                    parameters = {
                        groupId    = "feature_1"
                        rollout    = "100"
                        stickiness = "default"
                    }
                }
            ]
        }
        production = {
            enabled = false
            strategies = [
                {
                    disabled    = false
                    name        = "flexibleRollout"
                    parameters = {
                        groupId    = "feature_1"
                        rollout    = "100"
                        stickiness = "default"
                    }
                }
            ]
        }
    }
}
```

### Schema

* [provider](docs/index.md)
* [feature](docs/resources/feature.md)

## Generating existing features

Your unleash server may already have a lot of features defined. Initially you may want to put those features to .tf
file. This repository has go command to generate terraform resources from all features in a project. To generate from
all features under the `default` project, just run:-

```
# install command
go install github.com/LINEMANWongnai/terraform-provider-unleash/cmd/genunleash@latest

# run
genunleash default
```

The above command requires 2 environment variables:-

| Environment Variable Name   | Description                        |
|-----------------------------|------------------------------------|
| UNLEASH_BASE_URL            | Unleash URL. e.g. http://localhost |
| UNLEASH_AUTHORIZATION_TOKEN | Admin API Authorization token      |

If successfully run, you will see 2 output files which are 1) `gen.out.tf` and 2) `gen-import.out.tf`. The `gen.out.tf`
contains all features. The `gen-import.out.tf` contains import blocks to import those features to terraform state.

You can just use `terraform plan` to check for the changes then `terraform apply` to apply the changes. After
the `apply` process, you can remove the `gen-import.out.tf` .

Please be noted that the generated always generate `flexibleRollout` strategy for an empty one to avoid state conflict
after applying changes since Unleash server always creates a default one if there is no strategy defined.

## Development

To build all binaries in local machine:-

```
make 
```

To run all tests:-

```
make test
```

More information about plug-in development please
read [Terraform Plugin Development Tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework).

