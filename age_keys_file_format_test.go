package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func ageKeygenFileContent() string {
	return fmt.Sprintf("# created: 2026-01-01T00:00:00Z\n# public key: %s\n%s\n", testAgePublicKey, testAgeSecretKey)
}

func TestAccAgePublicKeyDataSource_KeygenFileContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "sops_age_public_key" "test" {
  private_key = %q
}
`, ageKeygenFileContent()),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.sops_age_public_key.test",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact(testAgePublicKey),
					),
				},
			},
		},
	})
}

func TestAccAgePrivateKeyResource_ImportKeygenFileContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "sops_age_private_key" "test" {}`,
			},
			{
				Config:        `resource "sops_age_private_key" "test" {}`,
				ResourceName:  "sops_age_private_key.test",
				ImportState:   true,
				ImportStateId: ageKeygenFileContent(),
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported instance, got %d", len(states))
					}
					if got := states[0].Attributes["private_key"]; got != testAgeSecretKey {
						return fmt.Errorf("imported private_key = %q, want %q", got, testAgeSecretKey)
					}
					if got := states[0].Attributes["public_key"]; got != testAgePublicKey {
						return fmt.Errorf("imported public_key = %q, want %q", got, testAgePublicKey)
					}
					return nil
				},
			},
		},
	})
}

const testAgePluginIdentity = "AGE-PLUGIN-EXAMPLE-1QGQQQKQQYVQQKQQZQPEQQQQQZQQPJQQTQQPSQYQ"

func TestAccAgePublicKeyDataSource_PluginIdentity(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "sops_age_public_key" "test" {
  private_key = %q
}
`, testAgePluginIdentity),
				ExpectError: regexp.MustCompile(`(?s)plugin identities are not\s+supported`),
			},
		},
	})
}

func TestAccAgePublicKeyResource_PluginIdentity(t *testing.T) {
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
`, testAgePluginIdentity),
				ExpectError: regexp.MustCompile(`(?s)plugin identities are not\s+supported`),
			},
		},
	})
}
