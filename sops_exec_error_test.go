package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEncrypt_SopsBinaryMissing(t *testing.T) {
	orig := sopsBinary
	sopsBinary = "definitely-not-a-real-sops-binary"
	t.Cleanup(func() { sopsBinary = orig })

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "sops_encrypt" "test" {
  input = {
    secret = "value"
  }
  age_recipients = [%q]
}
`, testAgePublicKey),
				ExpectError: regexp.MustCompile("executable file not found"),
			},
		},
	})
}

func TestAccDecrypt_SopsBinaryMissing(t *testing.T) {
	encrypted := encryptFixture(t)

	orig := sopsBinary
	sopsBinary = "definitely-not-a-real-sops-binary"
	t.Cleanup(func() { sopsBinary = orig })

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "sops" {
  age_identity_value = %q
}

data "sops_decrypt" "test" {
  input      = %q
  input_type = "json"
}
`, testAgeSecretKey, encrypted),
				ExpectError: regexp.MustCompile("executable file not found"),
			},
		},
	})
}

func encryptFixture(t *testing.T) string {
	t.Helper()

	encrypted, err := encryptWithSops(t.Context(), map[string]interface{}{"secret": "value"}, SopsEncryptOptions{
		AgeRecipients: []string{testAgePublicKey},
	})
	if err != nil {
		t.Fatalf("failed to build encrypted fixture: %s", err)
	}

	return string(encrypted)
}
