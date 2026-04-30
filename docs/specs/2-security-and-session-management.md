## 🛠 Tech Stack & Infrastructure
*   **Language:** Go v1.26.2
*   **Web Framework:** Gin v1.12
*   **Database:** PostgreSQL (untuk kredensial user)
*   **Cache:** Redis v9 (Hashes untuk data sesi)
*   **Migration:** Goose

---

## 📂 Project Structure (Feature Slice: `auth`)
Kita akan menambahkan slice baru bernama `auth` yang menangani registrasi, login, dan manajemen sesi.

```text
internal/features/
├── services/           # (Sudah di Fase 1)
└── auth/               # New Feature Slice
    ├── delivery/       # Gin Handlers & Middlewares
    ├── repository/     # Postgres (Users) & Redis (Sessions)
    ├── usecase/        # Auth Business Logic (Bcrypt, Session Creation)
    └── dto/            # LoginRequest, SessionResponse
```

---

## 🗄 1. Database Schema (PostgreSQL)
**File:** `migrations/202604300002_create_users_table.sql`

```sql
-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role VARCHAR(20) DEFAULT 'user', -- e.g., 'user', 'mitra', 'admin'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE users;
```

---

## 🧠 2. Redis Hash Strategy
Berbeda dengan Fase 1 (Strings), kita menggunakan **Hashes** (`HSET`) karena data sesi biasanya berbentuk objek dengan beberapa field yang mungkin perlu diupdate secara parsial.

| Komponen | Spesifikasi |
| :--- | :--- |
| **Key Pattern** | `session:{session_id}` |
| **Structure** | **Hashes** |
| **Fields** | `user_id`, `role`, `user_agent`, `ip_address`, `last_activity` |
| **TTL** | 24 Jam (Diperbarui setiap kali user aktif / *Sliding Expiration*) |

---

## 🚀 3. API Specification

### A. User Registration & Login
*   **Endpoint:** `POST /v1/auth/register` & `POST /v1/auth/login`
*   **Logic Login:**
    1.  Verifikasi kredensial di PostgreSQL.
    2.  Generate `session_id` (UUID v4).
    3.  Simpan ke Redis: `HSET session:<id> user_id <uid> role <role> ...`
    4.  Set TTL: `EXPIRE session:<id> 86400`
    5.  Kirim `session_id` ke client melalui **Secure HttpOnly Cookie**.

### B. Authenticated "Me" Profile
*   **Endpoint:** `GET /v1/auth/me`
*   **Logic:** Mengambil data dari Redis Hash `HGETALL session:<id>` untuk mendapatkan info user tanpa menyentuh database utama.

### C. Logout (Session Invalidation)
*   **Endpoint:** `POST /v1/auth/logout`
*   **Logic:** Langsung hapus key dari Redis: `DEL session:<id>`. Ini adalah cara paling efektif untuk *invalidation*.

---

## 🛡 4. Auth Middleware Logic
Middleware ini akan membungkus endpoint yang membutuhkan proteksi (termasuk CRUD Service di Fase 1).

1.  Ambil `session_id` dari Cookie/Header.
2.  Cek keberadaan key di Redis: `EXISTS session:<id>`.
3.  Jika ada:
    *   Ambil data role: `HGET session:<id> role`.
    *   Perpanjang umur sesi (Sliding Expiration): `EXPIRE session:<id> 86400`.
    *   Simpan data session ke Gin Context `c.Set("user", sessionData)`.
4.  Jika tidak ada: Return `401 Unauthorized`.

---

## 💻 5. Implementasi Repository Snippet (Golang)

```go
// internal/features/auth/repository/session_repository.go

func (r *sessionRepo) CreateSession(ctx context.Context, sessionID string, user entity.User) error {
    key := fmt.Sprintf("session:%s", sessionID)
    
    data := map[string]interface{}{
        "user_id":       user.ID.String(),
        "role":          user.Role,
        "last_activity": time.Now().Format(time.RFC3339),
    }

    // Menggunakan HSet untuk menyimpan objek session
    err := r.redis.HSet(ctx, key, data).Err()
    if err != nil {
        return err
    }

    return r.redis.Expire(ctx, key, 24*time.Hour).Err()
}

func (r *sessionRepo) GetSession(ctx context.Context, sessionID string) (map[string]string, error) {
    return r.redis.HGetAll(ctx, fmt.Sprintf("session:%s", sessionID)).Result()
}
```

---

## 📝 6. Definition of Done (DoD) Fase 2
1.  [ ] User bisa register (password tersimpan sebagai bcrypt hash di Postgres).
2.  [ ] Login menghasilkan Session ID yang tersimpan di Redis dalam format **Hash**.
3.  [ ] Middleware berhasil memvalidasi sesi dan menolak request jika sesi kadaluarsa/dihapus.
4.  [ ] Fitur logout berhasil menghapus data di Redis secara permanen.
5.  [ ] Implementasi menggunakan **Go v1.26.2** dan **Gin v1.12** dengan struktur folder yang konsisten.
