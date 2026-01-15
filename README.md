# terraform-provider-sops

Terraform provider for encrypting and decrypting data with SOPS.

## Usage

```terraform
resource "sops_encrypt" "secrets" {
  input = {
    username = "josh"
    password = "secret"
  }

  age_recipients = [
    "age1j7ce327ke8t905hr4ve97xh4jr5ujauq59nxxkr3tnz9pty78p6q26hnd0"
  ]

  output_type   = "yaml"
  output_indent = 2
}

output "secrets_yaml" {
  value = sops_encrypt.secrets.output
}
```

```terraform
provider "sops" {
  age_identity_path = "~/.config/sops/age/keys.txt"
}

data "sops_decrypt" "secrets" {
  input      = file("${path.module}/secrets.enc.yaml")
  input_type = "yaml"
}

output "password" {
  value     = data.sops_decrypt.secrets.output.password
  sensitive = true
}
```
