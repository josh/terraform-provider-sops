package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func testAccDecryptEphemeralPreCheck(t *testing.T) {
	testAccPreCheck(t)
}

func TestAccDecryptEphemeralResource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck:                 func() { testAccDecryptEphemeralPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		Steps: []resource.TestStep{
			{
				Config: testAccDecryptEphemeralResourceConfigBasic(testAgePublicKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("secret"),
						knownvalue.StringExact("my-secret-value"),
					),
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("key"),
						knownvalue.StringExact("my-key-data"),
					),
				},
			},
		},
	})
}

func TestAccDecryptEphemeralResource_NestedStructure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck:                 func() { testAccDecryptEphemeralPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		Steps: []resource.TestStep{
			{
				Config: testAccDecryptEphemeralResourceConfigNested(testAgePublicKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("database").AtMapKey("password"),
						knownvalue.StringExact("super-secret-db-password"),
					),
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("database").AtMapKey("host"),
						knownvalue.StringExact("prod-db.example.com"),
					),
				},
			},
		},
	})
}

func TestAccDecryptEphemeralResource_WithAgeIdentities(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck:                 func() { testAccDecryptEphemeralPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		Steps: []resource.TestStep{
			{
				Config: testAccDecryptEphemeralResourceConfigWithAgeIdentities(testAgePublicKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("message"),
						knownvalue.StringExact("ephemeral-secret-message"),
					),
				},
			},
		},
	})
}

func testAccDecryptEphemeralResourceConfigBasic(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

data "sops_encrypt" "source" {
  input = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
  age_recipients = [%q]
}

ephemeral "sops_decrypt" "test" {
  input      = data.sops_encrypt.source.output
  input_type = "json"
}

provider "echo" {
  data = ephemeral.sops_decrypt.test.output
}

resource "echo" "test" {}
`, testAgeSecretKey, ageRecipient)
}

func testAccDecryptEphemeralResourceConfigNested(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

data "sops_encrypt" "source" {
  input = {
    database = {
      host     = "prod-db.example.com"
      password = "super-secret-db-password"
      port     = 5432
    }
    api_keys = {
      stripe   = "sk_live_secret_key"
      sendgrid = "SG.secret_api_key"
    }
    array_data = ["item1", "item2", "item3"]
  }
  age_recipients = [%q]
}

ephemeral "sops_decrypt" "test" {
  input      = data.sops_encrypt.source.output
  input_type = "json"
}

provider "echo" {
  data = ephemeral.sops_decrypt.test.output
}

resource "echo" "test" {}
`, testAgeSecretKey, ageRecipient)
}

func testAccDecryptEphemeralResourceConfigWithAgeIdentities(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

data "sops_encrypt" "source" {
  input = {
    message = "ephemeral-secret-message"
    config = {
      enabled = true
      timeout = 30
    }
  }
  age_recipients = [%q]
}

ephemeral "sops_decrypt" "test" {
  input      = data.sops_encrypt.source.output
  input_type = "json"
}

provider "echo" {
  data = ephemeral.sops_decrypt.test.output
}

resource "echo" "test" {}
`, testAgeSecretKey, ageRecipient)
}

func TestAccDecryptEphemeralResource_InputType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck:                 func() { testAccDecryptEphemeralPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		Steps: []resource.TestStep{
			{
				Config: testAccDecryptEphemeralResourceConfigInputType(testAgePublicKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("foo"),
						knownvalue.StringExact("bar"),
					),
				},
			},
		},
	})
}

func testAccDecryptEphemeralResourceConfigInputType(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

data "sops_encrypt" "source" {
  input = {
    foo = "bar"
  }
  age_recipients = [%q]
}

ephemeral "sops_decrypt" "test" {
  input      = data.sops_encrypt.source.output
  input_type = "json"
}

provider "echo" {
  data = ephemeral.sops_decrypt.test.output
}

resource "echo" "test" {}
`, testAgeSecretKey, ageRecipient)
}
