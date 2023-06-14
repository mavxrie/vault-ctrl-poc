# Vault

* https://aviatrix.atlassian.net/wiki/spaces/AVXSRE/pages/1867579719/Vault+usage+in+Licensing+for+controllers


## Download, installation & start-up

```sh
> sudo dnf install -y vault
Last metadata expiration check: 1:09:21 ago on Tue 13 Jun 2023 03:41:11 PM CEST.
Package vault-1.13.3-1.x86_64 is already installed.
Dependencies resolved.
Nothing to do.
Complete!
```

Then:

```sh
> mkdir -p /tmp/vault-test
> cd /tmp/vault-test

> vault server -dev
==> Vault server configuration:

             Api Address: http://127.0.0.1:8200
                     Cgo: disabled
...
WARNING! dev mode is enabled! In this mode, Vault runs entirely in-memory
and starts unsealed with a single unseal key. The root token is already
authenticated to the CLI, so you can immediately begin using Vault.

You may need to set the following environment variables:

    $ export VAULT_ADDR='http://127.0.0.1:8200'

The unseal key and root token are displayed below in case you want to
seal/unseal the Vault or re-authenticate.

Unseal Key: tx0BG3prV154HXZG1ADgFvL9nT5xkUMO/Fmr8GprMxY=
Root Token: hvs.y4k1e0eCAWHtI1U0CwsYr6jX

Development mode should NOT be used in production installations!
```

### UI

* Log into http://127.0.0.1:8200 using the `root token`.

### CLI Usage

```sh
> set -x VAULT_ADDR http://127.0.0.1:8200
> vault token lookup
Key                 Value
---                 -----
...
id                  hvs.y4k1e0eCAWHtI1U0CwsYr6jX
```

#### Create token with default policy

```sh
> vault token create -policy=default
```

#### Create policy to allow stuff

```yaml
path "secret/*" {
    capabilities = ["read", "create", "update", "patch", "delete"]
}
```

#### App roles

##### Init

```sh
> vault auth enable -path="approle-controllers" approle
> vault auth list
Path                    Type       Accessor                 Description                Version
----                    ----       --------                 -----------                -------
approle-controllers/    approle    auth_approle_faf723cc    n/a                        n/a
token/                  token      auth_token_58af7eba      token based credentials    n/a

```

##### A place to store secrets per controller

```sh
> vault secrets enable -path="secret-controllers" kv-v2
```

```yaml
path "secret-controllers/data/controllers/{{identity.entity.aliases.auth_approle_faf723cc.metadata.role_name}}/*" {
    capabilities = [ "create", "update", "read", "delete", "list" ]
}
```

##### Create an app role for a single controller

```
> vault write auth/approle-controllers/role/uss-enterprise bind_secret_id=false token_ttl=24h token_max_ttl=768h token_bound_cidrs="0.0.0.0/0" token_policies="ctrl-policy"
Success! Data written to: auth/approle-controllers/role/uss-enterprise

> vault read auth/approle-controllers/role/uss-enterprise/role-id
Key        Value
---        -----
role_id    184be9b8-46ba-ac92-d1a1-c82095116540
```

##### Login using the app role

```sh
> vault write auth/approle-controllers/login role_id=184be9b8-46ba-ac92-d1a1-c82095116540
Key                     Value
---                     -----
token                   hvs.CAESICa3MOmpJJ4Cth67frl7y9SjGcMQaXIjXeDZ_LnQdI4gGh4KHGh2cy5ZbG1ub2hvR0s3aGdrakRSRjNYclBEMXg
token_accessor          LgjRobVZ2xdURtduUE1lEU2c
token_duration          24h
token_renewable         true
token_policies          ["ctrl-policy" "default"]
```

##### Use token to access the secrets storage

