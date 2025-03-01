PLUGIN_NAME := terraform-provider-dfcloud
PLUGIN_PATH := ./bin/

build:
	@echo "Building Terraform plugin..."
	goreleaser build --single-target --skip=validate --clean --snapshot

install-tfplugindocs:
	@echo "Installing tfplugindocs tool..."
	@go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.19.4

docs: install-tfplugindocs
	@echo "Updating documentation..."
	@tfplugindocs generate --provider-name dfcloud

.PHONY: install-tfplugindocs install-goreleaser docs

test:
	@echo "Run acceptance tests against the provider"
	TF_ACC=true go test ./... $(CLI_ARGS)

update-terraformrc:
	@echo 'provider_installation {\n  dev_overrides {\n    "registry.terraform.io/dragonflydb/dfcloud" = "$(PWD)/bin"\n    "registry.terraform.io/hashicorp/aws" = "$(PWD)/bin"\n  }\n\n  # For all other providers, install them directly from their origin provider\n  # registries as normal. If you omit this, Terraform will _only_ use\n  # the dev_overrides block, and so no other providers will be available.\n  direct {}\n}' > ~/.terraformrc

.PHONY: build install update-terraformrc
