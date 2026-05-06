## 🛠 Tech Stack Snapshot
*   **Language:** Go v1.26.2
*   **Web Framework:** Gin v1.12
*   **Cache:** Redis v9 (**Sorted Sets**)
*   **Database:** PostgreSQL (Point History)
*   **Migration:** Goose

---

## 📂 Project Structure (Feature Slice: `leaderboard`)
Kita akan mengisolasi logika ranking di dalam slice `leaderboard`.

```text
internal/features/
├── leaderboard/
│   ├── delivery/       # GET /leaderboard, POST /points
│   ├── repository/     # ZINCRBY, ZREVRANGE, ZREVRANK
│   ├── usecase/        # Ranking logic & point calculation
│   └── dto/            # LeaderboardResponse, PointRequest
```

---

## 🗄 1. Database Schema (PostgreSQL)
Meskipun ranking *real-time* ada di Redis, kita tetap butuh tabel di Postgres sebagai *Source of Truth* untuk riwayat perolehan poin (audit trail).

**File:** `migrations/202605100004_create_point_history_table.sql`

```sql
-- +goose Up
CREATE TABLE point_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id),
    amount INTEGER NOT NULL,
    reason VARCHAR(255), -- e.g., 'Completed Order #123'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE point_history;
```

---

## 🧠 2. Redis Sorted Sets Strategy
Sorted Sets adalah koleksi *member* unik yang masing-masing memiliki *score*. Redis akan otomatis menjaga urutan member berdasarkan score tersebut.

| Komponen | Spesifikasi |
| :--- | :--- |
| **Key Pattern** | `leaderboard:partner` |
| **Data Type** | **Sorted Set (ZSET)** |
| **Member** | `partner_id` (UUID) |
| **Score** | Total Poin (Integer/Float) |
| **Complexity** | **O(log(N))** untuk insert/update dan range query. |



---

## 🚀 3. API Specification

### A. Add Points to Partner
Setiap kali partner menyelesaikan tugas, sistem akan memanggil endpoint ini.
*   **Endpoint:** `POST /v1/leaderboard/points`
*   **Payload:** `{ "partner_id": "uuid", "amount": 10, "reason": "Job Success" }`
*   **Logic:**
    1.  Simpan riwayat ke PostgreSQL (`point_history`).
    2.  Update score di Redis: `ZINCRBY leaderboard:partner <amount> <partner_id>`.

### B. Get Global Leaderboard
Menampilkan daftar partner terbaik (Top 10).
*   **Endpoint:** `GET /v1/leaderboard?limit=10`
*   **Logic:**
    1.  Ambil dari Redis: `ZREVRANGE leaderboard:partner 0 9 WITHSCORES`.
    2.  Hasilnya berupa list `partner_id` dan `score`.

### C. Get My Personal Rank
Melihat posisi ranking partner yang sedang login.
*   **Endpoint:** `GET /v1/leaderboard/me`
*   **Logic:**
    1.  Ambil rank: `ZREVRANK leaderboard:partner <partner_id>`.
    2.  Ambil score: `ZSCORE leaderboard:partner <partner_id>`.
    3.  *Note:* Rank di Redis dimulai dari 0 (tertinggi), jadi tambahkan +1 untuk tampilan user.

---

## 💻 4. Implementasi Repository Snippet (Golang)

```go
// internal/features/leaderboard/repository/leaderboard_repo.go

func (r *leaderboardRepo) IncrementScore(ctx context.Context, partnerID string, amount float64) error {
	// ZIncrBy otomatis membuat member jika belum ada
	return r.redis.ZIncrBy(ctx, "leaderboard:partner", amount, partnerID).Err()
}

func (r *leaderboardRepo) GetTopRank(ctx context.Context, limit int64) ([]redis.Z, error) {
	// ZRevRangeWithScores mengambil dari score terbesar ke terkecil
	return r.redis.ZRevRangeWithScores(ctx, "leaderboard:partner", 0, limit-1).Result()
}

func (r *leaderboardRepo) GetUserRank(ctx context.Context, partnerID string) (int64, float64, error) {
	rank, err := r.redis.ZRevRank(ctx, "leaderboard:partner", partnerID).Result()
	if err != nil {
		return 0, 0, err
	}
	
	score, err := r.redis.ZScore(ctx, "leaderboard:partner", partnerID).Result()
	return rank + 1, score, err
}
```

---

## 📝 5. Definition of Done (DoD) Fase 5
1.  [ ] Migrasi `point_history` berhasil dijalankan.
2.  [ ] Endpoint `POST /points` memperbarui score di Redis dan record di Postgres secara konsisten.
3.  [ ] Endpoint `GET /leaderboard` mengembalikan data yang terurut secara *descending*.
4.  [ ] Penanganan kasus jika `partner_id` tidak ditemukan di Redis (Rank #0 atau Null).
5.  [ ] Performa tetap stabil dengan simulasi ribuan update per detik (ciri khas ZSET).

Sorted Sets ini adalah solusi paling elegan untuk masalah ranking. Jika Anda menggunakan SQL murni, semakin banyak data, query `ORDER BY` akan semakin lambat. Di Redis, kecepatannya hampir konstan.

Sudah siap untuk mengeksekusi spec ini ke dalam kode, atau ada detail perhitungan poin yang ingin Anda diskusikan?