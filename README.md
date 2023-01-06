# hello-mongodb

Simple "Hello world" application to try various MongoDB connection options.

For educational purposes only.

### Prepare TLS certificates for testing purposes

```bash
mkdir -p certs
cd certs

### CA
openssl genrsa -out rootCA.key 2048
openssl req -x509 -new -nodes -key rootCA.key -days 1024 -out rootCA.pem -subj "/CN=mongodb-root-ca"

### SERVER 
openssl genrsa -out mongodb.key 2048
openssl req -new -key mongodb.key -out mongodb.csr -subj "/CN=localhost"
openssl x509 -req -in mongodb.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out mongodb.crt -days 365
cat mongodb.crt mongodb.key > mongodb.pem

### CLIENT
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/CN=localhost"
openssl x509 -req -in client.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out client.crt -days 365
cat client.crt client.key > client.pem

```

### Connect

```bash
mongosh -tlsCertificateKeyFile=./certs/client.pem --tlsCAFile=./certs/rootCA.pem mongodb://localhost:7777/test?tls=true
```