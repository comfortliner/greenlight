# Create certificates

## Good explanation
[Youtube Video](https://www.youtube.com/watch?v=5vNFkLrYrIM&list=PL6QrD7_cU23kaZ05MvixcoJ5vctRD1qgC&index=13) \
(c) the native web GmbH

## How To create certificates

### 1. Create a private key
```code
$ openssl genrsa -out privateKey.pem 4096
```
#### (optional) Extract the public key
```code
$ openssl rsa -in privateKey.pem -pubout -out publicKey.pem
```
### 2. Create Certificate Signing Request (CSR) with private Key (send to CA)
```code
$ openssl req -new -key privateKey.pem -out csr.pem
```
The CA will check the request, after that you will hopefully get the certificate.pem.
Then the file `csr.pem` can be deleted. \
As a result, you now have the two files `certificate.pem` and `privateKey.pem`, which can be used until the end of the expire time (usually 365 days).

#### (optional) Show details of the certificate
```code
$ openssl x509 -in certificate.pem -text -noout
```

### (optional) Create self signed certificates

You will not want to have the CSR processed by a CA for testing or during development. \
Additionally, this wouldn't work for the domain `localhost`.

In this case you create a certificate by yourself.
```code
$ openssl x509 -in csr.pem -out certificate.pem -req -signkey privateKey.pem -days 365
```
