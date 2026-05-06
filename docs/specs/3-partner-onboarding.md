## 🛠 Tech Stack Snapshot
*   **Language:** Go v1.26.2
*   **Web Framework:** Gin v1.12
*   **Database:** PostgreSQL (Partners Table)
*   **Migration:** Goose (Previously applied)

---

## 📂 Project Structure (Feature Slice: `services`)
Kita akan menggunakan fitur `services` yang sudah ada untuk mengelola pendaftaran partner, dengan mengasosiasikan `user_id` ke `service_id`.

```text
internal/features/services/
├── delivery/       # POST /services/:serviceId/partners
├── repository/     # INSERT INTO partners (with service_id)
├── usecase/        # Onboarding logic
└── dto/            # JoinPartnerRequest, PartnerResponse
```

---

## 🗄 1. Database Schema (PostgreSQL)
Tabel ini menghubungkan user dengan jenis layanan yang mereka tawarkan.

**File:** `migrations/202605050003_create_partners_table.sql`

```sql
-- +goose Up
CREATE TABLE partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) UNIQUE,
    service_id UUID NOT NULL REFERENCES services(id),
    is_active BOOLEAN DEFAULT true,
    rating DECIMAL(2, 1) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE partners;
```

---

## 🚀 2. API Specification

### A. Join as Partner for a Service
Memungkinkan user yang sudah terautentikasi untuk mendaftarkan diri sebagai penyedia layanan tertentu.

*   **Endpoint:** `POST /v1/services/:serviceId/partners`
*   **Auth:** Required (Session)
*   **Logic:**
    1.  Ekstrak `user_id` dari session context.
    2.  Ambil `serviceId` dari path parameter.
    3.  Validasi apakah `service_id` tersebut ada di tabel `services`.
    4.  Cek apakah `user_id` sudah terdaftar di tabel `partners` (User hanya bisa punya satu profil partner).
    5.  Insert record baru ke tabel `partners` dengan `service_id` tersebut.
    6.  Kembalikan data partner lengkap.

---

## 💻 3. Proposed Implementation Details

### DTO (Data Transfer Object)
```go
// internal/features/services/dto/partner_dto.go
type PartnerResponse struct {
    ID          string  `json:"id"`
    UserID      string  `json:"user_id"`
    ServiceID   string  `json:"service_id"`
    IsActive    bool    `json:"is_active"`
    Rating      float64 `json:"rating"`
    CreatedAt   string  `json:"created_at"`
}
```

---

## 📝 4. Definition of Done (DoD) Fase 3
1.  [ ] Tabel `partners` berhasil didefinisikan dengan `service_id`.
2.  [ ] Endpoint `POST /v1/services/:serviceId/partners` berhasil diimplementasikan.
3.  [ ] Validasi Foreign Key ke tabel `services` berfungsi (404 jika service tidak ada).
4.  [ ] User yang baru mendaftar mendapatkan `partner_id` yang valid untuk digunakan di fitur lain.