```sh
> set -x VAULT_TOKEN hvs.CAESICa3MOmpJJ4Cth67frl7y9SjGcMQaXIjXeDZ_LnQdI4gGh4KHGh2cy5ZbG1ub2hvR0s3aGdrakRSRjNYclBEMXg
> vault kv put -mount=secret-controllers controllers/uss-enterprise/random foo=bar
====================== Secret Path ======================
secret-controllers/data/controllers/uss-enterprise/random

> vault kv get -mount=secret-controllers controllers/uss-enterprise/random
====================== Secret Path ======================
secret-controllers/data/controllers/uss-enterprise/random

> vault kv get -mount=secret-controllers ontrollers/uss-excelsior/random
Error reading secret-controllers/data/ontrollers/uss-excelsior/random: Error making API request.
```


##### Wrap the token (using the k8s app token/license server/whatever)

```sh
> vault write sys/wrapping/wrap token=hvs.CAESICa3MOmpJJ4Cth67frl7y9SjGcMQaXIjXeDZ_LnQdI4gGh4KHGh2cy5ZbG1ub2hvR0s3aGdrakRSRjNYclBEMXg
Key                              Value
---                              -----
wrapping_token:                  hvs.CAESIPN0R9f6ykeg8btMs3CIGXsPMvqs8oRPlXsnyMETmd5dGh4KHGh2cy5ESEV0UHhFTWxQWkNNM2tpZmZyQ0loS1U
wrapping_accessor:               0IKFsdPKITg1D5H2TvzhfQ9N
wrapping_token_ttl:              5m
wrapping_token_creation_time:    2023-06-13 17:27:47.071361623 +0200 CEST
wrapping_token_creation_path:    sys/wrapping/wrap
```

##### Unwrap the token on client side

```sh
> VAULT_TOKEN=hvs.CAESIDqDNm0AQTIni5atWpnEAlsmeMml5tLqsDP5jIom2wHXGh4KHGh2cy54UzZXN3U0Q2hTcUoyWVdKekhwcjJ0c2Y vault unwrap
Key      Value
---      -----
token    hvs.CAESICa3MOmpJJ4Cth67frl7y9SjGcMQaXIjXeDZ_LnQdI4gGh4KHGh2cy5ZbG1ub2hvR0s3aGdrakRSRjNYclBEMXg

> VAULT_TOKEN=hvs.CAESIDqDNm0AQTIni5atWpnEAlsmeMml5tLqsDP5jIom2wHXGh4KHGh2cy54UzZXN3U0Q2hTcUoyWVdKekhwcjJ0c2Y vault unwrap
Error unwrapping: Error making API request.

URL: PUT http://127.0.0.1:8200/v1/sys/wrapping/unwrap
Code: 400. Errors:

* wrapping token is not valid or does not exist
```

##### Using cubbyhole

###### Create cubbyhole temporary token

```sh
> vault token create -policy=default
Key                  Value
---                  -----
token                hvs.CAESIFzE_tDtRZRCbQRsvU-ZJRVTk2pNnbLNKEiPmv5DKDCxGh4KHGh2cy5UM1plNUFPUE90YWZLQ0FpaXpvM2hENEU
```

###### Wrap & store controller token into cubbyhole

```sh
> vault write sys/wrapping/wrap token=hvs.CAESICa3MOmpJJ4Cth67frl7y9SjGcMQaXIjXeDZ_LnQdI4gGh4KHGh2cy5ZbG1ub2hvR0s3aGdrakRSRjNYclBEMXg
Key                              Value
---                              -----
wrapping_token:                  hvs.CAESILQGmUDqot21ECxAIyreKr0wIvH_GMQaQupeLfOdwYlNGh4KHGh2cy4wbHlkQzZGRjFkNllpN3lPaHVpb1JDclg
wrapping_accessor:               5UnJ0UlHQwg3XwwhyBSBIy3y
wrapping_token_ttl:              5m
wrapping_token_creation_time:    2023-06-13 17:31:48.494576153 +0200 CEST
wrapping_token_creation_path:    sys/wrapping/wrap

> VAULT_TOKEN=hvs.CAESIFzE_tDtRZRCbQRsvU-ZJRVTk2pNnbLNKEiPmv5DKDCxGh4KHGh2cy5UM1plNUFPUE90YWZLQ0FpaXpvM2hENEU vault kv put -mount=cubbyhole token token=hvs.CAESILQGmUDqot21ECxAIyreKr0wIvH_GMQaQupeLfOdwYlNGh4KHGh2cy4wbHlkQzZGRjFkNllpN3lPaHVpb1JDclg
Success! Data written to: cubbyhole/token
```

