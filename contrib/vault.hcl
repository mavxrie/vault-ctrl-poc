ui            = true
disable_mlock = true

storage "file" {
  path = "contrib/vault/data"
}

listener "tcp" {
  address       = "127.0.0.1:8200"
  tls_disable = "true"
}
