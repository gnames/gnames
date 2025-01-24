# Migrations with `atlas`

Install [atlas](https://atlasgo.io)

Make `tmp` database and set it in configuration file.
Create run migration on `tmp` using current state of Gorm models.

```bash
gnidump create
```

Get HCL file with current database schema

```bash
atlas schema inspect \
  -u 'postgres://user:pass@0.0.0.0/tmp?sslmode=disable' > gnames.hcl
```

Delete materialized view (it prevents migration to happen)

```sql
drop materialized view verification;
```

Syncronize old state to new one:

```bash
atlas schema apply  \
  -u 'postgres://user:pass@0.0.0.0/gnames?sslmode=disable' \
  --to file://gnames.hcl
```

Save this new state with git, now git keeps history of states.
