package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const testAgePublicKeyResource = "age1j7ce327ke8t905hr4ve97xh4jr5ujauq59nxxkr3tnz9pty78p6q26hnd0"
const testAgeSecretKeyResource = "AGE-SECRET-KEY-18Z8D6LS5LCAZWERTYMK87NQ0N0ZEX5T50NZ9Q5XVPES2VRPWTC4SYAY5AT"

func testAccEncryptResourcePreCheck(t *testing.T) {
	testAccPreCheck(t)
}

func TestAccEncryptResource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigBasic(testAgePublicKeyResource),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptResource_InputChange_ForcesReplacement(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigUpdate("initial-value", testAgePublicKeyResource),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccEncryptResourceConfigUpdate("updated-value", testAgePublicKeyResource),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"sops_encrypt.test",
							plancheck.ResourceActionReplace,
						),
					},
				},
			},
		},
	})
}

func TestAccEncryptResource_AgeChange_ForcesReplacement(t *testing.T) {
	const testAgePublicKeyAlternate = "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigUpdate("test-value", testAgePublicKeyResource),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccEncryptResourceConfigUpdate("test-value", testAgePublicKeyAlternate),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"sops_encrypt.test",
							plancheck.ResourceActionReplace,
						),
					},
				},
			},
		},
	})
}

func TestAccEncryptResource_AgeListModification_ForcesReplacement(t *testing.T) {
	const testAgePublicKeyAlternate = "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigMultipleRecipients(
					testAgePublicKeyResource,
					testAgePublicKeyAlternate,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccEncryptResourceConfigBasic(testAgePublicKeyResource),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"sops_encrypt.test",
							plancheck.ResourceActionReplace,
						),
					},
				},
			},
		},
	})
}

func TestAccEncryptResource_NestedStructure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigNested(testAgePublicKeyResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptResource_MultipleRecipients(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigMultipleRecipients(testAgePublicKeyResource, testAgePublicKeyResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptResource_OutputTypeYAML(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigWithOutputType(testAgePublicKeyResource, "yaml"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("sops_encrypt.test", "output_type", "yaml"),
				),
			},
		},
	})
}

func TestAccEncryptResource_InvalidInputTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEncryptResourceConfigInvalidArray(testAgePublicKeyResource),
				ExpectError: regexp.MustCompile(`Input must be a map/object, got \[\]interface \{\}\. SOPS can only encrypt JSON\s+objects\.`),
			},
			{
				Config:      testAccEncryptResourceConfigInvalidString(testAgePublicKeyResource),
				ExpectError: regexp.MustCompile(`Input must be a map/object, got string\. SOPS can only encrypt JSON\s+objects\.`),
			},
			{
				Config:      testAccEncryptResourceConfigInvalidNumber(testAgePublicKeyResource),
				ExpectError: regexp.MustCompile(`Input must be a map/object, got float64\. SOPS can only encrypt JSON\s+objects\.`),
			},
		},
	})
}

func testAccEncryptResourceConfigBasic(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
  age = [%q]
}
`, testAgeSecretKeyResource, ageRecipient)
}

func testAccEncryptResourceConfigUpdate(value, ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = {
    secret = %q
    key    = "my-key-data"
  }
  age = [%q]
}
`, testAgeSecretKeyResource, value, ageRecipient)
}

func testAccEncryptResourceConfigNested(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = {
    database = {
      host     = "localhost"
      password = "secret123"
      port     = 5432
    }
    api_keys = {
      stripe   = "sk_live_secret"
      sendgrid = "SG.secret_key"
    }
    array_data = ["item1", "item2", "item3"]
  }
  age = [%q]
}
`, testAgeSecretKeyResource, ageRecipient)
}

func testAccEncryptResourceConfigMultipleRecipients(ageRecipient1, ageRecipient2 string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = {
    multi_recipient_secret = "shared-secret-value"
  }
  age = [%q, %q]
}
`, testAgeSecretKeyResource, ageRecipient1, ageRecipient2)
}

func testAccEncryptResourceConfigInvalidArray(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = ["not", "a", "map"]
  age = [%q]
}
`, testAgeSecretKeyResource, ageRecipient)
}

func testAccEncryptResourceConfigInvalidString(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = "not a map"
  age = [%q]
}
`, testAgeSecretKeyResource, ageRecipient)
}

func testAccEncryptResourceConfigInvalidNumber(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = 42
  age = [%q]
}
`, testAgeSecretKeyResource, ageRecipient)
}

func testAccEncryptResourceConfigWithOutputType(ageRecipient, outputType string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
  age = [%q]
  output_type = %q
}
`, testAgeSecretKeyResource, ageRecipient, outputType)
}

func TestAccEncryptResource_WriteOnlyBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigWriteOnlyBasic(testAgePublicKeyResource),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("input_wo"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("input_hash"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccEncryptResource_MutualExclusivity(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigBothInputs(testAgePublicKeyResource),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccEncryptResource_NoInputsProvided(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEncryptResourceConfigNoInputs(testAgePublicKeyResource),
				ExpectError: regexp.MustCompile("Either 'input' or 'input_wo' must be provided"),
			},
		},
	})
}

func TestAccEncryptResource_HashGeneration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigWriteOnlyValue("initial-value", testAgePublicKeyResource),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("input_hash"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("input_wo"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

func TestAccEncryptResource_VersionTriggerChangeDetection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigWriteOnlyWithVersion("test-value", "v1", testAgePublicKeyResource),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("input_hash"),
						knownvalue.Null(),
					),
				},
			},
			{
				Config: testAccEncryptResourceConfigWriteOnlyWithVersion("test-value", "v2", testAgePublicKeyResource),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"sops_encrypt.test",
							plancheck.ResourceActionReplace,
						),
					},
				},
			},
		},
	})
}

func TestAccEncryptResource_InputWOVersionWithoutInputWO(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigVersionWithoutWriteOnly(testAgePublicKeyResource),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccEncryptResourceConfigWriteOnlyBasic(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input_wo = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
  age = [%q]
}
`, testAgeSecretKeyResource, ageRecipient)
}

func testAccEncryptResourceConfigBothInputs(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = {
    secret = "from-input"
  }
  input_wo = {
    secret = "from-input-wo"
  }
  age = [%q]
}
`, testAgeSecretKeyResource, ageRecipient)
}

func testAccEncryptResourceConfigNoInputs(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  age = [%q]
}
`, testAgeSecretKeyResource, ageRecipient)
}

func testAccEncryptResourceConfigWriteOnlyValue(value, ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input_wo = {
    secret = %q
  }
  age = [%q]
}
`, testAgeSecretKeyResource, value, ageRecipient)
}

func testAccEncryptResourceConfigWriteOnlyWithVersion(value, version, ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input_wo = {
    secret = %q
  }
  input_wo_version = %q
  age = [%q]
}
`, testAgeSecretKeyResource, value, version, ageRecipient)
}

func testAccEncryptResourceConfigVersionWithoutWriteOnly(ageRecipient string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

resource "sops_encrypt" "test" {
  input = {
    secret = "my-secret"
  }
  input_wo_version = "v1"
  age = [%q]
}
`, testAgeSecretKeyResource, ageRecipient)
}
