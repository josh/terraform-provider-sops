package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const testAgePublicKeyResource = "age1j7ce327ke8t905hr4ve97xh4jr5ujauq59nxxkr3tnz9pty78p6q26hnd0"

func testAccEncryptResourcePreCheck(t *testing.T) {
	testAccPreCheck(t)
}

func testAccCheckEncryptedOutputIndentation(resourceName string, expectedIndent int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		output := rs.Primary.Attributes["output"]
		if output == "" {
			return fmt.Errorf("No output attribute found")
		}

		lines := strings.Split(output, "\n")

		if expectedIndent == 0 {
			for _, line := range lines {
				if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
					return fmt.Errorf("Expected compact output with no indentation, but found line starting with whitespace: %q", line)
				}
			}
		} else if expectedIndent > 0 {
			foundIndentedLine := false
			expectedPrefix := strings.Repeat(" ", expectedIndent)

			for _, line := range lines {
				if len(line) > expectedIndent && strings.HasPrefix(line, expectedPrefix) && !strings.HasPrefix(line, expectedPrefix+" ") {
					foundIndentedLine = true
					if strings.HasPrefix(line, "\t") {
						return fmt.Errorf("Expected %d spaces for indentation, but found tab character", expectedIndent)
					}
					break
				}
			}

			if !foundIndentedLine {
				return fmt.Errorf("Expected to find lines with %d-space indentation, but didn't find any", expectedIndent)
			}
		}

		return nil
	}
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
resource "sops_encrypt" "test" {
  input = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
  age_recipients = [%q]
}
`, ageRecipient)
}

func testAccEncryptResourceConfigUpdate(value, ageRecipient string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = {
    secret = %q
    key    = "my-key-data"
  }
  age_recipients = [%q]
}
`, value, ageRecipient)
}

func testAccEncryptResourceConfigNested(ageRecipient string) string {
	return fmt.Sprintf(`
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
  age_recipients = [%q]
}
`, ageRecipient)
}

func testAccEncryptResourceConfigMultipleRecipients(ageRecipient1, ageRecipient2 string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = {
    multi_recipient_secret = "shared-secret-value"
  }
  age_recipients = [%q, %q]
}
`, ageRecipient1, ageRecipient2)
}

func testAccEncryptResourceConfigInvalidArray(ageRecipient string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = ["not", "a", "map"]
  age_recipients = [%q]
}
`, ageRecipient)
}

func testAccEncryptResourceConfigInvalidString(ageRecipient string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = "not a map"
  age_recipients = [%q]
}
`, ageRecipient)
}

func testAccEncryptResourceConfigInvalidNumber(ageRecipient string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = 42
  age_recipients = [%q]
}
`, ageRecipient)
}

func testAccEncryptResourceConfigWithOutputType(ageRecipient, outputType string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
  age_recipients = [%q]
  output_type = %q
}
`, ageRecipient, outputType)
}

func TestAccEncryptResource_WriteOnlyBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
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
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
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
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
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
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
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
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
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
resource "sops_encrypt" "test" {
  input_wo = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
  age_recipients = [%q]
}
`, ageRecipient)
}

func testAccEncryptResourceConfigBothInputs(ageRecipient string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = {
    secret = "from-input"
  }
  input_wo = {
    secret = "from-input-wo"
  }
  age_recipients = [%q]
}
`, ageRecipient)
}

func testAccEncryptResourceConfigNoInputs(ageRecipient string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  age_recipients = [%q]
}
`, ageRecipient)
}

func testAccEncryptResourceConfigWriteOnlyValue(value, ageRecipient string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input_wo = {
    secret = %q
  }
  age_recipients = [%q]
}
`, value, ageRecipient)
}

func testAccEncryptResourceConfigWriteOnlyWithVersion(value, version, ageRecipient string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input_wo = {
    secret = %q
  }
  input_wo_version = %q
  age_recipients = [%q]
}
`, value, version, ageRecipient)
}

func testAccEncryptResourceConfigVersionWithoutWriteOnly(ageRecipient string) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = {
    secret = "my-secret"
  }
  input_wo_version = "v1"
  age_recipients = [%q]
}
`, ageRecipient)
}

func TestAccEncryptResource_UnknownInputWO_FromEphemeral(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigUnknownInputWO(testAgePublicKeyResource),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(
							"sops_encrypt.test",
							tfjsonpath.New("output"),
						),
						plancheck.ExpectUnknownValue(
							"sops_encrypt.test",
							tfjsonpath.New("input_hash"),
						),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
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

func testAccEncryptResourceConfigUnknownInputWO(ageRecipient string) string {
	return fmt.Sprintf(`
resource "terraform_data" "source" {
  input = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
}

ephemeral "sops_test_dynamic" "test" {
  value = terraform_data.source.output
}

resource "sops_encrypt" "test" {
  input_wo = ephemeral.sops_test_dynamic.test.output
  age_recipients = [%q]
}
`, ageRecipient)
}

func TestAccEncryptResource_UnknownInputWO_WithVersion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigUnknownInputWOWithVersion(testAgePublicKeyResource, "v1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(
							"sops_encrypt.test",
							tfjsonpath.New("output"),
						),
					},
				},
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

func testAccEncryptResourceConfigUnknownInputWOWithVersion(ageRecipient, version string) string {
	return fmt.Sprintf(`
