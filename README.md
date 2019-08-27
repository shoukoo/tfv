# Terraform Verifier


## HCL syntax validation

Perform a syntax check on a Terraform HCL file, verify that
there are no parse errors. Works with Terraform HCL 0.12.


## Check for missing keys

Use a YAML configuration file to specify particular HCL key/values
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
tf-verifier --debug --config config.yaml path/to/file1.tf path/to/morefiles.tf
```


## Config syntax for config.yaml

YAML syntax should mimic the layout of the HCL file, eg:

```
aws_instance:
  tags:
    - Name
    - Service

  volume_tags:
    - Name
```
