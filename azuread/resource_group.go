package azuread

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/ar"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/guid"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/p"
	`github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/tf`
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/validate"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"owners": {
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
				MinItems: 1,
				MaxItems: 100, //Group owners are maxed out at 100
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.UUID,
				},
			},
		},
	}
}

func resourceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).groupsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	tenantID := client.TenantID

	// first create group
	properties := graphrbac.GroupCreateParameters{
		DisplayName:     &name,
		MailEnabled:     p.Bool(false),           //we're defaulting to false, as the API currently only supports the creation of non-mail enabled security groups.
		MailNickname:    p.String(guid.New().String()), //this matches the portal behavior
		SecurityEnabled: p.Bool(true),            //we're defaulting to true, as the API currently only supports the creation of non-mail enabled security groups.
	}

	// todo require resources to be imported
	group, err := client.Create(ctx, properties)
	if err != nil {
		return fmt.Errorf("error creating group %q", name)
	}
	if group.ObjectID == nil {
		return fmt.Errorf("objectID is nil for group %q", name)
	}

	// we have to make a request for each owner we want to add to the group
	for _, owner := range tf.ExpandStringArray(d.Get("owners").(*schema.Set).List()) {
		add := graphrbac.AddOwnerParameters{
			URL: p.String(fmt.Sprintf("https://graph.windows.net/%s/directoryObjects/%s", tenantID, owner)),
		}

		if _, err := client.AddOwner(ctx, *group.ObjectID, add); err != nil {
			return fmt.Errorf("erro adding owner %q to group %q", owner, name)
		}
	}

	d.SetId(*group.ObjectID)
	return resourceGroupRead(d, meta)
}

func resourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).groupsClient
	ctx := meta.(*ArmClient).StopContext
	

	resp, err := client.Get(ctx, d.Id())
	if err != nil {
		if ar.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Azure AD group with id %q was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving Azure AD Group with ID %q: %+v", d.Id(), err)
	}

	d.Set("name", resp.DisplayName)

	respOwners, err := client.ListOwnersComplete(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving owners for Azure AD Group %q (ID %q): %+v", name, d.Id(), err)
	}

	ownersFlat, err := flattenGroupOwners(meta, respOwners)
	if err != nil {
		return fmt.Errorf("Error flattening `owners` for Azure AD group %q: %+v", *resp.DisplayName, err)
	}
	d.Set("owners", ownersFlat)
	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).groupsClient
	ctx := meta.(*ArmClient).StopContext

	objectID := d.Id()
	tenantID := client.TenantID

	if d.HasChange("owners") {

		owners := d.Get("owners").([]interface{})
		ownersExpanded, err := expandGroupOwners(owners)
		if err != nil {
			return fmt.Errorf("Error expanding `owners`: %+v", owners)
		}

		currentOwnersListIterator, err := client.ListOwnersComplete(ctx, objectID)
		if err != nil {
			return fmt.Errorf("Error retrieving Azure AD Group owners (groupObjectId: %q): %+v", objectID, err)
		}

		currentOwners := make([]string, 0)
		for currentOwnersListIterator.NotDone() {
			user, _ := currentOwnersListIterator.Value().AsUser()
			currentOwners = append(currentOwners, *user.ObjectID)
			if err := currentOwnersListIterator.NextWithContext(ctx); err != nil {
				return fmt.Errorf("Error listing Azure AD Group Owners: %s", err)
			}
		}

		//first we loop through all expanded owners and add them if necessary. We do the add/remove in separate loops as
		//we want to prevent that we're removing the last owner from the group which will cause an error in the Azure API.
		if ownersExpanded != nil {
			for _, expandedOwner := range *ownersExpanded {

				//check if the user is already in the list of current owners
				var alreadyOwner = false
				for _, currentOwner := range currentOwners {
					if expandedOwner == currentOwner {
						alreadyOwner = true
					}
				}

				if !alreadyOwner {
					log.Printf("[DEBUG] Adding %q as owner of group %q.", expandedOwner, objectID)
					ownerGraphURL := fmt.Sprintf("https://graph.windows.net/%s/directoryObjects/%s", tenantID, expandedOwner)
					addOwnerProperties := graphrbac.AddOwnerParameters{
						URL: &ownerGraphURL,
					}

					if _, err := client.AddOwner(ctx, objectID, addOwnerProperties); err != nil {
						return err
					}
				}
			}
		}

		//loop through all current owners of the group
		for _, currentOwner := range currentOwners {

			currentOwnerGraphURL := fmt.Sprintf("https://graph.windows.net/%s/directoryObjects/%s", tenantID, currentOwner)

			//check if the current owner should be kept or removed
			var keep = false
			for _, expandedOwner := range *ownersExpanded {

				ownerGraphURL := fmt.Sprintf("https://graph.windows.net/%s/directoryObjects/%s", tenantID, expandedOwner)
				if ownerGraphURL == currentOwnerGraphURL {
					keep = true
				}

				if !keep {
					//the owner should be removed from the list of group owners
					log.Printf("[DEBUG] Removing owner %q from group %q.", currentOwner, objectID)
					resp, err := client.RemoveOwner(ctx, objectID, currentOwner)
					if err != nil {
						if !ar.ResponseWasNotFound(resp) {
							return fmt.Errorf("Error removing Owner (ownerObjectId: %q) from Azure AD Group (groupObjectId: %q): %+v", expandedOwner, objectID, err)
						}
					}
				}
			}
		}
	}

	return resourceGroupRead(d, meta)
}

func resourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).groupsClient
	ctx := meta.(*ArmClient).StopContext

	if resp, err := client.Delete(ctx, d.Id()); err != nil {
		if !ar.ResponseWasNotFound(resp) {
			return fmt.Errorf("Error Deleting Azure AD Group with ID %q: %+v", d.Id(), err)
		}
	}

	return nil
}

func expandGroupOwners(input []interface{}) (*[]string, error) {
	output := make([]string, 0)

	for _, owner := range input {
		output = append(output, owner.(string))
	}

	return &output, nil
}

func flattenGroupOwners(meta interface{}, owners graphrbac.DirectoryObjectListResultIterator) ([]interface{}, error) {
	ctx := meta.(*ArmClient).StopContext
	result := make([]interface{}, 0)

	for owners.NotDone() {
		user, _ := owners.Value().AsUser()
		if user != nil {
			result = append(result, *user.ObjectID)
		}

		if err := owners.NextWithContext(ctx); err != nil {
			return nil, fmt.Errorf("Error listing Azure AD Group Owners: %s", err)
		}
	}

	return result, nil
}
