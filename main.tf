# This is a terraform script to provision high availability redundant cron servers
#
# Ensure terraform.tfstate is committed.

variable "name" {
  default = "operations-cron"
}

variable "instance_type" {
  default = "r4.large"
}

variable "num_instances" {
  default = "3"
}

resource "aws_route53_record" "main" {
  count   = var.num_instances
  zone_id = data.aws_route53_zone.camplexer.zone_id
  name    = format("cron%02d.camplexer.com", count.index + 1)
  type    = "A"
  ttl     = "60"
  records = [element(aws_instance.main.*.private_ip, count.index)]
}

# create new server
resource "aws_instance" "main" {
  tags = {
    Name      = "tes"
    product   = "operations"
    service   = "schedule"
    function  = "operations"
    terraform = local.terraform
    warn1     = "above 75 swap for 30"
    warn2     = "above 95 disk for 30"
    warn3     = "above 99 cpu for 60"
  }

  count         = var.num_instances
  ami           = data.aws_ami.default.id
  instance_type = var.instance_type
  subnet_id     = element(aws_subnet.main.*.id, count.index % 3)

  vpc_security_group_ids = [
    data.aws_security_group.ops-gateway-to-ssh-operations-vpc.id,
    data.aws_security_group.buildkite-to-ssh-operations-vpc.id,
  ]

  # "${data.aws_security_group.buildkite-to-ssh-operations-vpc.id}" # we can probably get away with not having this

  key_name                    = "bootstrap"
  iam_instance_profile        = aws_iam_instance_profile.main.id
  associate_public_ip_address = false
  root_block_device {
    volume_size = 128
    volume_type = "gp2"
  }
  lifecycle {
    create_before_destroy = true
    ignore_changes = [
      ami,
      user_data,
      instance_type,
    ]
  }
  provisioner "local-exec" {
    command = "wait_for_user_data.sh ${self.id}"
  }
}

resource "aws_security_group" "db" {
  tags = {
    Name      = "operations-crondb"
    terraform = local.terraform
  }

  name        = "operations-crondb"
  description = "rules for cronlock redis server"
  vpc_id      = data.aws_vpc.operations.id

  ingress {
    from_port = 6379
    to_port   = 6379
    protocol  = "tcp"

    cidr_blocks = [
      "172.19.2.0/24",
    ]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

// Breaking up the /24 network into 3 separate subnets
resource "aws_subnet" "main" {
  tags = {
    Name      = "Operations Cron ${format("%01d", count.index + 1)}"
    terraform = local.terraform
  }

  vpc_id     = data.aws_vpc.operations.id
  cidr_block = cidrsubnet(local.operations_cron_cidr, 2, count.index % 3)
  availability_zone = element(
    split(",", "ap-southeast-2a,ap-southeast-2b,ap-southeast-2c"),
    count.index % 3,
  )
  map_public_ip_on_launch = false
  count                   = "3"
}

resource "aws_route_table_association" "master" {
  subnet_id = element(aws_subnet.main.*.id, count.index)
  route_table_id = element(
    [
      data.aws_route_table.nat_a.id,
      data.aws_route_table.nat_b.id,
      data.aws_route_table.nat_c.id,
    ],
    count.index,
  )
  count = "3"
}

resource "aws_security_group" "ssh" {
  tags = {
    Name      = "operations-cron-http"
    terraform = local.terraform
  }

  name        = "operations-cron-http"
  description = "give cron servers http access to particular resources (es proxies)"
  vpc_id      = data.aws_vpc.service.id

  ingress {
    from_port = 80
    to_port   = 80
    protocol  = "tcp"

    cidr_blocks = [
      "172.19.2.0/24",
    ]
  }
}

resource "aws_security_group" "http_identity" {
  tags = {
    Name      = "identity-cron-http"
    terraform = local.terraform
  }

  name        = "identity-cron-http"
  description = "give cron servers http access in Identity VPC"
  vpc_id      = data.aws_vpc.identity.id

  ingress {
    from_port = 80
    to_port   = 80
    protocol  = "tcp"

    cidr_blocks = [
      "172.19.2.0/24",
    ]
  }
}

resource "aws_security_group" "sftp" {
  tags = {
    Name      = "identity-cron-sftp"
    terraform = local.terraform
  }

  name        = "identity-cron-sftp"
  description = "give cron servers sftp access"
  vpc_id      = data.aws_vpc.identity.id

  ingress {
    from_port = 122
    to_port   = 122
    protocol  = "tcp"

    cidr_blocks = [
      "172.19.2.0/24",
    ]
  }
}

resource "aws_security_group" "rabbitmq" {
  tags = {
    Name      = "operations-cron-rabbitmq"
    terraform = local.terraform
  }

  name        = "operations-cron-rabbitmq"
  description = "give cron servers rabbitmq access to particular queue servers"
  vpc_id      = data.aws_vpc.service.id

  ingress {
    from_port = 15672
    to_port   = 15672
    protocol  = "tcp"

    cidr_blocks = [
      "172.19.2.0/24",
    ]
  }
}

resource "aws_security_group" "rabbitmqred" {
  tags = {
    Name      = "operations-cron-rabbitmq"
    terraform = local.terraform
  }

  name        = "operations-cron-rabbitmq"
  description = "give cron servers rabbitmq access to particular queue servers"
  vpc_id      = data.aws_vpc.identity.id

  ingress {
    from_port = 15672
    to_port   = 15672
    protocol  = "tcp"

    cidr_blocks = [
      "172.19.2.0/24",
    ]
  }
}

resource "aws_security_group" "ssh_service" {
  tags = {
    Name      = "service-cron-ssh"
    terraform = local.terraform
  }

  name        = "service-cron-ssh"
  description = "give cron servers ssh access"
  vpc_id      = data.aws_vpc.service.id

  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"

    cidr_blocks = [
      "172.19.2.0/24",
    ]
  }
}

# setup the server's perms so that it can access the s3 buckets
resource "aws_iam_instance_profile" "main" {
  name       = "operations-cron"
  role       = aws_iam_role.main.id
  depends_on = [aws_iam_role.main]
}

# permit the box to assume a role
resource "aws_iam_role" "main" {
  name = "operations-cron"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
              "Service": ["ec2.amazonaws.com"]
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF

}

# access to the lexer-event-data bucket to list objects for historical-enrichment reports
resource "aws_iam_role_policy" "event_data" {
  name = "historical-enrichment-report-read-access"
  role = aws_iam_role.main.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action":["s3:List*"],
      "Resource":[
        "arn:aws:s3:::lexer-event-data",
        "arn:aws:s3:::lexer-event-data/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": ["sqs:Get*"],
      "Resource": "${data.aws_sqs_queue.lexer-historical-enrichment-timeline-loading.arn}"
    }
  ]
}
EOF

}

