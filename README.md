# Terraform Verifier


## HCL Syntax validation

Performs a syntax check on a terraform HCL file, verifies that
there are no parse errors. Works with terraform HCL syntax 0.12.


## Check for missing keys

Use a YAML configuration file to specify particular HCL key/values
that you're expecting to find in the HCL file. This is useful for
detecting commonly missing terraform attributes like eg: tags.


## Build

```
git clone https://github.com/shoukoo/tf-verifier
cd tf-verifier
go build
```


## Usage

```
tf-verifier --debug --config <tf.yaml> path/to/file1.tf path/to/morefiles.tf
```
