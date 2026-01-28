resource "thecloud_vpc" "main" {
  name = "main-vpc"
}

resource "thecloud_instance" "web" {
  name   = "web-server"
  image  = "ubuntu-22.04"
  vpc_id = thecloud_vpc.main.id
  ports  = "80:80,443:443"
}
