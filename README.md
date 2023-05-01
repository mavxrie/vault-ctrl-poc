# vault-ctrl-poc

## Launch vault

It is needed to have a vault running. You can use the `contrib/vault.hcl` provided to set up your vault:


```fish
$ vault -config=data/vault.hcl
```

Then, initialize the Vault instance and unseal it:

```fish
$ set -x VAULT_ADDR http://localhost:8200/
$ vault operator init
...
Unseal Key 1: mHSgmFzpa4aOHt6WfNu3w2G7YYP1Y71j0TnEKUUDTIOP
Unseal Key 2: lHTkMYsLsSkgLNZihy6gPL+LUAN3wdnLY8pJo7Hk+AtD
Unseal Key 3: 5rN6h1++uKVDllhVMGSXt5uoJfUeUOlGxUlJok27du9N
Unseal Key 4: RoGfdfsdjzjaLKACcTiZ9/OaHWN+Bzn2pMFNi9R9GVih
Unseal Key 5: GSubnXnSa5an8eBfWXbr5cYayS+86pw0C8ogip4T/5Ev

Initial Root Token: hvs.MFt7xkPro6TchQDF2BIVvkfQ

$ vault operator unseal mHSgmFzpa4aOHt6WfNu3w2G7YYP1Y71j0TnEKUUDTIOP
$ vault operator unseal lHTkMYsLsSkgLNZihy6gPL+LUAN3wdnLY8pJo7Hk+AtD
$ vault operator unseal GSubnXnSa5an8eBfWXbr5cYayS+86pw0C8ogip4T/5Ev
```

Finally, set-up a `kv-v2` storage engine:

```fish
$ set -x VAULT_TOKEN hvs.MFt7xkPro6TchQDF2BIVvkfQ
$ vault secrets enable -path=secret/ kv-v2
```

## Launch vault-ctrl-poc

```fish
$ set -x VAULT_TOKEN hvs.MFt7xkPro6TchQDF2BIVvkfQ
$ set -x VAULT_ADDR http://localhost:8200/
$ go build
$ ./vault-ctrl-poc
```

## API endpoints

This part must be written

## Algos


## AppRole

This part is currently under testing.

AppRole were initialized:

```fish
> vault auth enable -path=approle-controllers/ approle
Success! Enabled approle auth method at: approle-controllers/

> vault secrets enable -path=secret-controllers/ kv-v2
Success! Enabled the kv-v2 secrets engine at: secret-controllers/
```

Create a policy for allowing controllers to read into secret-controllers/...

Create a controller:

```fish
> vault write auth/approle-controllers/role/alpha-centauri bind_secret_id=false token_ttl=24h token_max_ttl=768h token_bound_cidrs="0.0.0.0/0" token_policies="ctrl-policy"
Success! Data written to: auth/approle-controllers/role/alpha-centauri

> vault read auth/approle-controllers/role/alpha-centauri/role-id
Key        Value
---        -----
role_id    2aadc38e-59e8-f1ba-2302-565fe4480cb9

> vault kv put -mount=secret controllers/alpha-centauri role_id=2aadc38e-59e8-f1ba-2302-565fe4480cb9
============= Secret Path =============
secret/data/controllers/alpha-centauri

> vault kv put -mount=secret-controllers controllers/alpha-centauri/identity name=alpha-centauri
=================== Secret Path ===================
secret-controllers/data/controllers/alpha-centauri
```

We have a role-id for the controller. We stored it in Vault in controllers/alpha-centauri for future internal usage.

Login as the controller to get a token:

```fish
> vault write auth/approle-controllers/login role_id=2aadc38e-59e8-f1ba-2302-565fe4480cb9
Key                     Value
---                     -----
token                   hvs.CAESIPli4kfdUWbbEiTwhiDIGZCME_VYYN52cVlMNuKRpPWDGh4KHGh2cy5zakpNNlAyQ0Vsdmd3ZEp0aEZmZWllR3g
token_accessor          mRNUUm1Lmb7pePr67gw63ILb
token_duration          24h
token_renewable         true
token_policies          ["ctrl-policy" "default"]
identity_policies       []
policies                ["ctrl-policy" "default"]
token_meta_role_name    alpha-centauri
```


Get secrets using token:

```fish
> VAULT_TOKEN=hvs.CAESIPli4kfdUWbbEiTwhiDIGZCME_VYYN52cVlMNuKRpPWDGh4KHGh2cy5zakpNNlAyQ0Vsdmd3ZEp0aEZmZWllR3g vault kv get -mount=secret-controllers controllers/alpha-centauri/identity
======================= Secret Path =======================
secret-controllers/data/controllers/alpha-centauri/identity

======= Metadata =======
Key                Value
---                -----
created_time       2023-04-29T18:51:34.737661746Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1

==== Data ====
Key     Value
---     -----
name    alpha-centauri
```

