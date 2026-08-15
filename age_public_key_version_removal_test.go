package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccAgePublicKeyResource_VersionRemoval_DoesNotReplace(t *testing.T) {
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
  private_key_wo = %q
}
`, testAgeSecretKey),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sops_age_public_key.test", plancheck.ResourceActionUpdate),
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
			{
				Config: fmt.Sprintf(`
resource "sops_age_public_key" "test" {
  private_key_wo         = %q
  private_key_wo_version = 5
}
`, testAgeSecretKey),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sops_age_public_key.test", plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}
