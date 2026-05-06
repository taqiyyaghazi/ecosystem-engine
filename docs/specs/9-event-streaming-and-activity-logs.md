### 🛠 Tech Stack Snapshot
*   **Language:** Go v1.26.2[cite: 1].
*   **Web Framework:** Gin v1.12[cite: 1].
*   **Database:** PostgreSQL (Archival & Auditing)[cite: 1].
*   **Cache/Streaming:** Redis v9 (**Streams** data structure)[cite: 1, 4].
*   **Architecture:** Feature Slice (Vertical Slice) dengan **Worker Pattern**[cite: 1, 5].

---

### 📂 Project Structure (Feature Slice: `activity`)
Log aktivitas akan dikelola dalam slice tersendiri untuk memantau jejak audit di seluruh sistem[cite: 1, 5].

```text
internal/features/
└── activity/
    ├── repository/     # XADD (Producer) & XREADGROUP (Consumer)
    ├── usecase/        # Event processing logic
    └── dto/            # ActivityEvent Schema
```

---

### 🗄 1. Database Schema (PostgreSQL)
Meskipun Redis Streams menyimpan riwayat, kita memindahkan data lama ke PostgreSQL untuk penyimpanan jangka panjang (archiving) dan pelaporan legal[cite: 1, 6].

**File:** `migrations/202605250007_create_activity_logs_table.sql`[cite: 1]

```sql
-- +goose Up
CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID NOT NULL, -- User/Partner ID yang melakukan aksi[cite: 2, 3]
    action VARCHAR(100) NOT NULL, -- e.g., 'LOGIN', 'UPDATE_LOCATION', 'POINTS_EARNED'
    payload JSONB, -- Data detail terkait aksi[cite: 6]
    ip_address VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE activity_logs;
```

---

### 🧠 2. Redis Streams Strategy (Consumer Groups)
Kita menggunakan **Consumer Groups** agar pemrosesan log bersifat *scalable* dan *fault-tolerant*[cite: 4, 6].

| Komponen | Perintah Redis | Deskripsi |
| :--- | :--- | :--- |
| **Producer** | `XADD stream:activity * ...` | Menambahkan event baru ke stream dengan ID otomatis[cite: 4]. |
| **Setup Group** | `XGROUP CREATE ...` | Membuat grup konsumen untuk membagi beban kerja[cite: 4]. |
| **Consumer** | `XREADGROUP ...` | Membaca pesan baru yang belum diproses oleh grup[cite: 4]. |
| **Reliability** | `XACK` | Konfirmasi bahwa pesan telah berhasil diproses[cite: 4]. |

---

### 🚀 3. Implementasi: Activity Tracking Usecase
Setiap aksi penting di sistem (seperti perubahan lokasi di Fase 5 atau penambahan poin di Fase 6) akan menghasilkan sebuah "Event"[cite: 5, 6].

#### A. Producer (API Server - `cmd/api`)
1.  Setiap kali fungsi kritikal dipanggil, panggil repository `activity` untuk menjalankan `XADD`[cite: 1, 4].
2.  Data yang dikirim: `actor_id`, `action`, dan metadata lainnya[cite: 2, 6].

#### B. Consumer (Background Worker - `cmd/worker`)
1.  Worker bergabung ke grup `cg:activity_processor`[cite: 1, 4].
2.  Membaca pesan secara berkelompok (batching) untuk efisiensi[cite: 1].
3.  Simpan setiap event ke tabel `activity_logs` di PostgreSQL[cite: 1, 6].
4.  Kirim `XACK` ke Redis agar pesan tidak dikirim ulang ke consumer lain[cite: 4].

---

### 💻 4. Implementasi Repository Snippet (Golang)

```go
// internal/features/activity/repository/stream_repo.go

func (r *streamRepo) RecordEvent(ctx context.Context, event map[string]interface{}) error {
    // XADD untuk menambahkan event ke stream
    return r.redis.XAdd(ctx, &redis.XAddArgs{
        Stream: "stream:activity",
        MaxLen: 10000, // Membatasi jumlah log di RAM agar tidak bengkak
        Approx: true,
        Values: event,
    }).Err()[cite: 4]
}

func (r *streamRepo) ReadEvents(ctx context.Context, consumerName string) ([]redis.XMessage, error) {
    // Membaca pesan baru dari group
    streams, err := r.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
        Group:    "cg:activity_processor",
        Consumer: consumerName,
        Streams:  []string{"stream:activity", ">"}, // ">" berarti pesan yang belum pernah dibaca
        Count:    10,
        Block:    0,
    }).Result()[cite: 4]
    
    if err != nil { return nil, err }
    return streams[0].Messages, nil
}
```

---

### 📝 5. Definition of Done (DoD) Fase 9
1.  [ ] Migrasi tabel `activity_logs` berhasil di PostgreSQL[cite: 1].
2.  [ ] Consumer Group berhasil diinisialisasi pada saat aplikasi *startup*[cite: 1, 4].
3.  [ ] Peristiwa penting (Login, Join Partner, Update Location) terekam di Redis Stream[cite: 2, 3, 5].
4.  [ ] Worker berhasil memindahkan data dari Redis ke PostgreSQL dan mengirimkan `XACK`[cite: 1, 4, 6].
5.  [ ] Sistem tetap stabil meskipun terjadi lonjakan volume event yang tinggi (karakteristik utama Redis Streams)[cite: 1, 4].