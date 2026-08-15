package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDecrypt_UnknownProviderIdentityValue(t *testing.T) {
	// The ambient key must not silently stand in while the configured
	// identity is unknown during planning.
	t.Setenv("SOPS_AGE_KEY", testAgeSecretKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sops_age_private_key" "k" {}

provider "sops" {
  alias              = "keyed"
  age_identity_value = sops_age_private_key.k.private_key
}

data "sops_encrypt" "test" {
  input = {
    secret = "unknown-provider-config"
  }
  age_recipients = [%q]
}

data "sops_decrypt" "test" {
  provider   = sops.keyed
  input      = data.sops_encrypt.test.output
  input_type = "json"
}
`, testAgePublicKey),
				ExpectError: regexp.MustCompile("Unknown Configuration Value"),
			},
		},
	})
}

func TestAccDecrypt_UnknownPathIgnoredWhenValueKnown(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// The known identity value takes precedence, so the unknown
				// path must not block planning.
				Config: fmt.Sprintf(`
resource "sops_age_private_key" "k" {}

provider "sops" {
  alias              = "keyed"
  age_identity_value = %q
  age_identity_path  = sops_age_private_key.k.public_key
}

data "sops_encrypt" "test" {
  input = {
    secret = "value-precedence"
  }
  age_recipients = [%q]
}

data "sops_decrypt" "test" {
  provider   = sops.keyed
  input      = data.sops_encrypt.test.output
  input_type = "json"
}
`, testAgeSecretKey, testAgePublicKey),
				Check: resource.TestCheckResourceAttr(
					"data.sops_decrypt.test", "output.secret", "value-precedence",
				),
			},
		},
	})
}
