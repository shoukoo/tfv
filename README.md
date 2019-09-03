<img src="https://github.com/shoukoo/tfv/workflows/Build%20on%20Push/badge.svg" class="image mod-full-width" /> <img src="https://img.shields.io/github/v/release/shoukoo/tfv?sort=semver" class="image mod-full-width" />

# TFV - Terraform Verifier


## HCL syntax validation

Perform a syntax check on a Terraform HCL file, verify that
there are no parse errors. Works with Terraform HCL 0.12.


## Check for missing keys

Use a YAML configuration file to specify particular HCL key/attributes
that you're expecting to find. This is useful for detecting
commonly missing Terraform attributes like eg: tags.


## Build

```
git clone https://github.com/shoukoo/tf-verifier
cd tf-verifier
go build
```


## Usage

```
tfv --debug --config config.yaml path/to/file1.tf path/to/morefiles.tf
```

## Example
```
go run main.go --config tfv.yaml test/terraform.tf test/terraform12.tf
```


## Config syntax for config.yaml

TFV only accepts the following format in the config file

```
aws_resource:
	main_key:
		- key
		- key
		- key
```

### Examples

A simple Configuration file looks like this:
```
aws_instance:
  tags:
    - Name
    - Service

  volume_tags:
    - Name
```

But you want to check if see_algorithm exists in aws_s3_bucket resource
```
terraform.tf
server_side_encryption_configuration {
	rule {
	  apply_server_side_encryption_by_default {
		kms_master_key_id = "${aws_kms_key.mykey.arn}"
		sse_algorithm     = "aws:kms"
	  }
	}
}
```

This is how you contructs the configuration file
```
aws_s3_bucket:
	server_side_encryption_configuration:
		- apply_server_side_encryption_by_default
		- kms_master_key_id
		- sse_algorithm
```
