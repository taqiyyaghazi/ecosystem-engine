### 🛠 Tech Stack Snapshot
*   **Language:** Go v1.26.2[cite: 1].
*   **Web Framework:** Gin v1.12[cite: 1].
*   **Database:** PostgreSQL (Task Tracking)[cite: 1].
*   **Cache/Queue:** Redis v9 (**Lists** data structure)[cite: 1, 4].
*   **Architecture:** Feature Slice (Vertical Slice) dengan **Worker Pattern**[cite: 1].

---

### 📂 Project Structure (Feature Slice: `tasks`)
Kita akan menambahkan slice baru bernama `tasks` yang akan digunakan oleh fitur lain (seperti `services`) untuk mendelegasikan tugas berat[cite: 1].

```text
internal/features/
└── tasks/
    ├── repository/     # LPUSH (Producer) & BRPOP (Consumer)
    ├── usecase/        # Task logic (e.g., Invoice generation)
    └── dto/            # TaskPayload (JSON)
```

---

### 🗄 1. Database Schema (PostgreSQL)
Meskipun antrean berada di Redis, kita membutuhkan tabel di PostgreSQL untuk melacak status pengerjaan tugas (Audit Trail)[cite: 6].

**File:** `migrations/202605200006_create_tasks_table.sql`[cite: 1]

```sql
-- +goose Up
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_type VARCHAR(50) NOT NULL, -- e.g., 'GENERATE_INVOICE'
    payload JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, PROCESSING, COMPLETED, FAILED
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE tasks;
```

---

### 🧠 2. Redis Lists Strategy (Producer-Consumer)
Kita menggunakan pola **FIFO (First-In, First-Out)** untuk memastikan tugas diproses sesuai urutan masuk[cite: 4].

| Komponen | Perintah Redis | Fungsi |
| :--- | :--- | :--- |
| **Producer** | `LPUSH queue:default` | Menambahkan tugas baru ke ujung antrean[cite: 4]. |
| **Consumer** | `BRPOP queue:default 0` | Mengambil tugas dari ujung lainnya secara *blocking* (menunggu jika kosong)[cite: 4]. |
| **Data Format** | JSON String | Berisi `task_id` agar worker bisa memperbarui status di PostgreSQL[cite: 1, 6]. |



---

### 🚀 3. Implementasi: Invoice Generation Usecase
Usecase paling tepat untuk fase ini adalah pembuatan invoice PDF setelah layanan selesai, karena proses ini memakan waktu dan resource CPU yang besar[cite: 1].

#### A. Producer (API Server - `cmd/api`)
Saat transaksi selesai, API tidak langsung membuat PDF, melainkan hanya memasukkan tugas ke antrean[cite: 1].
1.  Simpan record tugas ke PostgreSQL dengan status `PENDING`[cite: 6].
2.  Jalankan `LPUSH queue:default` dengan payload `{ "task_id": "...", "type": "GENERATE_INVOICE" }`[cite: 4].
3.  Kirim respon sukses ke user dengan cepat[cite: 1].

#### B. Consumer (Background Worker - `cmd/worker`)
Worker yang berjalan secara terpisah akan terus memantau antrean[cite: 1].
1.  Worker menjalankan `BRPOP` secara terus menerus[cite: 4].
2.  Saat tugas diterima, status di DB diubah menjadi `PROCESSING`[cite: 6].
3.  Worker melakukan simulasi pembuatan PDF (misal: *sleep* 3 detik)[cite: 1].
4.  Setelah selesai, status di DB diubah menjadi `COMPLETED`[cite: 6].

---

### 💻 4. Implementasi Repository Snippet (Golang)

```go
// internal/features/tasks/repository/queue_repo.go

func (r *queueRepo) PushTask(ctx context.Context, queueName string, taskID string) error {
	// LPUSH untuk memasukkan task ID ke antrean
	return r.redis.LPush(ctx, queueName, taskID).Err()[cite: 4]
}

func (r *queueRepo) PopTask(ctx context.Context, queueName string) (string, error) {
	// BRPOP dengan timeout 0 (block selamanya hingga ada data)
	res, err := r.redis.BRPop(ctx, 0, queueName).Result()[cite: 4]
	if err != nil {
		return "", err
	}
	// BRPop mengembalikan []string{key, value}
	return res[1], nil
}
```

---

### 📝 5. Definition of Done (DoD) Fase 8
1.  [ ] Tabel `tasks` berhasil dideploy melalui migrasi `goose`[cite: 1].
2.  [ ] API Server berhasil memasukkan tugas ke Redis menggunakan `LPUSH`[cite: 4].
3.  [ ] Background Worker (`cmd/worker`) berhasil mengambil tugas menggunakan `BRPOP` dan memprosesnya satu per satu[cite: 1, 4].
4.  [ ] Status tugas di PostgreSQL berubah secara otomatis dari `PENDING` ke `COMPLETED`[cite: 6].
5.  [ ] Jika worker dimatikan, tugas yang belum diproses tetap aman berada di dalam Redis List[cite: 4].
