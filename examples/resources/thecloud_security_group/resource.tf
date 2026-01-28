resource "thecloud_vpc" "main" {
  name = "main-vpc"
}

resource "thecloud_security_group" "web" {
  name        = "web-firewall"
  description = "Allow HTTP and HTTPS"
  vpc_id      = thecloud_vpc.main.id
}
