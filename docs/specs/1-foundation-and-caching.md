## 🛠 Tech Stack Snapshot

- **Language:** Go v1.26.2
- **Web Framework:** Gin v1.12
- **Database:** PostgreSQL (Driver: pgx/v5)
- **Migration:** Goose
- **Cache:** Redis (go-redis/v9)
- **Architecture:** Feature Slice (Vertical Slice)

---

## 📂 Project Structure (Feature Slice)

Kita akan mengelompokkan kode berdasarkan fitur `services`, bukan berdasarkan layer teknis.

```text
.
├── cmd/api/main.go
├── migrations/             # Goose migrations
├── internal/
│   ├── features/
│   │   └── services/       # Feature Slice: Services
│   │       ├── delivery/   # Gin Handlers
│   │       ├── repository/ # Postgres & Redis logic
│   │       ├── usecase/    # Business Logic
│   │       └── dto/        # Request & Response
│   └── platform/           # Shared (Database, Redis, Config)
└── go.mod
```

---

## 🗄 1. Database Schema & Migrations

**File:** `migrations/202604250001_create_services_table.sql`

```sql
-- +goose Up
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE services;
```

---

## 🧠 2. Redis Caching Strategy

Kita menggunakan pola **Cache-Aside (Lazy Loading)**.

| Action         | Redis Key            | Expiry (TTL) | Invalidation Logic              |
| :------------- | :------------------- | :----------- | :------------------------------ |
| **Get All**    | `services:all`       | 1 Hour       | `DEL` saat Create/Update/Delete |
| **Get Detail** | `services:id:{uuid}` | 1 Hour       | `DEL` saat Update/Delete        |

---

## 🚀 3. API Specification

### A. Create Service

- **Endpoint:** `POST /v1/services`
- **Logic:** Simpan ke DB -> **Hapus** cache `services:all`.

### B. Get All Services (Cached)

- **Endpoint:** `GET /v1/services`
- **Logic:** Cek cache `services:all` -> Jika nilainya ada, return. Jika tidak, query DB -> Set cache -> Return.

### C. Update Service

- **Endpoint:** `PUT /v1/services/:id`
- **Logic:** Update DB -> **Hapus** cache `services:all` & `services:id:{id}`.

### D. Delete Service

- **Endpoint:** `DELETE /v1/services/:id`
- **Logic:** Delete DB -> **Hapus** cache `services:all` & `services:id:{id}`.

---

## 🔄 4. Sequence Flow (Data Interaction)

---

## 💻 5. Implementasi Repository (Golang Slice)

Contoh implementasi `repository` yang menggabungkan Postgres dan Redis:

```go
// internal/features/services/repository/repository.go

func (r *serviceRepo) GetAll(ctx context.Context) ([]entity.Service, error) {
    key := "services:all"

    // 1. Try get from Redis
    val, err := r.redis.Get(ctx, key).Result()
    if err == nil {
        var services []entity.Service
        json.Unmarshal([]byte(val), &services)
        return services, nil
    }

    // 2. Fallback to DB
    services, err := r.db.QueryAll(ctx)
    if err != nil {
        return nil, err
    }

    // 3. Set to Redis (Async or Sync)
    data, _ := json.Marshal(services)
    r.redis.Set(ctx, key, data, 1*time.Hour)

    return services, nil
}

func (r *serviceRepo) Invalidate(ctx context.Context, id string) {
    // Menghapus cache list dan detail secara atomik
    r.redis.Del(ctx, "services:all", fmt.Sprintf("services:id:%s", id))
}
```

---

## 📝 6. Definition of Done (DoD) Fase 1

1.  [ ] Migrasi database berhasil dijalankan via `goose up`.
2.  [ ] Endpoint CRUD berfungsi di Postman/Insomnia.
3.  [ ] Log Redis menunjukkan `SET` saat akses pertama dan `DEL` saat ada mutasi data (Create/Update/Delete).
4.  [ ] Struktur folder mengikuti pola Feature Slice tanpa _circular dependency_.
