resource "thecloud_vpc" "main" {
  name = "main-vpc"
}

resource "thecloud_load_balancer" "web" {
  name   = "web-lb"
  vpc_id = thecloud_vpc.main.id
  port   = 80
}
