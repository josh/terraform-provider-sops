package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAgeRecipient = "age1j7ce327ke8t905hr4ve97xh4jr5ujauq59nxxkr3tnz9pty78p6q26hnd0"

func TestAccEncryptDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigBasic(testAgeRecipient),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_NestedStructure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigNested(testAgeRecipient),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_MultipleRecipients(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigMultipleRecipients(testAgeRecipient, testAgeRecipient),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_EdgeCases(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigEdgeCases(testAgeRecipient),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_OutputTypeYAML(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigWithOutputType(testAgeRecipient, "yaml"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "output_type", "yaml"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_InvalidInputTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEncryptDataSourceConfigInvalidArray(testAgeRecipient),
				ExpectError: regexp.MustCompile(`Input must be a map/object, got \[\]interface \{\}\. SOPS can only encrypt JSON\s+objects\.`),
			},
			{
				Config:      testAccEncryptDataSourceConfigInvalidString(testAgeRecipient),
				ExpectError: regexp.MustCompile(`Input must be a map/object, got string\. SOPS can only encrypt JSON\s+objects\.`),
			},
			{
				Config:      testAccEncryptDataSourceConfigInvalidNumber(testAgeRecipient),
				ExpectError: regexp.MustCompile(`Input must be a map/object, got float64\. SOPS can only encrypt JSON\s+objects\.`),
			},
		},
	})
}

func testAccEncryptDataSourceConfigBasic(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = {
    secret = "value"
    key    = "data"
  }
  age_recipients = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigNested(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = {
    database = {
      host     = "localhost"
      password = "secret123"
    }
    api_keys = ["key1", "key2"]
  }
  age_recipients = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigMultipleRecipients(ageRecipient1, ageRecipient2 string) string {
	return `
data "sops_encrypt" "test" {
  input = {
    secret = "multi-recipient-value"
  }
  age_recipients = ["` + ageRecipient1 + `", "` + ageRecipient2 + `"]
}
`
}

func testAccEncryptDataSourceConfigInvalidArray(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = ["not", "a", "map"]
  age_recipients = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigInvalidString(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = "not a map"
  age_recipients = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigInvalidNumber(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = 42
  age_recipients = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigEdgeCases(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = {
    string_val    = "hello world"
    int_val       = 42
    float_val     = 3.14159
    bool_true     = true
    bool_false    = false
    null_val      = null
    array_strings = ["one", "two", "three"]
    array_numbers = [1, 2, 3, 4, 5]
    array_mixed   = ["string", 123, true, 45.67]
    empty_array   = []
    empty_object  = {}
    nested_deep   = {
      level1 = {
        level2 = {
          level3 = {
            value = "deep"
          }
        }
      }
    }
  }
  age_recipients = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigWithOutputType(ageRecipient, outputType string) string {
	return `
data "sops_encrypt" "test" {
  input = {
    secret = "value"
    key    = "data"
  }
  age_recipients = ["` + ageRecipient + `"]
  output_type = "` + outputType + `"
}
`
}

func TestAccEncryptDataSource_OutputIndent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigWithOutputIndent(testAgeRecipient, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "output_indent", "2"),
					testAccCheckEncryptedOutputIndentation("data.sops_encrypt.test", 2),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_OutputIndentWithYAML(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigWithOutputTypeAndIndent(testAgeRecipient, "yaml", 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "output_type", "yaml"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "output_indent", "4"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_OutputIndentCompact(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigWithOutputIndent(testAgeRecipient, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "output_indent", "0"),
					testAccCheckEncryptedOutputIndentation("data.sops_encrypt.test", 0),
				),
			},
		},
	})
}

func testAccEncryptDataSourceConfigWithOutputIndent(ageRecipient string, indent int) string {
	return fmt.Sprintf(`
data "sops_encrypt" "test" {
  input = {
    secret = "value"
    key    = "data"
  }
  age_recipients = [%q]
  output_indent = %d
}
`, ageRecipient, indent)
}

func testAccEncryptDataSourceConfigWithOutputTypeAndIndent(ageRecipient, outputType string, indent int) string {
	return fmt.Sprintf(`
data "sops_encrypt" "test" {
  input = {
    secret = "value"
    key    = "data"
  }
  age_recipients = [%q]
  output_type = %q
  output_indent = %d
}
`, ageRecipient, outputType, indent)
}
func TestAccEncryptDataSource_UnencryptedSuffix(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigWithUnencryptedSuffix(testAgeRecipient, "_unencrypted"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "unencrypted_suffix", "_unencrypted"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_EncryptedSuffix(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigWithEncryptedSuffix(testAgeRecipient, "_secret"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "encrypted_suffix", "_secret"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_UnencryptedRegex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigWithUnencryptedRegex(testAgeRecipient, "^public_"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "unencrypted_regex", "^public_"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_EncryptedRegex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigWithEncryptedRegex(testAgeRecipient, "^secret_"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "encrypted_regex", "^secret_"),
				),
			},
		},
	})
}

func testAccEncryptDataSourceConfigWithUnencryptedSuffix(ageRecipient, suffix string) string {
	return fmt.Sprintf(`
data "sops_encrypt" "test" {
  input = {
    secret          = "encrypted-value"
    public_unencrypted = "unencrypted-value"
  }
  age_recipients = [%q]
  unencrypted_suffix = %q
}
`, ageRecipient, suffix)
}

func testAccEncryptDataSourceConfigWithEncryptedSuffix(ageRecipient, suffix string) string {
	return fmt.Sprintf(`
data "sops_encrypt" "test" {
  input = {
    password_secret = "encrypted-value"
    username        = "plain-value"
  }
  age_recipients = [%q]
  encrypted_suffix = %q
}
`, ageRecipient, suffix)
}

func testAccEncryptDataSourceConfigWithUnencryptedRegex(ageRecipient, regex string) string {
	return fmt.Sprintf(`
data "sops_encrypt" "test" {
  input = {
    public_key    = "unencrypted-value"
    secret_token  = "encrypted-value"
  }
  age_recipients = [%q]
  unencrypted_regex = %q
}
`, ageRecipient, regex)
}

func testAccEncryptDataSourceConfigWithEncryptedRegex(ageRecipient, regex string) string {
	return fmt.Sprintf(`
data "sops_encrypt" "test" {
  input = {
    secret_password = "encrypted-value"
    username        = "plain-value"
  }
  age_recipients = [%q]
  encrypted_regex = %q
}
`, ageRecipient, regex)
}
