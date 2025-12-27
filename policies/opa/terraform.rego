package terraform.security

deny[msg] {
  input.resource.aws_security_group[sg]
  ingress := input.resource.aws_security_group[sg].ingress[_]
  ingress.cidr_blocks[_] == "0.0.0.0/0"
  ingress.from_port <= 22
  ingress.to_port >= 22
  msg := sprintf("Security group %s allows SSH from 0.0.0.0/0", [sg])
}
