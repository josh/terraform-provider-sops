package main

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var agePrivateKeyRegex = regexp.MustCompile(`^AGE-SECRET-KEY-1[0-9A-Z]+$`)
var agePublicKeyRegex = regexp.MustCompile(`^age1[0-9a-z]+$`)

func TestAccAgePrivateKeyResource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "sops_age_private_key" "test" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sops_age_private_key.test",
						tfjsonpath.New("private_key"),
						knownvalue.StringRegexp(agePrivateKeyRegex),
					),
					statecheck.ExpectKnownValue(
						"sops_age_private_key.test",
						tfjsonpath.New("public_key"),
						knownvalue.StringRegexp(agePublicKeyRegex),
					),
					statecheck.ExpectSensitiveValue(
						"sops_age_private_key.test",
						tfjsonpath.New("private_key"),
					),
				},
				Check: testAccCheckAgeKeyPairMatches("sops_age_private_key.test"),
			},
		},
	})
}

func TestAccAgePrivateKeyResource_EncryptRoundTrip(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "sops_age_private_key" "test" {}

resource "sops_encrypt" "test" {
  input = {
    secret = "round-trip-value"
  }
  age_recipients = [sops_age_private_key.test.public_key]
}
`,
				Check: testAccCheckAgeKeyDecrypts(
					"sops_age_private_key.test",
					"sops_encrypt.test",
					`"secret": "round-trip-value"`,
				),
			},
		},
	})
}

func TestAccAgePrivateKeyResource_Import(t *testing.T) {
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
				ImportStateId: testAgeSecretKey,
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

func TestAccAgePrivateKeyResource_ImportInvalidKey(t *testing.T) {
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
				ImportStateId: "not-an-age-key",
				ExpectError:   regexp.MustCompile("Invalid Age Private Key"),
			},
		},
	})
}

func testAccCheckAgeKeyPairMatches(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		privateKey := rs.Primary.Attributes["private_key"]
		publicKey := rs.Primary.Attributes["public_key"]

		derived, err := deriveAgePublicKey(privateKey)
		if err != nil {
			return fmt.Errorf("generated private key is not parseable: %s", err)
		}

		if derived != publicKey {
			return fmt.Errorf("public_key %q does not match key derived from private_key %q", publicKey, derived)
		}

		return nil
	}
}

func testAccCheckAgeKeyDecrypts(keyResourceName, encryptResourceName, expectedContent string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keyRs, ok := s.RootModule().Resources[keyResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", keyResourceName)
		}

		encryptRs, ok := s.RootModule().Resources[encryptResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", encryptResourceName)
		}

		decrypted, err := decryptWithSops(context.Background(), []byte(encryptRs.Primary.Attributes["output"]), SopsDecryptOptions{
			AgeIdentityValue: keyRs.Primary.Attributes["private_key"],
			InputType:        "json",
		})
		if err != nil {
			return fmt.Errorf("failed to decrypt with generated private key: %s", err)
		}

		if !bytes.Contains(decrypted, []byte(expectedContent)) {
			return fmt.Errorf("decrypted output %q does not contain %q", decrypted, expectedContent)
		}

		return nil
	}
}
