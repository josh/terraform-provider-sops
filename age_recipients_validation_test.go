package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEncryptResource_EmptyAgeRecipients(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "sops_encrypt" "test" {
  input = {
    secret = "value"
  }
  age_recipients = []
}
`,
				ExpectError: regexp.MustCompile("at least 1 elements"),
			},
		},
	})
}

func TestAccEncryptResource_EmptyStringAgeRecipient(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sops_encrypt" "test" {
  input = {
    secret = "value"
  }
  age_recipients = [%q, ""]
}
`, testAgePublicKey),
				ExpectError: regexp.MustCompile("at least 1"),
			},
		},
	})
}

func TestAccEncryptDataSource_EmptyAgeRecipients(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "sops_encrypt" "test" {
  input = {
    secret = "value"
  }
  age_recipients = []
}
`,
				ExpectError: regexp.MustCompile("at least 1 elements"),
			},
		},
	})
}
