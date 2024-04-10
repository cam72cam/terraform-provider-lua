terraform {
  required_providers {
    tester = {
      source  = "terraform.local/local/testfunctions"
      version = "0.0.1"
    }
  }
}

output "test" {
	value = provider::tester::lua(file("./main.lua"), tomap({"foo": {"bar": toset([1,2,3])}}))
}
