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
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccAgePublicKeyResource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sops_age_public_key" "test" {
  private_key_wo = %q
}
`, testAgeSecretKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_age_public_key.test",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact(testAgePublicKey),
					),
					statecheck.ExpectKnownValue(
						"sops_age_public_key.test",
						tfjsonpath.New("private_key_wo"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

func TestAccAgePublicKeyResource_VersionBump_ForcesReplacement(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sops_age_public_key" "test" {
  private_key_wo         = %q
  private_key_wo_version = 1
}
`, testAgeSecretKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_age_public_key.test",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact(testAgePublicKey),
					),
				},
			},
			{
				Config: fmt.Sprintf(`
resource "sops_age_public_key" "test" {
  private_key_wo         = %q
  private_key_wo_version = 2
}
`, testAgeSecretKey2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sops_age_public_key.test", plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_age_public_key.test",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact(testAgePublicKey2),
					),
				},
			},
		},
	})
}

func TestAccAgePublicKeyResource_KeyChangeWithoutVersionBump_NoOp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sops_age_public_key" "test" {
  private_key_wo         = %q
  private_key_wo_version = 1
}
`, testAgeSecretKey),
			},
			{
				Config: fmt.Sprintf(`
resource "sops_age_public_key" "test" {
  private_key_wo         = %q
  private_key_wo_version = 1
}
`, testAgeSecretKey2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sops_age_public_key.test", plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_age_public_key.test",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact(testAgePublicKey),
					),
				},
			},
		},
	})
}

func TestAccAgePublicKeyResource_InvalidKeyWithoutVersionBump_FailsAtPlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sops_age_public_key" "test" {
  private_key_wo         = %q
  private_key_wo_version = 1
}
`, testAgeSecretKey),
			},
			{
				// Even though a key change without a version bump plans as NoOp,
				// the malformed key must still be rejected at plan time.
				Config: `
resource "sops_age_public_key" "test" {
  private_key_wo         = "not-an-age-key"
  private_key_wo_version = 1
}
`,
				ExpectError: regexp.MustCompile("Invalid Age Private Key"),
			},
			{
				// Restore a valid key so the post-test destroy plan passes validation.
				Config: fmt.Sprintf(`
resource "sops_age_public_key" "test" {
  private_key_wo         = %q
  private_key_wo_version = 1
}
`, testAgeSecretKey),
			},
		},
	})
}

func TestAccAgePublicKeyResource_InvalidPrivateKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "sops_age_public_key" "test" {
  private_key_wo = "not-an-age-key"
}
`,
				ExpectError: regexp.MustCompile("Invalid Age Private Key"),
			},
		},
	})
}

func TestAccAgePublicKeyResource_FromEphemeralKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// The flagship flow: generate a key pair ephemerally, persist only
				// the public key in state, and encrypt against it. Decrypting later
				// requires the ephemeral private key to have been written to a real
				// secret store; a fresh pair is generated on every run.
				Config: `
ephemeral "sops_age_private_key" "test" {}

resource "sops_age_public_key" "test" {
  private_key_wo         = ephemeral.sops_age_private_key.test.private_key
  private_key_wo_version = 1
}

data "sops_encrypt" "test" {
  input = {
    secret = "encrypted-for-ephemeral-key"
  }
  age_recipients = [sops_age_public_key.test.public_key]
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_age_public_key.test",
						tfjsonpath.New("public_key"),
						knownvalue.StringRegexp(agePublicKeyRegex),
					),
					statecheck.ExpectKnownValue(
						"data.sops_encrypt.test",
						tfjsonpath.New("output"),
						knownvalue.StringRegexp(regexp.MustCompile(`"sops"`)),
					),
				},
			},
		},
	})
}
