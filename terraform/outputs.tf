output "ecr_repository_url" {
  description = "ECRリポジトリURL"
  value       = aws_ecr_repository.bot.repository_url
}

output "ecs_instance_public_ip" {
  description = "ECSコンテナインスタンスのパブリックIP"
  value       = aws_instance.ecs.public_ip
}
