# Grants secret access to controllers
path "secret-controllers/data/controllers/{{identity.entity.aliases.auth_approle_07463cae.metadata.role_name}}/*" {
    capabilities = [ "create", "update", "read", "delete", "list" ]
}

