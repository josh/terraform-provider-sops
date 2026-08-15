package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// echoedAgeKeyPairMatchCheck asserts that the public_key echoed into the given
// resource's "data" attribute is derived from the echoed private_key.
type echoedAgeKeyPairMatchCheck struct {
	resourceAddress string
}

func (c echoedAgeKeyPairMatchCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	if req.State == nil || req.State.Values == nil || req.State.Values.RootModule == nil {
		resp.Error = fmt.Errorf("state is missing values")
		return
	}

	for _, rc := range req.State.Values.RootModule.Resources {
		if rc.Address != c.resourceAddress {
			continue
		}

		data, ok := rc.AttributeValues["data"].(map[string]interface{})
		if !ok {
			resp.Error = fmt.Errorf("%s data attribute is not an object: %T", c.resourceAddress, rc.AttributeValues["data"])
			return
		}

		privateKey, ok := data["private_key"].(string)
		if !ok {
			resp.Error = fmt.Errorf("%s data.private_key is not a string", c.resourceAddress)
			return
		}

		publicKey, ok := data["public_key"].(string)
		if !ok {
			resp.Error = fmt.Errorf("%s data.public_key is not a string", c.resourceAddress)
			return
		}

		derived, err := deriveAgePublicKey(privateKey)
		if err != nil {
			resp.Error = fmt.Errorf("generated private key is not parseable: %s", err)
			return
		}

		if derived != publicKey {
			resp.Error = fmt.Errorf("public_key %q does not match key derived from private_key %q", publicKey, derived)
			return
		}

		return
	}

	resp.Error = fmt.Errorf("%s not found in state", c.resourceAddress)
}

func TestAccAgePrivateKeyEphemeralResource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "sops_age_private_key" "test" {}

provider "echo" {
  data = {
    private_key = ephemeral.sops_age_private_key.test.private_key
    public_key  = ephemeral.sops_age_private_key.test.public_key
  }
}

resource "echo" "test" {}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("private_key"),
						knownvalue.StringRegexp(agePrivateKeyRegex),
					),
					statecheck.ExpectKnownValue(
						"echo.test",
						tfjsonpath.New("data").AtMapKey("public_key"),
						knownvalue.StringRegexp(agePublicKeyRegex),
					),
					echoedAgeKeyPairMatchCheck{resourceAddress: "echo.test"},
				},
			},
		},
	})
}
