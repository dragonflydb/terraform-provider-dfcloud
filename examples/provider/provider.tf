terraform {
  required_providers {
    dfcloud = {
      source = "registry.terraform.io/dragonflydb/dfcloud"
    }
  }
}

provider "dfcloud" {
  # Set the API key via the DFCLOUD_API_KEY environment variable.
  # Optionally, you can set the api_host to point to a different API endpoint.
}
