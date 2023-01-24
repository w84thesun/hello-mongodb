package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestRunCursorAfterSessionClosed(t *testing.T) {
	ctx, client := connectToMongo(t)

	for i := 0; i < 200; i++ {
		_, err := client.Database("test").Collection("foo").InsertOne(ctx, bson.D{{"v", int32(i)}})
		require.NoError(t, err)
	}

	defer func() {
		client.Database("test").Collection("foo").Drop(ctx)
	}()

	// Get cursor.
	result := client.Database("test").RunCommand(ctx, bson.D{{"find", "foo"}})
	require.NoError(t, result.Err())

	var resDoc bson.D

	err := result.Decode(&resDoc)
	require.NoError(t, err)

	cursor, ok := resDoc.Map()["cursor"].(bson.D)
	require.True(t, ok)

	cursorID, ok := cursor.Map()["id"].(int64)
	require.True(t, ok)

	// Start new session.
	sessionOpts := options.Session()
	session, err := client.StartSession(sessionOpts)
	require.NoError(t, err)

	// Run getMore command within the session.
	result = session.Client().
		Database("test").
		RunCommand(ctx, bson.D{
			{"getMore", cursorID},
			{"collection", "foo"},
		})

	cmdErr, ok := result.Err().(mongo.CommandError)
	require.True(t, ok)

	require.Equal(t, int32(50738), cmdErr.Code)
	require.Regexp(t, "Cannot run getMore on cursor \\d+, which was created in session"+
		" .{36} - .{44} -  - , in session .{36} - .{44} -  - ", cmdErr.Message)

	subtype, value, ok := session.ID().Lookup("id").BinaryOK()
	require.True(t, ok)
	require.Equal(t, byte(4), subtype)

	id, err := uuid.FromBytes(value)
	require.NoError(t, err)

	require.Regexp(t, ", in session "+id.String(), cmdErr.Message)

	// Close the session.
	session.EndSession(ctx)
}
