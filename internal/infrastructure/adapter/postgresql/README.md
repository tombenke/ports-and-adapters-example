# PostgreSQL Order Repository

Implements `ports.OrderRepository` and `ports.OrderLister`. The caller owns the
`*sql.DB` connection pool and is responsible for closing it. The existing servers
continue to use the memory adapter.

```go
import (
    "context"
    "database/sql"

    _ "github.com/lib/pq"
    "github.com/tombenke/ports-and-adapters-example/internal/infrastructure/adapter/postgresql"
)

func newRepository(ctx context.Context, dsn string) (*sql.DB, *postgresql.OrderRepository, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, nil, err
    }
    repo := postgresql.NewOrderRepository(db)
    if err := repo.EnsureSchema(ctx); err != nil {
        _ = db.Close()
        return nil, nil, err
    }
    return db, repo, nil
}
```

`EnsureSchema` is an explicit, idempotent setup operation. Alternatively, apply
`schema.sql` using your migration tooling. It does not migrate an existing table
with an incompatible schema. A runtime account does not need DDL permissions if
the schema has already been provisioned.

Each order is stored as one row. Items are a JSONB array containing product ID,
quantity, price in cents, and currency. `Save` atomically replaces every field of
an existing order using an upsert; no partial item updates are possible. Like the
memory adapter, concurrent saves use last-writer-wins semantics. PostgreSQL stores
timestamps with microsecond precision, rather than Go's nanosecond precision.

`FindByID` returns `domain.ErrOrderNotFound` for missing orders. `ListByCustomer`
returns orders sorted by creation time, then ID. All database calls use the
supplied context, and wrapped errors retain their identity through `errors.Is`.

## Integration Test

Use a test database account with permission to create schemas. The test creates
and drops its own isolated schema, without modifying existing application tables.
Without `POSTGRES_TEST_DSN`, the database integration test is skipped.

```bash
POSTGRES_TEST_DSN='postgres://postgres:postgres@localhost:5432/orders_test?sslmode=disable' \
  go test -count=1 -v ./internal/infrastructure/adapter/postgresql
```
