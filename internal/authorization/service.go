package authorization

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	_ "embed"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
)

//go:embed schema/schema.zed
var schema string

func NewService(
	authzedClient *authzed.Client,
) *Service {
	_, err := authzedClient.WriteSchema(context.TODO(), &v1.WriteSchemaRequest{
		Schema: string(schema),
	})
	if err != nil {
		log.Fatalf("Failed to apply schema to spiceDB")
	}
	return &Service{
		authzed: authzedClient,
	}
}

func newObject(resourceType, resourceId string) *v1.ObjectReference {
	return &v1.ObjectReference{
		ObjectType: resourceType,
		ObjectId:   resourceId,
	}
}

func newSubject(subjectType, subjectId string) *v1.SubjectReference {
	return &v1.SubjectReference{
		Object: newObject(subjectType, subjectId),
	}
}

func (svc *Service) WriteRelationship(
	ctx context.Context,
	resourceType, resourceID string,
	relation string,
	subjectType, subjectID string,
) (string, error) {
	relationship := &v1.Relationship{
		Resource: newObject(resourceType, resourceID),
		Relation: relation,
		Subject:  newSubject(subjectType, subjectID),
	}
	res, err := svc.authzed.WriteRelationships(
		ctx,
		&v1.WriteRelationshipsRequest{
			Updates: []*v1.RelationshipUpdate{
				{
					Operation:    v1.RelationshipUpdate_OPERATION_TOUCH,
					Relationship: relationship,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return res.WrittenAt.Token, nil
}

func (svc *Service) CheckPermOnResource(
	ctx context.Context,
	subjectType, subjectID string,
	resourceType, resourceID string,
	permission string,
	precise bool,
	zedToken string,
) (bool, error) {
	var consistency *v1.Consistency = nil
	if precise {
		consistency = &v1.Consistency{
			Requirement: &v1.Consistency_AtLeastAsFresh{
				AtLeastAsFresh: &v1.ZedToken{Token: zedToken},
			},
		}
	} else {
		consistency = &v1.Consistency{
			Requirement: &v1.Consistency_MinimizeLatency{
				MinimizeLatency: true,
			},
		}
	}
	req := &v1.CheckPermissionRequest{
		Consistency: consistency,
		Resource:    newObject(resourceType, resourceID),
		Subject:     newSubject(subjectType, subjectID),
		Permission:  permission,
	}
	res, err := svc.authzed.CheckPermission(ctx, req)

	if err != nil {
		return false, fmt.Errorf("spicedb check failed: %w", err)
	}

	hasPermission := res.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION
	return hasPermission, nil
}