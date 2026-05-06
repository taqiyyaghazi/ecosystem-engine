## 🛠 Tech Stack Snapshot
*   **Language:** Go v1.26.2
*   **Web Framework:** Gin v1.12
*   **Cache:** Redis v9 (Atomic Counters)
*   **Logic:** Fixed Window Algorithm

---

## 📂 Project Structure Integration
Meskipun kita menggunakan *Feature Slice Architecture*, Rate Limiter biasanya bersifat *cross-cutting concern*. Kita akan meletakkannya di dalam folder `platform` agar bisa digunakan oleh seluruh slice (Auth, Services, dll).

```text
internal/platform/
└── middleware/
    └── ratelimit/
        ├── ratelimiter.go   # Logic pembatasan (Redis interaction)
        └── middleware.go    # Gin Middleware wrapper
```

---

## 🧠 1. Redis Rate Limiting Strategy (Fixed Window)
Kita akan menggunakan algoritma **Fixed Window** karena sangat efisien dalam penggunaan memori dan CPU.

| Komponen | Spesifikasi |
| :--- | :--- |
| **Key Pattern** | `ratelimit:{endpoint_path}:{ip_address}` |
| **Data Type** | **Strings** (sebagai Counter) |
| **Limit** | 10 requests |
| **Window Size** | 60 Detik |
| **Status Code** | `429 Too Many Requests` |

---

## 🚀 2. Algoritma & Logika Eksekusi

Setiap request yang masuk akan melewati *pipeline* berikut:

1.  **Identifier Extraction:** Ambil Client IP dari `c.ClientIP()`.
2.  **Key Construction:** Gabungkan path endpoint dan IP menjadi satu key unik.
3.  **Atomic Increment:** Jalankan `INCR <key>`.
4.  **Expiration Check:**
    *   Jika hasil `INCR` adalah **1**, berarti ini request pertama dalam jendela waktu baru. Jalankan `EXPIRE <key> 60`.
5.  **Threshold Validation:**
    *   Jika hasil `INCR` **> 10**, hentikan request dan kirimkan response error.
    *   Jika **<= 10**, lanjutkan ke handler berikutnya (`c.Next()`).



---

## 💻 3. Spesifikasi Teknis Middleware (Golang)

### A. Core Logic (Repository Layer)
```go
// internal/platform/middleware/ratelimit/ratelimiter.go

func (rl *RateLimiter) IsAllowed(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	// 1. Increment secara atomik
	count, err := rl.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// 2. Set expiry hanya pada request pertama
	if count == 1 {
		rl.redis.Expire(ctx, key, window)
	}

	// 3. Cek apakah melebihi ambang batas
	return count <= int64(limit), nil
}
```

### B. Gin Middleware Wrapper
```go
// internal/platform/middleware/ratelimit/middleware.go

func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s:%s", c.FullPath(), c.ClientIP())
		
		allowed, err := rl.IsAllowed(c.Request.Context(), key, 10, 60*time.Second)
		if err != nil || !allowed {
			c.JSON(429, gin.H{
				"error": "Too many requests. Please try again in a minute.",
				"retry_after_seconds": 60,
			})
			c.Abort() // Menghentikan eksekusi handler selanjutnya
			return
		}
		
		c.Next()
	}
}
```

---

## 🛡 4. Response Headers (Best Practice)
Untuk memberikan informasi yang jelas kepada client (atau developer frontend), kita sebaiknya menyertakan header standar pada setiap response:
*   `X-RateLimit-Limit`: 10
*   `X-RateLimit-Remaining`: (10 - current_count)
*   `X-RateLimit-Reset`: Waktu sisa hingga window berakhir.

---

## 📝 5. Definition of Done (DoD) Fase 3
1.  [ ] Middleware dapat diaplikasikan secara global atau per-route di `main.go`.
2.  [ ] Melakukan pengetesan dengan 11 request berturut-turut; request ke-11 harus mengembalikan status `429`.
3.  [ ] Key di Redis otomatis terhapus (expired) setelah 60 detik.
4.  [ ] Tidak ada kebocoran memori; key Redis tidak menumpuk selamanya (wajib ada TTL).
5.  [ ] Logika menggunakan `go-redis/v9` dan berjalan di **Go v1.26.2**.

**Catatan Keamanan:** Fixed window memiliki kelemahan di "batas jendela" (misal: 10 request di akhir detik ke-59 dan 10 request di awal detik ke-61). Namun, untuk tahap awal aplikasi backend ini, metode ini adalah yang paling seimbang antara performa dan kompleksitas.