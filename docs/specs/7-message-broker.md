Fase 7 berfokus pada implementasi **Message Broker** menggunakan pola **Publish/Subscribe (Pub/Sub)**. Pada fase ini, kita akan memisahkan tanggung jawab antara pengirim pesan (API Server) dan penerima pesan (Background Worker) untuk meningkatkan skalabilitas dan isolasi kegagalan sistem[cite: 1].

Berikut adalah spesifikasi teknis lengkap untuk **Fase 7: Message Broker (Pub/Sub) & Background Worker**:

---

## 🛠 Tech Stack Snapshot
*   **Language:** Go v1.26.2[cite: 1].
*   **Web Framework:** Gin v1.12[cite: 1].
*   **Database:** PostgreSQL (Driver: `pgx/v5`)[cite: 1].
*   **Cache/Broker:** Redis v9 (**Pub/Sub**)[cite: 1, 4].
*   **Architecture:** Feature Slice (Vertical Slice) dengan **Worker Pattern**[cite: 1].

---

## 📂 Project Structure (Separated Entry Points)
Kita akan memisahkan *entry point* aplikasi menjadi dua: satu untuk melayani request HTTP dan satu untuk memproses pesan asinkron[cite: 1].

```text
.
├── cmd/
│   ├── api/                # API Server (Publisher)
│   │   └── main.go
│   └── worker/             # Background Worker (Subscriber)
│       └── main.go         # Entry point baru untuk Phase 7
├── internal/
│   ├── features/
│   │   └── notifications/  # New Feature Slice
│   │       ├── delivery/   # HTTP Handlers
│   │       ├── repository/ # Redis Pub/Sub & Postgres Persistence
│   │       ├── usecase/    # Subscriber Loop & Business Logic
│   │       └── dto/        # Message Schema
│   └── platform/           # Shared Infrastructure
```

---

## 🗄 1. Database Schema (PostgreSQL)
Riwayat notifikasi wajib disimpan di database agar user dapat melihat pesan lama (Inbox) meskipun pesan Pub/Sub bersifat *fire-and-forget*[cite: 6].

**File:** `migrations/202605150005_create_notifications_table.sql`[cite: 1]

```sql
-- +goose Up
CREATE TABLE notification_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL, -- Merujuk pada User ID atau Partner ID[cite: 2, 3]
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE notification_history;
```

---

## 🧠 2. Redis Pub/Sub Strategy
Redis akan menangani distribusi pesan secara *real-time* ke seluruh subscriber yang aktif[cite: 4, 6].

| Komponen | Spesifikasi |
| :--- | :--- |
| **Channel Pattern** | `notifications:user:{id}` (Private) atau `notifications:broadcast`[cite: 2]. |
| **Payload Format** | JSON String[cite: 1]. |
| **Karakteristik** | Asinkron, *Decoupled*, dan *Fire-and-forget*[cite: 1, 4]. |

---

## 🚀 3. API Specification (The Publisher)
API Server bertugas memicu pengiriman pesan ke Redis setelah melakukan mutasi data di database[cite: 1, 3].

*   **Logic:**
    1.  Menerima *trigger* dari aksi sistem (misal: **Partner Onboarding** di Fase 3 atau **Leaderboard Update** di Fase 6)[cite: 3, 6].
    2.  Menyimpan data awal ke database jika diperlukan[cite: 1].
    3.  Menjalankan perintah `PUBLISH` ke channel Redis yang relevan menggunakan `go-redis/v9`[cite: 4].

---

## 🤖 4. Background Worker Specification (The Subscriber)
Aplikasi di `cmd/worker/main.go` akan berjalan secara independen untuk mendengarkan pesan[cite: 1].

*   **Subscriber Loop:** Menjalankan loop `for { select { ... } }` yang mendengarkan channel Redis secara *blocking*[cite: 1].
*   **Asynchronous Processing:** Setiap pesan yang diterima akan diproses di dalam *goroutine* baru untuk mencegah hambatan pada antrean pesan berikutnya[cite: 1].
*   **Persistence:** Setiap pesan yang valid akan langsung disimpan ke tabel `notification_history` di PostgreSQL[cite: 6].
*   **Graceful Shutdown:** Worker harus menangani sinyal `SIGTERM` untuk menutup koneksi Redis dan menyelesaikan proses pesan yang sedang berjalan sebelum benar-benar berhenti[cite: 1].

---

## 📝 5. Definition of Done (DoD) Fase 7
1.  [ ] Tabel `notification_history` berhasil dideploy melalui migrasi `goose`[cite: 1].
2.  [ ] Aplikasi `worker` dapat berjalan secara terpisah dari aplikasi `api`[cite: 1].
3.  [ ] Pesan yang di-`PUBLISH` dari API Server berhasil ditangkap oleh Worker dan tercatat di database[cite: 4, 6].
4.  [ ] Logika menggunakan standar **Go v1.26.2** dan pola **Feature Slice** yang konsisten[cite: 1].
5.  [ ] Implementasi **Partner Welcome Notification** (Fase 3) berfungsi sebagai use case validasi pertama[cite: 3].
