package main

deny[msg] {
  input.kind == "Dockerfile"
  not user_set
  msg := "Dockerfile must set a non-root USER"
}

user_set {
  some i
  lower(trim(input.stages[_].commands[i].name)) == "user"
  val := input.stages[_].commands[i].value
  lower(trim(val)) != "root"
}
