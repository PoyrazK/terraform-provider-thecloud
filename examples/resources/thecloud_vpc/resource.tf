resource "thecloud_vpc" "example" {
  name       = "production-vpc"
  cidr_block = "10.0.0.0/16"
}
