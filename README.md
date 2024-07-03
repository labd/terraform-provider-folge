# Folge Terraform Provider

[![Test status](https://github.com/labd/terraform-provider-folge/workflows/Run%20Tests/badge.svg)](https://github.com/labd/terraform-provider-folge/actions?query=workflow%3A%22Run+Tests%22)
[![codecov](https://codecov.io/gh/LabD/terraform-provider-folge/branch/master/graph/badge.svg)](https://codecov.io/gh/LabD/terraform-provider-folge)
[![Go Report Card](https://goreportcard.com/badge/github.com/labd/terraform-provider-folge)](https://goreportcard.com/report/github.com/labd/terraform-provider-folge)

The Terraform Folge provider allows you to configure your
[folge](https://app.folge.io/) space with infrastructure-as-code principles.

# Quick start

[Read our documentation](https://registry.terraform.io/providers/labd/folge/latest/docs)
and check out the [examples](https://registry.terraform.io/providers/labd/folge/latest/docs/guides/examples).

## Usage

The provider is distributed via the Terraform registry. To use it you need to configure
the [`required_provider`](https://www.terraform.io/language/providers/requirements#requiring-providers) block. For example:

```hcl
terraform {
  required_providers {
    folge = {
      source = "labd/folge"

      # It's recommended to pin the version, e.g.:
      # version = "~> 0.0.1"
    }
  }
}

provider "folge" {
  client_id     = "<client-id>
  client_secret = "<client-secret>"
}

resource "folge_application" "test" {
  name = "My application"
}

resource "folge_datasource" "test" {
  application_id = folge_application.test.id

  name = "My datasource"
  url    = "http://localhost:8000/api/healthcheck"

  basic_auth {
    username = "my-user"
    password = "my-password"
  }
}

resource "folge_check_http_status" "test" {
  datasource_id = folge_datasource.test.id

  crontab     = "*/5 * * * *"
  label       = "My check"
  status_code = 200
}


resource "folge_check_json_property" "test" {
  datasource_id = folge_datasource.test.id

  crontab    = "*/5 * * * *"
  label      = "My check"
  path       = "authentication.failed"
  operator   = "HAS_VALUE"
  value_bool = false
}


# Forwards metrics to an OpenTelemetry Collector
resource "folge_metrics_reader" "test" {
  datasource_id = folge_datasource.test.id

  crontab = "*/5 * * * *"
  fields  = ["*"]

  target = {
    type = "otel"
    url  = "http://localhost:4317"
    headers = {
      "X-Api-Key" = "secret"
    }
  }
}


# Store metrics in folge
resource "folge_metrics_reader" "test" {
  datasource_id = folge_datasource.test.id

  crontab = "*/5 * * * *"
  fields  = ["*"]

  target = {
    type = "internal"
  }
}
```

# Contributing

## Building the provider

Clone the repository and run the following command:

```sh
$ task build-local
```

## Debugging / Troubleshooting

There are two environment settings for troubleshooting:

- `TF_LOG=INFO` enables debug output for Terraform.

Note this generates a lot of output!

## Releasing

Install "changie"

```
brew tap miniscruff/changie https://github.com/miniscruff/changie
brew install changie
```

Add unreleased change files by running for each change (add/fix/remove/etc.)

```
changie new
```

Commit this and a new PR will be created.

Once that's merged and its Github action is complete, a new release will be live.

## Testing

### Running unit tests

```sh
$ task test
```

### Running acceptance tests

```sh
$ task testacc
```

Note that acceptance tests by default run based on pre-recorded results. The
test stubs can be found in [internal/assets] (./internal/assets). A good habit
is to create a separate stub file per test case, as otherwise there might be
conflicts when multiple tests are run in parallel.

When adding or updating tests locally you can set `RECORD=true` to re-record
results. This will clear all previous results and create a new snapshot of the
API interaction.

## Authors

This project is developed by [Lab Digital](https://www.labdigital.nl). We
welcome additional contributors. Please see our
[GitHub repository](https://github.com/labd/terraform-provider-folge)
for more information.

