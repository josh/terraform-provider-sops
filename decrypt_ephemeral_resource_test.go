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

const testAgePublicKeyEphemeral = "age1j7ce327ke8t905hr4ve97xh4jr5ujauq59nxxkr3tnz9pty78p6q26hnd0"
const testAgeSecretKeyEphemeral = "AGE-SECRET-KEY-18Z8D6LS5LCAZWERTYMK87NQ0N0ZEX5T50NZ9Q5XVPES2VRPWTC4SYAY5AT"

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
				Config: testAccDecryptEphemeralResourceConfigBasic(testAgePublicKeyEphemeral),
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
				Config: testAccDecryptEphemeralResourceConfigNested(testAgePublicKeyEphemeral),
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
				Config: testAccDecryptEphemeralResourceConfigWithAgeIdentities(testAgePublicKeyEphemeral),
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
  input = data.sops_encrypt.source.output
}

provider "echo" {
  data = ephemeral.sops_decrypt.test.output
}

resource "echo" "test" {}
`, testAgeSecretKeyEphemeral, ageRecipient)
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
  input = data.sops_encrypt.source.output
}

provider "echo" {
  data = ephemeral.sops_decrypt.test.output
}

resource "echo" "test" {}
`, testAgeSecretKeyEphemeral, ageRecipient)
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
  input = data.sops_encrypt.source.output
}

provider "echo" {
  data = ephemeral.sops_decrypt.test.output
}

resource "echo" "test" {}
`, testAgeSecretKeyEphemeral, ageRecipient)
}
