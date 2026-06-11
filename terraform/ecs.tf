# ========== ECS コンテナインスタンス用 IAM ==========

data "aws_iam_policy_document" "ecs_instance_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ecs_instance" {
  name               = "katamichi-bot-ecs-instance"
  assume_role_policy = data.aws_iam_policy_document.ecs_instance_assume.json
}

resource "aws_iam_role_policy_attachment" "ecs_instance" {
  role       = aws_iam_role.ecs_instance.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance" {
  name = "katamichi-bot-ecs-instance"
  role = aws_iam_role.ecs_instance.name
}

# ========== タスク実行ロール: ECRプル・CloudWatchログ・SSMシークレット取得 ==========

data "aws_iam_policy_document" "ecs_task_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ecs_task_exec" {
  name               = "katamichi-bot-ecs-exec"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_assume.json
}

resource "aws_iam_role_policy_attachment" "ecs_task_exec" {
  role       = aws_iam_role.ecs_task_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

data "aws_iam_policy_document" "ecs_task_exec_ssm" {
  statement {
    actions   = ["ssm:GetParameters"]
    resources = [var.ssm_slack_token]
  }
}

resource "aws_iam_role_policy" "ecs_task_exec_ssm" {
  name   = "ssm-secrets"
  role   = aws_iam_role.ecs_task_exec.id
  policy = data.aws_iam_policy_document.ecs_task_exec_ssm.json
}

# ========== タスクロール: S3読み書き（ボット本体の権限） ==========

resource "aws_iam_role" "ecs_task" {
  name               = "katamichi-bot-ecs-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_assume.json
}

data "aws_iam_policy_document" "ecs_task" {
  statement {
    actions = ["s3:GetObject", "s3:PutObject"]
    resources = [
      "arn:aws:s3:::katamichi-go-bot-v3/state.json",
      "arn:aws:s3:::katamichi-go-bot-v3/sta/state.json",
    ]
  }
}

resource "aws_iam_role_policy" "ecs_task" {
  name   = "s3-state"
  role   = aws_iam_role.ecs_task.id
  policy = data.aws_iam_policy_document.ecs_task.json
}

# ========== ネットワーク ==========

data "aws_vpc" "default" {
  default = true
}

# EC2インスタンス用セキュリティグループ（egress のみ）
resource "aws_security_group" "ecs_instance" {
  name        = "katamichi-bot-ecs-instance"
  description = "ECS container instance for katamichi-go-bot (egress only)"
  vpc_id      = data.aws_vpc.default.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# ========== ECS コンテナインスタンス (EC2) ==========

data "aws_ami" "ecs_al2023_arm" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-ecs-kernel-5.10-hvm-*-arm64-ebs"]
  }

  filter {
    name   = "architecture"
    values = ["arm64"]
  }
}

resource "aws_instance" "ecs" {
  ami                    = data.aws_ami.ecs_al2023_arm.id
  instance_type          = "t4g.micro"
  iam_instance_profile   = aws_iam_instance_profile.ecs_instance.name
  vpc_security_group_ids = [aws_security_group.ecs_instance.id]

  # ECSクラスターへの登録
  user_data = base64encode(<<-EOF
    #!/bin/bash
    echo ECS_CLUSTER=katamichi-bot >> /etc/ecs/ecs.config
  EOF
  )

  root_block_device {
    volume_type = "gp3"
    volume_size = 30
  }

  tags = {
    Name = "katamichi-bot-ecs"
  }
}

# ========== ECS クラスター ==========

resource "aws_ecs_cluster" "bot" {
  name = "katamichi-bot"
}

# ========== CloudWatch Logs ==========

resource "aws_cloudwatch_log_group" "sta" {
  name              = "/ecs/katamichi-bot/sta"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "pro" {
  name              = "/ecs/katamichi-bot/pro"
  retention_in_days = 30
}

# ========== タスク定義 ==========

resource "aws_ecs_task_definition" "sta" {
  family                   = "katamichi-bot-sta"
  requires_compatibilities = ["EC2"]
  network_mode             = "bridge"
  memory                   = 256
  execution_role_arn       = aws_iam_role.ecs_task_exec.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  # /data/sta をEC2ホストのEBSにバインドマウント → タスク再起動でも永続
  volume {
    name      = "data"
    host_path = "/data/sta"
  }

  container_definitions = jsonencode([{
    name   = "bot"
    image  = "${aws_ecr_repository.bot.repository_url}:latest"
    memory = 256
    environment = [
      { name = "APP_ENV",    value = "sta" },
      { name = "S3_BUCKET",  value = "katamichi-go-bot-v3" },
      { name = "AWS_REGION", value = var.aws_region },
    ]
    secrets = [{
      name      = "SLACK_BOT_TOKEN"
      valueFrom = var.ssm_slack_token
    }]
    mountPoints = [{
      sourceVolume  = "data"
      containerPath = "/data"
      readOnly      = false
    }]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.sta.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "bot"
      }
    }
  }])
}

resource "aws_ecs_task_definition" "pro" {
  family                   = "katamichi-bot-pro"
  requires_compatibilities = ["EC2"]
  network_mode             = "bridge"
  memory                   = 256
  execution_role_arn       = aws_iam_role.ecs_task_exec.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  # /data/pro をEC2ホストのEBSにバインドマウント
  volume {
    name      = "data"
    host_path = "/data/pro"
  }

  container_definitions = jsonencode([{
    name   = "bot"
    image  = "${aws_ecr_repository.bot.repository_url}:latest"
    memory = 256
    environment = [
      { name = "APP_ENV",    value = "pro" },
      { name = "S3_BUCKET",  value = "katamichi-go-bot-v3" },
      { name = "AWS_REGION", value = var.aws_region },
    ]
    secrets = [{
      name      = "SLACK_BOT_TOKEN"
      valueFrom = var.ssm_slack_token
    }]
    mountPoints = [{
      sourceVolume  = "data"
      containerPath = "/data"
      readOnly      = false
    }]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.pro.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "bot"
      }
    }
  }])
}

# ========== ECS サービス ==========

resource "aws_ecs_service" "sta" {
  name            = "katamichi-bot-sta"
  cluster         = aws_ecs_cluster.bot.id
  task_definition = aws_ecs_task_definition.sta.arn
  desired_count   = 1
  launch_type     = "EC2"

  lifecycle {
    ignore_changes = [desired_count]
  }
}

resource "aws_ecs_service" "pro" {
  name            = "katamichi-bot-pro"
  cluster         = aws_ecs_cluster.bot.id
  task_definition = aws_ecs_task_definition.pro.arn
  desired_count   = 0
  launch_type     = "EC2"

  lifecycle {
    ignore_changes = [desired_count]
  }
}
