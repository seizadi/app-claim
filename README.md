# app-claim
Research application claim pattern.

See [app-claim docs](https://seizadi.github.io/app-claim) for more
detail on the application claim pattern.

## CLI Help

For Discovery
```bash
claims search --dir <path to manifest repo> amazonaws.com bucket
claims search --graphdb <user>:<password>@neo4j://<host>:<port> --stage prod --env prd-env --app identity amazonaws.com
claims search --graphdb <user>:<password>@neo4j://<host>:<port> --aws
claims search --graphdb <user>:<password>@neo4j://<host>:<port> --claims
```

For reports on resource usage by applications
```bash
claims report --graphdb <user>:<password>@neo4j://<host>:<port> --stage prod --env prd-env --appsfile <path to application list>.csv
claims report --graphdb <user>:<password>@neo4j://<host>:<port> --stage prod --env prd-env cert-manager identity
claims report --graphdb <user>:<password>@neo4j://<host>:<port> --stage prod --env prd-env --app identity
```