resource "terraform_data" "source" {
  input = {
    secret = "my-secret-value"
  }
}

ephemeral "sops_test_dynamic" "test" {
  value = terraform_data.source.output
}

resource "sops_encrypt" "test" {
  input_wo = ephemeral.sops_test_dynamic.test.output
  input_wo_version = %q
  age_recipients = [%q]
}
`, version, ageRecipient)
}

func TestAccEncryptResource_FullUnknownNestedInput(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigFullUnknown(testAgePublicKeyResource),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(
							"sops_encrypt.test",
							tfjsonpath.New("output"),
						),
					},
				},
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

func testAccEncryptResourceConfigFullUnknown(ageRecipient string) string {
	return fmt.Sprintf(`
resource "terraform_data" "source" {
  input = {
    database = {
      host     = "prod-db.example.com"
      password = "super-secret-db-password"
    }
    api_key = "my-api-key"
  }
}

ephemeral "sops_test_dynamic" "test" {
  value = terraform_data.source.output
}

resource "sops_encrypt" "test" {
  input_wo = ephemeral.sops_test_dynamic.test.output
  age_recipients = [%q]
}
`, ageRecipient)
}
func TestAccEncryptResource_ValidatorMixedKnownUnknown(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigValidatorMixedKnownUnknown(testAgePublicKeyResource),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(
							"sops_encrypt.test",
							tfjsonpath.New("output"),
						),
					},
				},
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

func testAccEncryptResourceConfigValidatorMixedKnownUnknown(ageRecipient string) string {
	return fmt.Sprintf(`
resource "terraform_data" "source" {
  input = {
    secret_value = "my-secret-value"
  }
}

ephemeral "sops_test_dynamic" "test" {
  value = terraform_data.source.output
}

resource "sops_encrypt" "test" {
  input_wo = {
    static_key = "static-value"
    dynamic_key = ephemeral.sops_test_dynamic.test.output.secret_value
    another_static = "another-value"
  }
  age_recipients = [%q]
}
`, ageRecipient)
}

func TestAccEncryptResource_ValidatorNestedUnknownInMap(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigValidatorNestedUnknownInMap(testAgePublicKeyResource),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(
							"sops_encrypt.test",
							tfjsonpath.New("output"),
						),
					},
				},
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

func testAccEncryptResourceConfigValidatorNestedUnknownInMap(ageRecipient string) string {
	return fmt.Sprintf(`
resource "terraform_data" "source" {
  input = {
    config = {
      database_password = "secret123"
      api_key = "key456"
    }
  }
}

ephemeral "sops_test_dynamic" "test" {
  value = terraform_data.source.output.config
}

resource "sops_encrypt" "test" {
  input_wo = {
    credentials = ephemeral.sops_test_dynamic.test.output
  }
  age_recipients = [%q]
}
`, ageRecipient)
}

func TestAccEncryptResource_ValidatorDeeplyNestedUnknown(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigValidatorDeeplyNestedUnknown(testAgePublicKeyResource),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(
							"sops_encrypt.test",
							tfjsonpath.New("output"),
						),
					},
				},
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

func testAccEncryptResourceConfigValidatorDeeplyNestedUnknown(ageRecipient string) string {
	return fmt.Sprintf(`
resource "terraform_data" "source" {
  input = {
    secret = "my-secret-value"
  }
}

ephemeral "sops_test_dynamic" "test" {
  value = terraform_data.source.output
}

resource "sops_encrypt" "test" {
  input_wo = {
    level1 = {
      level2 = {
        level3 = ephemeral.sops_test_dynamic.test.output.secret
      }
    }
  }
  age_recipients = [%q]
}
`, ageRecipient)
}

func TestAccEncryptResource_OutputIndent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigWithOutputIndent(testAgePublicKeyResource, 2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("sops_encrypt.test", "output_indent", "2"),
					testAccCheckEncryptedOutputIndentation("sops_encrypt.test", 2),
				),
			},
		},
	})
}

func TestAccEncryptResource_OutputIndentChange_ForcesReplacement(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigWithOutputIndent(testAgePublicKeyResource, 0),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccEncryptResourceConfigWithOutputIndent(testAgePublicKeyResource, 2),
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

func TestAccEncryptResource_OutputIndentWithYAML(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigWithOutputTypeAndIndent(testAgePublicKeyResource, "yaml", 4),
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
					resource.TestCheckResourceAttr("sops_encrypt.test", "output_indent", "4"),
				),
			},
		},
	})
}

func TestAccEncryptResource_OutputIndentCompact(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccEncryptResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptResourceConfigWithOutputIndent(testAgePublicKeyResource, 0),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("sops_encrypt.test", "output_indent", "0"),
					testAccCheckEncryptedOutputIndentation("sops_encrypt.test", 0),
				),
			},
		},
	})
}

func testAccEncryptResourceConfigWithOutputIndent(ageRecipient string, indent int) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
  age_recipients = [%q]
  output_indent = %d
}
`, ageRecipient, indent)
}

func testAccEncryptResourceConfigWithOutputTypeAndIndent(ageRecipient, outputType string, indent int) string {
	return fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = {
    secret = "my-secret-value"
    key    = "my-key-data"
  }
  age_recipients = [%q]
  output_type = %q
  output_indent = %d
}
`, ageRecipient, outputType, indent)
}
