package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
