package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const sgRuleResourceName = "thecloud_security_group_rule.test"

func TestAccSecurityGroupRuleResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	vpcName := fmt.Sprintf("sg-rule-vpc-%s", rName)
	sgName := fmt.Sprintf("test-sg-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_vpc" "rule_vpc" {
  name       = "%s"
  cidr_block = "10.0.0.0/16"
}

resource "thecloud_security_group" "rule_sg" {
  name   = "%s"
  vpc_id = thecloud_vpc.rule_vpc.id
}

resource "thecloud_security_group_rule" "test" {
  security_group_id = thecloud_security_group.rule_sg.id
  direction         = "ingress"
  protocol          = "tcp"
  port_min          = 80
  port_max          = 80
  cidr              = "0.0.0.0/0"
  priority          = 100
}
`, vpcName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(sgRuleResourceName, "direction", "ingress"),
					resource.TestCheckResourceAttr(sgRuleResourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(sgRuleResourceName, "port_min", "80"),
					resource.TestCheckResourceAttr(sgRuleResourceName, "port_max", "80"),
					resource.TestCheckResourceAttr(sgRuleResourceName, "cidr", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(sgRuleResourceName, "priority", "100"),
					resource.TestCheckResourceAttrSet(sgRuleResourceName, "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      sgRuleResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[sgRuleResourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", sgRuleResourceName)
					}
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["security_group_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}
