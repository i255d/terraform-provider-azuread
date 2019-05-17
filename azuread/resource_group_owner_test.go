package azuread

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureADGroupOwner_complete(t *testing.T) {
	resourceName := "azuread_group_owner.testA"
	id := acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)
	password := id + "p@$$wR2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureADGroupOwner(id, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "group_object_id"),
					resource.TestCheckResourceAttrSet(resourceName, "owner_object_id"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAzureADGroupOwner(id string, password string) string {
	return fmt.Sprintf(`

data "azuread_domains" "tenant_domain" {
	only_initial = true
}

resource "azuread_user" "testA" {
	user_principal_name   = "acctestA%[1]s@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
	display_name          = "acctestA%[1]s"
	password              = "%[2]s"
}

resource "azuread_user" "testB" {
	user_principal_name   = "acctestB%[1]s@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
	display_name          = "acctestB%[1]s"
	password              = "%[2]s"
}
	
resource "azuread_group" "test" {
	name = "acctest%[1]s"
}

resource "azuread_group_owner" "testA" {
	group_object_id = "${azuread_group.test.id}"
	owner_object_id = "${azuread_user.testA.id}"
}

resource "azuread_group_owner" "testB" {
	group_object_id = "${azuread_group.test.id}"
	owner_object_id = "${azuread_user.testB.id}"
}

`, id, password)
}
