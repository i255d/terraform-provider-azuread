---
layout: "azuread"
page_title: "Azure Active Directory: azuread_group_owner"
sidebar_current: "docs-azuread-resource-azuread-group-owner"
description: |-
  Manages a Group Owner within Azure Active Directory.

---

# azuread_group_owner

Manages a Group Owner within Azure Active Directory.

## Example Usage

```hcl
resource "azuread_group" "my_group" {
  name = "MyGroup"
}

data "azuread_user" "my_user" {
  user_principal_name = "johndoe@hashicorp.com"
}

resource "azuread_group_owner" "default_owner" {
  group_object_id = "${azuread_group.my_group.id}"
  owner_object_id = "${data.azuread_user.my_user.id}"
}
```

## Argument Reference

The following arguments are supported:

* `group_object_id` - (Required) The object id of the Azure AD Group where the Owner should be added.
* `owner_object_id` - (Required) The object id of the Azure AD User you want to add as Owner.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Azure AD Group Owner.

## Import

Azure Active Directory Group Owners can be imported using the `object id`, e.g.

```shell
terraform import azuread_group_owner.test 00000000-0000-0000-0000-000000000000/11111111-1111-1111-1111-111111111111
```

-> **NOTE:** This ID format is unique to Terraform and is composed of the Azure AD Group Object ID and the target Owner's Object ID in the format `{GroupObjectID}/{OwnerObjectID}`.
