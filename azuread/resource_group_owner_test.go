package azuread

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/ar"
)

func TestAccAzureADGroupOwner_complete(t *testing.T) {
	resourceName := "azuread_group_owner.testA"
	id := acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)
	password := id + "p@$$wR2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureADGroupOwnerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureADGroupOwner(id, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "group_object_id"),
					resource.TestCheckResourceAttrSet(resourceName, "owner_object_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckAzureADGroupOwnerDestroy(s *terraform.State) error {
	var i = 0
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azuread_group_owner" {
			continue
		}

		// The Azure API throws an error if you try to remove the last owner of an
		// Azure AD Group, therefore we create two owners during testing and only
		// remove one of these owners. As the group gets deleted after testing there
		// will be no orphaned objects from these resource tests.

		if i > 0 {
			// we aleady deleted one of the azuread_group_owners, skip the current resource
			continue
		}

		client := testAccProvider.Meta().(*ArmClient).groupsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if ar.ResponseWasNotFound(resp.Response) {
				return nil
			}

			return err
		}

		return fmt.Errorf("Azure AD group owner still exists:\n%#v", resp)
	}

	return nil
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
