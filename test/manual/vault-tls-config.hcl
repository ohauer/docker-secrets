ui = true
disable_mlock = true

storage "inmem" {}

listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/certs/vault-cert.pem"
  tls_key_file  = "/certs/vault-key.pem"
}

api_addr = "https://0.0.0.0:8200"
