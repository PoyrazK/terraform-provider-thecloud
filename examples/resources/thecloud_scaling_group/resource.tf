resource "thecloud_vpc" "main" {
  name = "main-vpc"
}

resource "thecloud_scaling_group" "web" {
  name          = "web-asg"
  vpc_id        = thecloud_vpc.main.id
  image         = "ubuntu-22.04"
  min_instances = 2
  max_instances = 10
  desired_count = 3
}
