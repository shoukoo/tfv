run:
	go run main.go --debug --config tfv.yaml test/terraform.tf test/terraform12.tf

test:
	go test -v

.PHONY: run test
