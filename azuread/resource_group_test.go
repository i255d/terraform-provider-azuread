package azuread

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/ar"
)

//todo with SP/group/ application?
func TestAccAzureADGroup_basic(t *testing.T) {
	resourceName := "azuread_group.test"
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureADGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureADGroup(id),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureADGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", id)),
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

func TestAccAzureADGroup_withOwner(t *testing.T) {
	resourceName := "azuread_group.test"
	id := acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)
	password := id + "p@$$wR2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureADGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureADGroupWithOwner(id, password),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureADGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "owners.#", "1"),
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

func TestAccAzureADGroup_withMultipleOwners(t *testing.T) {
	resourceName := "azuread_group.test"
	id := acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)
	password := id + "p@$$wR2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureADGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureADGroupWithMultipleOwners(id, password),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureADGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "owners.#", "2"),
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

func testCheckAzureADGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		client := testAccProvider.Meta().(*ArmClient).groupsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if ar.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Azure AD Group %q does not exist", rs.Primary.ID)
			}
			return fmt.Errorf("Bad: Get on Azure AD groupsClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureADGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azuread_group" {
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

		return fmt.Errorf("Azure AD group still exists:\n%#v", resp)
	}

	return nil
}

func testAccAzureADGroup(id string) string {
	return fmt.Sprintf(`
resource "azuread_group" "test" {
  name = "acctest%s"
}
`, id)
}

func testAccAzureADGroupWithOwner(id string, password string) string {
	return fmt.Sprintf(`

data "azuread_domains" "tenant_domain" {
	only_initial = true
}
	
resource "azuread_user" "testA" {
	user_principal_name   = "acctestA%[1]s@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
	display_name          = "acctestA%[1]s"
	password              = "%[2]s"
}

resource "azuread_group" "test" {
	name 		= "acctest%[1]s"
	owners 	= [
		"${azuread_user.testA.id}"
	] 
}
`, id, password)
}

func testAccAzureADGroupWithMultipleOwners(id string, password string) string {
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
	name 		= "acctest%[1]s"
	owners 	= [
		"${azuread_user.testB.id}",
		"${azuread_user.testA.id}",
	]
}
`, id, password)
}