# access to the enrichment bucket to publish attribute metadata exports
resource "aws_iam_role_policy" "enrichment" {
name = "s3-read-access-enrichment"
role = aws_iam_role.main.id

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["s3:Get*","s3:List*","s3:Head*","s3:Put*"],
      "Effect": "Allow",
      "Resource":[
        "arn:aws:s3:::lexer-enrichment-tasks",
        "arn:aws:s3:::lexer-enrichment-tasks/*"
      ]
    }
  ]
}
EOF

}

# access to lexer-client-amobee bucket for exports
resource "aws_iam_role_policy" "amobee" {
name = "s3-put-access-amobee"
role = aws_iam_role.main.id

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["s3:Put*"],
      "Effect": "Allow",
      "Resource":[
        "arn:aws:s3:::lexer-client-amobee/outgoing/*"
      ]
    }
  ]
}
EOF

}

# access to optus-development-cloudera-data bucket for exports
resource "aws_iam_role_policy" "optus" {
  name = "s3-put-access-optus-cloudera"
  role = aws_iam_role.main.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["s3:Put*"],
      "Effect": "Allow",
      "Resource":[
        "arn:aws:s3:::optus-development-cloudera-data/Lexer/*"
      ]
    }
  ]
}
EOF

}

resource "aws_iam_role_policy" "mevo" {
  name = "s3-put-access-optus-mevo"
  role = aws_iam_role.main.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["s3:Put*"],
      "Effect": "Allow",
      "Resource":[
        "arn:aws:s3:::lexer-client-optus-mevo",
        "arn:aws:s3:::lexer-client-optus-mevo/*"
      ]
    }
  ]
}
EOF

}

# access to lexer-client-punters bucket for exports
resource "aws_iam_role_policy" "punters" {
name = "s3-put-access-punters"
role = aws_iam_role.main.id

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["s3:Put*"],
      "Effect": "Allow",
      "Resource":[
        "arn:aws:s3:::lexer-client-punters/outbox/*"
      ]
    }
  ]
}
EOF

}

resource "aws_iam_role_policy" "identity-etl" {
name = "s3-access-identity-etl"
role = aws_iam_role.main.id

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["s3:PutObject", "s3:ListBucket", "s3:GetObject"],
      "Effect": "Allow",
      "Resource":["arn:aws:s3:::lexer-client-sftp-australia",
                  "arn:aws:s3:::lexer-client-sftp-australia/*"]
    }
  ]
}
EOF

}

# access the specific s3 buckets for the wallboard functionality
resource "aws_iam_role_policy" "wb" {
  name = "s3-write-access-wallboard"
  role = aws_iam_role.main.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["s3:GetObject","s3:PutObject","s3:ListBucket", "s3:AbortMultipartUpload", "s3:ListMultipartUploadParts", "s3:ListBucketMultipartUploads"],
      "Effect": "Allow",
      "Resource":["arn:aws:s3:::lexer-wallboard",
                  "arn:aws:s3:::lexer-wallboard/*"]
    }
  ]
}
EOF

}

# access the lexer-client-liveramp s3 buckets
resource "aws_iam_role_policy" "liveramp" {
  name = "s3-put-access-liveramp"
  role = aws_iam_role.main.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["s3:PutObject","s3:ListBucket"],
      "Effect": "Allow",
      "Resource":["arn:aws:s3:::lexer-client-liveramp",
                  "arn:aws:s3:::lexer-client-liveramp/*"]
    }
  ]
}
EOF

}