###### Pass the cubbyhole token to controller, then read & unwrap

```sh
> VAULT_TOKEN=hvs.CAESIFzE_tDtRZRCbQRsvU-ZJRVTk2pNnbLNKEiPmv5DKDCxGh4KHGh2cy5UM1plNUFPUE90YWZLQ0FpaXpvM2hENEU vault token lookup
policies            [default]

> VAULT_TOKEN=hvs.CAESIFzE_tDtRZRCbQRsvU-ZJRVTk2pNnbLNKEiPmv5DKDCxGh4KHGh2cy5UM1plNUFPUE90YWZLQ0FpaXpvM2hENEU vault kv get -mount=cubbyhole token
==== Data ====
Key      Value
---      -----
token    hvs.CAESILQGmUDqot21ECxAIyreKr0wIvH_GMQaQupeLfOdwYlNGh4KHGh2cy4wbHlkQzZGRjFkNllpN3lPaHVpb1JDclg

> VAULT_TOKEN=hvs.CAESILQGmUDqot21ECxAIyreKr0wIvH_GMQaQupeLfOdwYlNGh4KHGh2cy4wbHlkQzZGRjFkNllpN3lPaHVpb1JDclg vault unwrap
Key      Value
---      -----
token    hvs.CAESICa3MOmpJJ4Cth67frl7y9SjGcMQaXIjXeDZ_LnQdI4gGh4KHGh2cy5ZbG1ub2hvR0s3aGdrakRSRjNYclBEMXg

> VAULT_TOKEN=hvs.CAESILQGmUDqot21ECxAIyreKr0wIvH_GMQaQupeLfOdwYlNGh4KHGh2cy4wbHlkQzZGRjFkNllpN3lPaHVpb1JDclg vault unwrap
Error unwrapping: Error making API request.

> VAULT_TOKEN=hvs.CAESICa3MOmpJJ4Cth67frl7y9SjGcMQaXIjXeDZ_LnQdI4gGh4KHGh2cy5ZbG1ub2hvR0s3aGdrakRSRjNYclBEMXg vault token lookup
Key                 Value
---                 -----
accessor            LgjRobVZ2xdURtduUE1lEU2c
bound_cidrs         [0.0.0.0/0]
creation_time       1686669541
creation_ttl        24h
display_name        approle-controllers
entity_id           5e62e5a6-2c99-ab53-7eef-397836a2b938
expire_time         2023-06-14T17:19:01.270041674+02:00
explicit_max_ttl    0s
id                  hvs.CAESICa3MOmpJJ4Cth67frl7y9SjGcMQaXIjXeDZ_LnQdI4gGh4KHGh2cy5ZbG1ub2hvR0s3aGdrakRSRjNYclBEMXg
issue_time          2023-06-13T17:19:01.270044136+02:00
meta                map[role_name:uss-enterprise]
```

Read a secret:

```sh
> VAULT_TOKEN=hvs.CAESICa3MOmpJJ4Cth67frl7y9SjGcMQaXIjXeDZ_LnQdI4gGh4KHGh2cy5ZbG1ub2hvR0s3aGdrakRSRjNYclBEMXg vault kv get -mount=secret-controllers controllers/uss-enterprise/random
====================== Secret Path ======================
secret-controllers/data/controllers/uss-enterprise/random

======= Metadata =======
Key                Value
---                -----
created_time       2023-06-13T15:21:15.660394035Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1

=== Data ===
Key    Value
---    -----
foo    bar
```





