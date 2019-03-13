package azuread

import (
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/ar"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/validate"
)

func resourceGroupOwner() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupOwnerCreate,
		Read:   resourceGroupOwnerRead,
		Delete: resourceGroupOwnerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"group_object_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.UUID,
			},
			"owner_object_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.UUID,
			},
		},
	}
}

func resourceGroupOwnerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).groupsClient
	ctx := meta.(*ArmClient).StopContext

	groupID := d.Get("group_object_id").(string)
	ownerID := d.Get("owner_object_id").(string)
	tenantID := client.TenantID

	ownerGraphURL := fmt.Sprintf("https://graph.windows.net/%s/directoryObjects/%s", tenantID, ownerID)

	properties := graphrbac.AddOwnerParameters{
		URL: &ownerGraphURL,
	}

	_, err := client.AddOwner(ctx, groupID, properties)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s/%s", groupID, ownerID)
	d.SetId(id)

	return resourceGroupOwnerRead(d, meta)
}

func resourceGroupOwnerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).groupsClient
	ctx := meta.(*ArmClient).StopContext

	id := strings.Split(d.Id(), "/")
	if len(id) != 2 {
		return fmt.Errorf("ID should be in the format {groupObjectId}/{ownerObjectId} - but got %q", d.Id())
	}

	groupID := id[0]
	ownerID := id[1]

	owners, err := client.ListOwnersComplete(ctx, groupID)
	if err != nil {
		return fmt.Errorf("Error retrieving Azure AD Group owners (groupObjectId: %q): %+v", groupID, err)
	}

	var ownerDirectoryObject *graphrbac.User
	for owners.NotDone() {
		directoryObject, _ := owners.Value().AsUser()
		if directoryObject != nil {
			if *directoryObject.ObjectID == ownerID {
				ownerDirectoryObject = directoryObject
				break
			}
		}

		err = owners.NextWithContext(ctx)
		if err != nil {
			return fmt.Errorf("Error listing Azure AD Group Owners: %s", err)
		}
	}

	if ownerDirectoryObject == nil {
		log.Printf("[DEBUG] Azure AD Group Owner was not found (groupObjectId:%q / ownerObjectId:%q ) - removing from state!", groupID, ownerID)
		d.SetId("")
		return fmt.Errorf("Azure AD Group Owner not found - groupObjectId:%q / ownerObjectId:%q", groupID, ownerID)
	}

	d.Set("group_object_id", groupID)
	d.Set("owner_object_id", ownerDirectoryObject.ObjectID)

	return nil
}

func resourceGroupOwnerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).groupsClient
	ctx := meta.(*ArmClient).StopContext

	id := strings.Split(d.Id(), "/")
	if len(id) != 2 {
		return fmt.Errorf("ID should be in the format {groupObjectId}/{ownerObjectId} - but got %q", d.Id())
	}

	groupID := id[0]
	ownerID := id[1]

	resp, err := client.RemoveOwner(ctx, groupID, ownerID)

	if err != nil {
		if !ar.ResponseWasNotFound(resp) {
			return fmt.Errorf("Error removing Owner (ownerObjectId: %q) from Azure AD Group (groupObjectId: %q): %+v", ownerID, groupID, err)
		}
	}

	return nil
}