# allow cron to put files into lexer-client buckets
resource "aws_iam_role_policy" "cron-s3" {
name = "s3-put-access-lexer-client"
role = aws_iam_role.main.id

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["s3:PutObject*","s3:ListBucket"],
      "Effect": "Allow",
      "Resource":["arn:aws:s3:::lexer-client-*/sftp/exports/*",
                  "arn:aws:s3:::lexer-client-*/sftp/data/outbox/*",
                  "arn:aws:s3:::lexer-client-*/sftp/data/GPT/*"]
    }
  ]
}
EOF

}

# perms for lexer-jobs
resource "aws_iam_role_policy" "monkey" {
name = "stop-start-instances"
role = aws_iam_role.main.id

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":[
        "ec2:StartInstances",
        "ec2:StopInstances",
        "ec2:DescribeInstances",
        "ec2:DeregisterImage",
        "ec2:DeleteSnapshot"
      ],
      "Effect": "Allow",
      "Resource":"*"
    }
  ]
}
EOF

}

# perm for ES scaling scripts
resource "aws_iam_role_policy" "es" {
  name = "create-instance-propogate-role"
  role = aws_iam_role.main.id

  # PassRole ensures that newly created instances inherit permissions
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":["ec2:RunInstances"],
      "Effect": "Allow",
      "Resource":"*"
    },
    {
      "Effect":"Allow",
      "Action":"iam:PassRole",
      "Resource":"*"
    }
  ]
}
EOF

}

# perm for creating dynamodb tables for global identity and scaling
resource "aws_iam_role_policy" "dynamodb" {
  name = "create-dynamodb-role"
  role = aws_iam_role.main.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":[
        "dynamodb:*",
        "lambda:CreateEventSourceMapping",
        "lambda:InvokeFunction",
        "lambda:GetEventSourceMapping"
      ],
      "Effect": "Allow",
      "Resource":"*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:DescribeTable",
        "dynamodb:ListTables",
        "dynamodb:UpdateTable",
        "cloudwatch:GetMetricStatistics"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sns:Publish"
      ],
      "Resource": [
        "arn:aws:sns:*::dynamic-dynamodb"
      ]
    }
  ]
}
EOF

}

# configure iam policy for unicreds
resource "aws_iam_role_policy" "credentials_read_cron" {
name = "credentials_read_cron"
role = aws_iam_role.main.id

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "kms:Decrypt"
      ],
      "Effect": "Allow",
      "Resource": "${data.aws_kms_key.secrets.arn}",
      "Condition": {
          "StringEquals": {
            "kms:EncryptionContext:app": "${var.name}"
          }
       }
    },
    {
      "Action": [
        "dynamodb:GetItem", "dynamodb:Query"
      ],
      "Effect": "Allow",
      "Resource": "${data.aws_dynamodb_table.secrets.arn}"
    }
  ]
}
EOF

}

# perm for new relic aws monitoring (gekkie/newrelic-cloudwatch container)
resource "aws_iam_role_policy" "nr" {
name = "cron-new-relic"
role = aws_iam_role.main.id

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "autoscaling:Describe*",
        "cloudwatch:Describe*",
        "cloudwatch:List*",
        "cloudwatch:Get*",
        "ec2:Describe*",
        "ec2:Get*",
        "ec2:ReportInstanceStatus",
        "elasticache:DescribeCacheClusters",
        "elasticloadbalancing:Describe*",
        "sqs:GetQueueAttributes",
        "sqs:ListQueues",
        "rds:DescribeDBInstances",
        "iam:DeleteAccessKey",
        "ce:GetCostAndUsage",
        "SNS:ListTopics"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF

}

# call actions to admin acc
resource "aws_iam_role_policy" "call_action_to_admin_acc" {
  name = "call_action_to_admin_acc"
  role = aws_iam_role.main.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ce:GetCostForecast",
        "sts:AssumeRole"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:iam::956318495558:role/get-aws-cost-forecast"
    }
  ]
}
EOF

}

# access for prowler AWS security tool
resource "aws_iam_role_policy_attachment" "prowler" {
  role = aws_iam_role.main.id
  policy_arn = "arn:aws:iam::aws:policy/SecurityAudit"
}

# prowler
resource "aws_iam_role_policy" "prowler" {
  name = "prowler"
  role = aws_iam_role.main.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action":[
        "ses:SendRawEmail"
      ],
      "Effect": "Allow",
      "Resource":"*"
    }
  ]
}
EOF

}

resource "aws_iam_role_policy_attachment" "default" {
role       = aws_iam_role.main.name
policy_arn = data.aws_iam_policy.default.arn
}

resource "aws_iam_role_policy" "dashboard-system-api-token-operations-cron" {
name = "dashboard-system-api-token-operations-cron"
role = aws_iam_role.main.id

policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": [
        "${data.aws_secretsmanager_secret.dashboard-system-api-token-operations-cron.arn}"
      ]
    }
  ]
}
EOF

}

