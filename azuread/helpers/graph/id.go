package graph

import (
	"fmt"
	"net/url"
	"strings"
)

// representation of a graph Object Resource ID: objectID/type/SubObjectId
// used with group members owners or application passwords
//
type objectResouceID struct {
	Type	 string
	ObjectID string
	Path 	 map[string]string
}


func ParseObjectID(path, objectType string,) (*ObjectID, error) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	parts := strings.Split(path, "/")

	// We should have an even number of key-value pairs.
	if len(parts) % 2 != 0 {
		return nil, fmt.Errorf("The number of path segments is not divisible by 2 in %q", path)
	}


	// Put the constituent key-value pairs into a map
	m := make(map[string]string, len(parts)/2)
	for current := 0; current < len(parts); current += 2 {
		key := parts[current]
		value := parts[current+1]

		// Check key/value for empty strings.
		if key == "" || value == "" {
			return nil, fmt.Errorf("Key/Value cannot be empty strings. Key: '%s', Value: '%s'", key, value)
		}
	}

	// build the graphID from pars
	idObj := &ResourceID{}
	idObj.Path = componentMap

	if subscriptionID != "" {
		idObj.SubscriptionID = subscriptionID
	} else {
		return nil, fmt.Errorf("No subscription ID found in: %q", path)
	}

	if resourceGroup, ok := componentMap["resourceGroups"]; ok {
		idObj.ResourceGroup = resourceGroup
		delete(componentMap, "resourceGroups")
	} else {
		// Some Azure APIs are weird and provide things in lower case...
		// However it's not clear whether the casing of other elements in the URI
		// matter, so we explicitly look for that case here.
		if resourceGroup, ok := componentMap["resourcegroups"]; ok {
			idObj.ResourceGroup = resourceGroup
			delete(componentMap, "resourcegroups")
		} else {
			return nil, fmt.Errorf("No resource group name found in: %q", path)
		}
	}

	// It is OK not to have a provider in the case of a resource group
	if provider, ok := componentMap["providers"]; ok {
		idObj.Provider = provider
		delete(componentMap, "providers")
	}

	return idObj, nil
}
