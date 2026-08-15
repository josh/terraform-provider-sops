package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccAgePublicKeyDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "sops_age_public_key" "test" {
  private_key = %q
}
`, testAgeSecretKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.sops_age_public_key.test",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact(testAgePublicKey),
					),
					statecheck.ExpectSensitiveValue(
						"data.sops_age_public_key.test",
						tfjsonpath.New("private_key"),
					),
				},
			},
		},
	})
}

func TestAccAgePublicKeyDataSource_InvalidPrivateKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "sops_age_public_key" "test" {
  private_key = "not-an-age-key"
}
`,
				ExpectError: regexp.MustCompile("Invalid Age Private Key"),
			},
		},
	})
}
