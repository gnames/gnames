# Migrations with `atlas`

Install [atlas](https://atlasgo.io)

Create empty gnames database and run current state of Gorm models.

Get HCL file with current database schema

```bash
atlas schema inspect \
  -u 'postgres://user:pass@0.0.0.0/gnames?sslmode=disable' > gnames.hcl
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
