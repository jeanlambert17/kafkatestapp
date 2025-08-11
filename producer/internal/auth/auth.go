package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetOrgID reads the x-org header and returns both the ObjectID and its hex string
func GetOrgID(ctx *gin.Context) (primitive.ObjectID, string, error) {
	org := ctx.GetHeader("x-org")
	if org == "" {
		return primitive.NilObjectID, "", errors.New("missing x-org header")
	}
	id, err := primitive.ObjectIDFromHex(org)
	if err != nil {
		return primitive.NilObjectID, "", errors.New("invalid x-org header format")
	}
	return id, org, nil
}
