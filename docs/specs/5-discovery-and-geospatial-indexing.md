## 🛠 Tech Stack Snapshot
*   **Language:** Go v1.26.2
*   **Web Framework:** Gin v1.12
*   **Database:** PostgreSQL (Metadata partner)
*   **Cache:** Redis v9 (**GEO** Data Structure)
*   **Migration:** Goose

---

## 📂 Project Structure (Feature Slice: `discovery`)
Kita akan membuat slice baru bernama `discovery`. Slice ini akan berinteraksi erat dengan data partner yang mungkin sudah ada di database, namun fokus utamanya adalah koordinat di Redis.

```text
internal/features/
├── discovery/
│   ├── delivery/       # GET /nearby, POST /location
│   ├── repository/     # Redis GEO commands (GEOADD, GEOSEARCH)
│   ├── usecase/        # Distance logic & filtering
│   └── dto/            # LocationRequest, NearbyPartnerResponse
```

---

## 🗄 1. Database Schema (PostgreSQL)
Meskipun lokasi *real-time* ada di Redis, kita butuh tabel master partner untuk mengambil detail profil (nama, rating, jenis layanan) saat hasil pencarian jarak muncul.

**File:** `migrations/202605050003_create_partners_table.sql`

```sql
-- +goose Up
CREATE TABLE partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    service_type VARCHAR(50) NOT NULL, -- e.g., 'AC', 'Clean', 'Massage'
    is_active BOOLEAN DEFAULT true,
    rating DECIMAL(2, 1) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE partners;
```

---

## 🧠 2. Redis Geospatial Strategy
Redis menyimpan data GEO di dalam **Sorted Set** secara internal, namun kita menggunakan perintah khusus `GEO` untuk mengelolanya.

| Komponen | Spesifikasi |
| :--- | :--- |
| **Key Pattern** | `partner:locations` |
| **Data Type** | **GEO** (Sorted Set) |
| **Member** | `partner_id` (UUID) |
| **Score** | Geohash (Otomatis oleh Redis) |
| **TTL** | Tidak ada (lokasi partner diupdate terus menerus) |

---

## 🚀 3. API Specification

### A. Update Partner Location
Setiap beberapa detik/menit, aplikasi partner akan mengirimkan koordinat terbaru.
*   **Endpoint:** `POST /v1/discovery/location`
*   **Payload:** `{ "latitude": -6.2088, "longitude": 106.8456 }`
*   **Logic:** 
    1. Ambil `partner_id` dari session (Fase 2).
    2. Jalankan `GEOADD partner:locations <lon> <lat> <partner_id>`.

### B. Find Nearby Partners
Mencari partner di sekitar user untuk ditampilkan di peta atau list.
*   **Endpoint:** `GET /v1/discovery/nearby?lat=-6.2&lon=106.8&radius=5&unit=km`
*   **Logic:**
    1. Gunakan `GEOSEARCH` (lebih modern dibanding `GEORADIUS`) untuk mencari member dalam radius tertentu.
    2. Tambahkan argumen `WITHDIST` untuk mendapatkan jarak presisi.
    3. (Opsional) Ambil detail profil partner dari PostgreSQL berdasarkan ID yang didapat dari Redis.

---

## 💻 4. Implementasi Repository Snippet (Golang)

### A. Update Lokasi (GEOADD)
```go
func (r *discoveryRepo) UpdateLocation(ctx context.Context, partnerID string, lat, lon float64) error {
	return r.redis.GeoAdd(ctx, "partner:locations", &redis.GeoLocation{
		Name:      partnerID,
		Latitude:  lat,
		Longitude: lon,
	}).Err()
}
```

### B. Cari Terdekat (GEOSEARCH)
```go
func (r *discoveryRepo) GetNearby(ctx context.Context, lat, lon, radius float64) ([]redis.GeoLocation, error) {
	return r.redis.GeoSearch(ctx, "partner:locations", &redis.GeoSearchQuery{
		Longitude:  lon,
		Latitude:   lat,
		Radius:     radius,
		RadiusUnit: "km",
		Sort:       "asc", // Urutkan dari yang terdekat
	}).Result()
}
```



---

## 📝 5. Definition of Done (DoD) Fase 4
1.  [ ] Endpoint `POST /location` berhasil memperbarui data di Redis (cek dengan `GEOPOS partner:locations <id>`).
2.  [ ] Endpoint `GET /nearby` mengembalikan daftar ID partner yang berada dalam radius yang ditentukan.
3.  [ ] Response menyertakan jarak (misal: "1.2 km") untuk kenyamanan user.
4.  [ ] Penanganan error jika koordinat yang dikirim tidak valid (Latitude -90 s/d 90, Longitude -180 s/d 180).
5.  [ ] Kode mengikuti standar **Feature Slice** dan menggunakan **Go v1.26.2**.

**Tips Profesional:** Untuk aplikasi skala besar, jangan lupa menghapus lokasi partner dari Redis saat mereka *offline* (menggunakan `ZREM`) agar user tidak memesan partner yang sebenarnya tidak siap bertugas.
