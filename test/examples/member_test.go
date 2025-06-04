package examples

import (
	"app/store/db"
	testhelpers "app/test/helpers"
	"testing"
	"time"

	"github.com/nhnghia272/gopkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMemberCRUD(t *testing.T) {
	// Setup test environment
	tc := testhelpers.SetupTest(t, "member")
	defer testhelpers.CleanupTest(t)

	t.Run("Create Member", func(t *testing.T) {
		// Create test member
		memberID, member := tc.CreateTestMember()
		assert.NotEmpty(t, memberID)

		// Verify member was created
		var retrievedMember db.MemberDomain
		tc.GetDocumentByID(memberID, &retrievedMember)

		assert.Equal(t, *member.FirstName, *retrievedMember.FirstName)
		assert.Equal(t, *member.Email, *retrievedMember.Email)
		assert.Equal(t, *member.MemberStatus, *retrievedMember.MemberStatus)
	})

	t.Run("Update Member", func(t *testing.T) {
		// Create test member
		memberID, _ := tc.CreateTestMember()
		objID, err := primitive.ObjectIDFromHex(memberID)
		require.NoError(t, err)

		// Update member
		newName := "Updated Test Member"
		update := bson.M{"$set": bson.M{"first_name": newName}}
		_, err = tc.Database.Collection("member").UpdateOne(
			tc.Context,
			bson.M{"_id": objID},
			update,
		)
		require.NoError(t, err)

		// Verify update
		var updatedMember db.MemberDomain
		tc.GetDocumentByID(memberID, &updatedMember)
		assert.Equal(t, newName, *updatedMember.FirstName)
	})

	t.Run("Query Members", func(t *testing.T) {
		// Create multiple members
		tc.CreateTestMember()
		tc.CreateTestMember()

		// Create a member with different status
		id := primitive.NewObjectID()
		now := time.Now()
		inactiveMember := &db.MemberDomain{
			BaseDomain: db.BaseDomain{
				ID:        id,
				CreatedAt: gopkg.Pointer(now),
				UpdatedAt: gopkg.Pointer(now),
				CreatedBy: gopkg.Pointer("test_user"),
				UpdatedBy: gopkg.Pointer("test_user"),
			},
			FirstName:    gopkg.Pointer("Inactive"),
			LastName:     gopkg.Pointer("Member"),
			Email:        gopkg.Pointer("inactive@example.com"),
			MemberStatus: gopkg.Pointer("inactive"),
			TenantId:     gopkg.Pointer(tc.TenantID),
		}
		tc.Database.Collection("member").InsertOne(tc.Context, inactiveMember)

		// Query active members
		filter := bson.M{"member_status": "active"}
		tc.AssertCollectionCount(filter, 2)

		// Query inactive members
		filter = bson.M{"member_status": "inactive"}
		tc.AssertCollectionCount(filter, 1)
	})

	t.Run("Delete Member", func(t *testing.T) {
		// Create test member
		memberID, _ := tc.CreateTestMember()
		objID, err := primitive.ObjectIDFromHex(memberID)
		require.NoError(t, err)

		// Delete member
		_, err = tc.Database.Collection("member").DeleteOne(
			tc.Context,
			bson.M{"_id": objID},
		)
		require.NoError(t, err)

		// Verify delete
		count, err := tc.Database.Collection("member").CountDocuments(
			tc.Context,
			bson.M{"_id": objID},
		)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
