package graph

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-uuid"
)

type objectSubResourceId struct {
	guid1 string
	guid2 string
}

func (id objectSubResourceId) String() string {
	return id.guid1 + "/" + id.guid2
}

func objectSubResourceIdParse(id string) (objectSubResourceId, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return objectSubResourceId{}, fmt.Errorf("Object Subresource ID should be in the format {guid1}/{guid2} - but got %q", id)
	}

	if _, err := uuid.ParseUUID(parts[0]); err != nil {
		return objectSubResourceId{}, fmt.Errorf("guid1 in {guid1}/{guid2} (%q) is not a valid GUID: %+v", id[0], err)
	}

	if _, err := uuid.ParseUUID(parts[1]); err != nil {
		return objectSubResourceId{}, fmt.Errorf("guid2 in {guid1}/{guid2} (%q) is not a valid GUID: %+v", id[0], err)
	}

	return objectSubResourceId{
		guid1: parts[0],
		guid2: parts[1],
	}, nil

}
