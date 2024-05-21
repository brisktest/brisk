rm *.pem
set -e  
# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=US/ST=CA/L=LA/O=Brisk/CN=*.brisktest.com"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout logservice-key.pem -out logservice-req.pem -subj "/C=US/ST=CA/L=LA/O=Brisk/CN=*.brisktest.com"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in logservice-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out logservice-cert.pem -extfile server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in logservice-cert.pem -noout -text

mv *.pem ../../certs/logservice