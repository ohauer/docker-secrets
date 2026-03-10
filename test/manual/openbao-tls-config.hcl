ui = true
disable_mlock = true

storage "inmem" {}

listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/certs/openbao-cert.pem"
  tls_key_file  = "/certs/openbao-key.pem"
}

api_addr = "https://0.0.0.0:8200"
