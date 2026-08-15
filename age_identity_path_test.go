package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccAgeIdentityPathConfig(identityPath string) string {
	return fmt.Sprintf(`
provider "sops" {
  age_identity_path = %q
}

data "sops_encrypt" "test" {
  input = {
    secret = "identity-path-value"
  }
  age_recipients = [%q]
}

data "sops_decrypt" "test" {
  input      = data.sops_encrypt.test.output
  input_type = "json"
}
`, identityPath, testAgePublicKey)
}

func TestAccDecrypt_AgeIdentityPathMissingFile(t *testing.T) {
	// An ambient key must not silently take over when the configured
	// identity file does not exist.
	t.Setenv("SOPS_AGE_KEY", testAgeSecretKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAgeIdentityPathConfig("~/definitely-not-here/keys.txt"),
				ExpectError: regexp.MustCompile("[Aa]ge identity file"),
			},
		},
	})
}

func TestAccDecrypt_AgeIdentityPathTildeExpansion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	keyFile := filepath.Join(home, "keys.txt")
	if err := os.WriteFile(keyFile, []byte(testAgeSecretKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgeIdentityPathConfig("~/keys.txt"),
				Check: resource.TestCheckResourceAttr(
					"data.sops_decrypt.test", "output.secret", "identity-path-value",
				),
			},
		},
	})
}

func TestAccDecrypt_AgeIdentityPathPermissionDenied(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	locked := filepath.Join(home, "locked")
	if err := os.Mkdir(locked, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "keys.txt"), []byte(testAgeSecretKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })

	if _, err := os.Stat(filepath.Join(locked, "keys.txt")); err == nil {
		t.Skip("filesystem permissions are not enforced in this environment (e.g. running as root)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAgeIdentityPathConfig("~/locked/keys.txt"),
				ExpectError: regexp.MustCompile(`(?s)failed to access age identity\s+file`),
			},
		},
	})
}
