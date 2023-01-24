package tests

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"testing"
)

func connectToMongo(t *testing.T) (context.Context, *mongo.Client) {
	t.Helper()

	ctx := context.Background()

	// Connect to MongoDB
	connStr := "mongodb://localhost:37017/test"
	opts := options.Client().ApplyURI(connStr)

	client, err := mongo.NewClient(opts)
	require.NoError(t, err)

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Check the connection.
	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	return ctx, client
}

func TestSessionAfterDisconnect(t *testing.T) {
	ctx, client := connectToMongo(t)

	// Start new session.
	sessionOpts := options.Session()
	session, err := client.StartSession(sessionOpts)
	require.NoError(t, err)

	// Get document within the session.
	res := session.Client().Database("test").Collection("foo").FindOne(ctx, bson.D{})
	require.NoError(t, res.Err())

	var result bson.D
	err = res.Decode(&result)
	require.NoError(t, err)

	log.Println(result)

	// Disconnect from MongoDB
	err = client.Disconnect(ctx)
	require.NoError(t, err)

	ctx, client = connectToMongo(t)

	// Try to get document within the same session.
	res = session.Client().Database("test").Collection("foo").FindOne(ctx, bson.D{})

	require.Equal(t, errors.New("client is disconnected"), res.Err())
}

func TestCursorErrorAfterSessionClosed(t *testing.T) {
	ctx, client := connectToMongo(t)

	// Start new session.
	sessionOpts := options.Session()
	session, err := client.StartSession(sessionOpts)
	require.NoError(t, err)

	result := session.Client().Database("test").RunCommand(ctx, bson.D{{"find", "foo"}})
	require.NoError(t, result.Err())

	var resDoc bson.D

	err = result.Decode(&resDoc)
	require.NoError(t, err)

	cursor, ok := resDoc.Map()["cursor"].(bson.D)
	require.True(t, ok)

	cursorID, ok := cursor.Map()["id"].(int64)
	require.True(t, ok)

	// Close the session.
	session.EndSession(ctx)

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
}

func TestRunCursorAfterSessionClosed(t *testing.T) {
	ctx, client := connectToMongo(t)

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
