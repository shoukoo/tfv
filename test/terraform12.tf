resource "aws_instance" main {
  tags = {
    Name      = var.name
    terraform = local.terraform
  }

  volume_tags = {
    Name      = var.name
    terraform = local.terraform
  }

  count                       = var.num_instance
  associate_public_ip_address = true
  ami                         = data.aws_ami.default.image_id
  subnet_id                   = aws_subnet.main[0].id

  vpc_security_group_ids = [
    data.aws_security_group.ssh_from_vpn.id,
    aws_security_group.main.id,
  ]

  instance_type        = var.instance_type
  key_name             = "default"
  iam_instance_profile = aws_iam_instance_profile.main.id

  root_block_device {
    volume_size = 10
    volume_type = "gp2"
  }


  user_data = file("./user_data.sh")

  provisioner "local-exec" {
    command = "../wait_for_user_data.sh ${self.public_ip}"
  }
}

resource aws_s3_bucket main {

  logging {
    target_bucket = "${aws_s3_bucket.log_bucket.id}"
    target_prefix = "log/"
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = "${aws_kms_key.mykey.arn}"
        sse_algorithm     = "aws:kms"
      }
    }
  }

}
