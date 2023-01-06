package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	// Load the root CA certificate
	rootCA, err := os.ReadFile("./certs/rootCA.pem")
	if err != nil {
		log.Fatal(err)
	}
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(rootCA)
	if !ok {
		log.Fatal("failed to parse root certificate")
	}

	// Load the client certificate and key
	cert, err := tls.LoadX509KeyPair("./certs/client.crt", "./certs/client.key")
	if err != nil {
		log.Fatal(err)
	}

	// Create the TLS config
	tlsConfig := &tls.Config{
		RootCAs:            roots,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	fmt.Println("Opened cert files.")

	// Connect to MongoDB
	connstr := "mongodb://root:pass@127.0.0.1:7777"
	client, err := mongo.NewClient(options.Client().ApplyURI(connstr).SetTLSConfig(tlsConfig))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected.")

	// Insert
	collection := client.Database("foo").Collection("bar")
	_, err = collection.InsertOne(ctx, map[string]interface{}{"hello": "world"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted.")

	// Find
	var result *bson.D
	err = collection.FindOne(ctx, bson.D{}).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found: %+v\n.", result)

	// Disconnect from MongoDB
	err = client.Disconnect(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
