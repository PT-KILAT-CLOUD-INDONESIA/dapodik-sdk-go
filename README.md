# dapodik-sdk-go

> **Dibuat dan dirawat oleh [Kilat Cloud](https://github.com/kilat-cloud) — PT Kilat Cloud Indonesia.**
> Port Go resmi dari Kilat Cloud atas [`dapodik-sdk` Python v1.1.0](https://github.com/ardianryan/dapodik-sdk-python) karya **Ryan Ardian & SMA Negeri 1 Gedeg** (lisensi MIT-NC, atribusi dipertahankan — lihat [LICENSE](LICENSE)).

SDK Go modern, ringan (**zero external dependencies** — stdlib saja), untuk integrasi penarikan data **WebService Dapodik Kemendikdasmen** (port 5774).

Port Go dari [`dapodik-sdk` Python v1.1.0](https://github.com/ardianryan/dapodik-sdk-python) karya **Ryan Ardian & SMA Negeri 1 Gedeg**, yang mengadaptasi spesifikasi integrasi WebService Dapodik dari pustaka PHP karya **Ade Reksi Susanto** ([`adereksisusanto/dapodik-api-php`](https://github.com/adereksisusanto/dapodik-api-php)).

> [!IMPORTANT]
> ### Pernyataan Penyangkalan (Disclaimer)
> **`dapodik-sdk-go` adalah pustaka *Unofficial* (tidak resmi) dan independen**, dikembangkan sebagai inisiatif komunitas sumber terbuka, tanpa afiliasi langsung secara struktural dengan Kementerian Pendidikan Dasar dan Menengah (Kemendikdasmen). Seluruh hak cipta nama, logo, dan merek dagang **Dapodik (Data Pokok Pendidikan)** adalah milik sah **Kementerian Pendidikan Dasar dan Menengah Republik Indonesia**.

## 🏢 Dibuat oleh Kilat Cloud

SDK ini dibuat dan dirawat oleh **[Kilat Cloud](https://github.com/kilat-cloud)** (PT Kilat Cloud Indonesia) sebagai port Go dari `dapodik-sdk` Python v1.1.0 karya **Ryan Ardian & SMA Negeri 1 Gedeg**. Kode asli Python tetap diakui penuh — lisensi [MIT-NC](LICENSE) dan atribusi pemegang hak cipta dipertahankan apa adanya. Pertanyaan, laporan bug, dan permintaan fitur silakan lewat [Issues](https://github.com/kilat-cloud/dapodik-sdk-go/issues) repo ini.

## ⚠️ Kepatuhan UU Perlindungan Data Pribadi (UU PDP No. 27/2022)

Aplikasi Dapodik memproses **Data Pribadi Siswa dan Guru** (NIK, NISN, nomor kontak, data orang tua). Pengembang wajib mematuhi **UU Perlindungan Data Pribadi No. 27 Tahun 2022, khususnya Pasal 67**. Jaga kerahasiaan token dan dilarang memublikasikan atau menyalahgunakan data tanpa hak.

## 📥 Instalasi

Membutuhkan Go 1.23+ (range-over-func untuk iterator streaming).

```bash
go get github.com/kilat-cloud/dapodik-sdk-go
```

Tidak ada dependensi eksternal — hanya Go standard library.

## 🚀 Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kilat-cloud/dapodik-sdk-go"
)

func main() {
	client, err := dapodik.NewClient(
		"20300001",               // NPSN
		"TOKEN_WEBSERVICE_DAPODIK", // token dari menu Web Service di aplikasi Dapodik
		dapodik.WithHost("192.168.1.100"), // IP komputer Dapodik
		dapodik.WithPort(5774),            // default port Dapodik
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 1. Profil sekolah
	sekolah, err := client.GetSekolah(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sekolah:", sekolah.First()["nama"])

	// 2. Data siswa (50 baris pertama)
	siswa, err := client.GetPesertaDidik(ctx, 1, 50)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Total ditarik:", siswa.Len(), "siswa")

	// 3. Filter bawaan SDK
	laki := siswa.Filter(map[string]any{"jenis_kelamin": "L"})
	fmt.Println("Jumlah siswa laki-laki:", laki.Len())
}
```

Contoh lengkap: [`examples/quickstart`](examples/quickstart/main.go) — jalankan dengan `DAPODIK_NPSN`, `DAPODIK_TOKEN`, `DAPODIK_HOST`, `DAPODIK_PORT`.

### Factory

```go
dapo := dapodik.NewDapodik("127.0.0.1", 5774, 0) // timeout 0 = default 30 detik
client, err := dapo.API("TOKEN_WEBSERVICE_DAPODIK", "20300001")
```

### Opsi konstruktor

| Opsi | Default | Keterangan |
| --- | --- | --- |
| `WithHost(host)` | `127.0.0.1` | Host WebService |
| `WithPort(port)` | `5774` | Port WebService |
| `WithBaseURL(url)` | — | Base URL langsung, mengesampingkan host/port; trailing `/` dihapus |
| `WithTimeout(d)` | `30s` | Timeout per request |
| `WithHTTPClient(hc)` | — | `*http.Client` kustom (testing / transport kustom) |

Validasi konstruktor sama seperti Python: NPSN dan token wajib diisi (tidak boleh kosong/hanya spasi) dan tidak boleh mengandung `\r`/`\n` (pencegahan CRLF injection). Error validasi bertipe `*dapodik.DapodikError`.

Base URL dibangun sebagai `http://<host>:<port>/WebService`, dengan `http://` ditambahkan bila host belum berawalan `http://` atau `https://` — sama seperti Python.

## 📋 Daftar Endpoint Lengkap

| Endpoint WebService | Python | Go | Deskripsi |
| :--- | :--- | :--- | :--- |
| `/getSekolah` | `client.get_sekolah(sem)` | `client.GetSekolah(ctx, sem, params...)` | Profil & izin operasional sekolah |
| `/getPengguna` | `client.get_pengguna()` | `client.GetPengguna(ctx, params...)` | Akun operator & pengguna Dapodik |
| `/getGtk` | `client.get_gtk(page, limit)` | `client.GetGTK(ctx, page, limit, params...)` | Data Guru & Tenaga Kependidikan |
| `/getRombonganBelajar` | `client.get_rombongan_belajar(sem)` | `client.GetRombonganBelajar(ctx, sem, params...)` | Data rombel beserta anggota & mapel |
| `/getPesertaDidik` | `client.get_peserta_didik(page, limit)` | `client.GetPesertaDidik(ctx, page, limit, params...)` | Data seluruh siswa lengkap |
| `/getPrasarana` | `client.get_prasarana(page, limit)` | `client.GetPrasarana(ctx, page, limit, params...)` | Data sarana & prasarana (Tanah, Bangunan & Ruang) |
| `/getMataPelajaran` | `client.get_mata_pelajaran(sem)` | `client.GetMataPelajaran(ctx, sem, params...)` | Referensi mata pelajaran nasional |
| `/getMatevNilai` | `client.get_matev_nilai(sem)` | `client.GetMatevNilai(ctx, sem, params...)` | Referensi mata evaluasi nilai |
| `/postNilai` | `client.post_nilai(body, table)` | `client.PostNilai(ctx, body, table, params...)` | Pengiriman nilai rapor (HTTP POST) |
| `/postMatevRapor` | `client.post_matev_rapor(body)` | `client.PostMatevRapor(ctx, body, params...)` | Pengiriman mata evaluasi (HTTP POST) |
| *(bebas)* | `client.post(endpoint, body)` | `client.Post(ctx, endpoint, body, params...)` | POST ke endpoint apa pun |
| *(bebas)* | `client.request(method, endpoint, params, body)` | `client.Request(ctx, method, endpoint, params, body)` | Request mentah |

`table` pada `PostNilai` default `"rapor"` bila string kosong, dikirim sebagai query param `table=<table>` — sama seperti Python.

### Idiom parameter

Python memakai `**params` (keyword arguments). Go memakai **variadic `params ...map[string]any`** — satu idiom untuk semua method; map yang disebut belakangan menang bila ada key yang sama. Contoh:

```go
siswa, err := client.GetPesertaDidik(ctx, 1, 50,
	map[string]any{"semester_id": "20241"},
)
```

Aturan parameter, sama seperti Python:

- `page <= 0` dianggap `1`; `limit <= 0` tidak dikirim (sama seperti `limit=None` di Python).
- `semester_id` hanya dikirim bila non-empty.
- Param dengan nilai `nil` tidak dikirim.
- Nilai `bool` di-encode sebagai `"True"`/`"False"` (seperti `str()` Python di `urllib.parse.urlencode`).

### Alias

Sama seperti Python (nama Go diekspor): `Sekolah`, `Pengguna`, `GTK`, `Rombel`, `PD`, `Prasarana`, `MataPelajaran`, `MatevNilai` — masing-masing mendelegasikan ke method `Get*` yang bersangkutan.

## 🔄 Auto-Pagination & Streaming (RAM Efisien)

```go
// Tarik seluruh siswa, 100 per halaman, callback progres tiap halaman.
semua, err := client.FetchAllPesertaDidik(ctx, 100, 0,
	func(page, count, total int) {
		fmt.Printf("halaman %d: %d baris (total %d)\n", page, count, total)
	})
// FetchAllGTK dan FetchAllPrasarana polanya sama.
// delay (time.Duration) diterapkan antar halaman, seperti delay_seconds Python.

// Stream per-batch tanpa menampung semuanya di RAM (range-over-func, Go 1.23+).
for batch, err := range client.IteratePesertaDidik(ctx, 100) {
	if err != nil {
		log.Fatal(err)
	}
	for _, siswa := range batch {
		fmt.Println(siswa["nama"])
	}
}
// Juga tersedia: IterateGTK, IteratePrasarana.
```

Iterator memakai pola **range-over-func** `func(yield func([]map[string]any, error) bool)` — bukan channel — agar tidak ada goroutine yang bocor, context yang dibatalkan berhenti secara natural, dan error fetch halaman dapat disampaikan lewat `yield` (generator Python akan raise; di Go error disampaikan sebagai nilai kedua).

## 📊 Ekspor JSON / CSV

Go tidak punya pandas. Pengganti `to_dataframe()` Python adalah `ToJSON` dan `ToCSV`:

```go
semua, _ := client.FetchAllPesertaDidik(ctx, 100, 0, nil)

// JSON
data, err := semua.ToJSON()

// CSV (header = gabungan semua key, urutan pertama muncul)
f, _ := os.Create("data_siswa.csv")
defer f.Close()
err = semua.ToCSV(f)
```

Method lain pada `Response` (port `DapodikResponse` Python): `Len()`, `Count()`, `First()` (nil bila kosong), `Filter(map[string]any)` (exact equality), `Pluck(key)` (hanya baris yang punya key), `ToList()`.

## 🛡️ Error Handling

| Go | Python | Kapan |
| --- | --- | --- |
| `*dapodik.DapodikError` | `DapodikError` | Validasi, endpoint tidak valid, JSON tidak valid |
| `*dapodik.AuthError` | `DapodikAuthError` | HTTP 401 / 403 |
| `*dapodik.ConnectionError` | `DapodikConnectionError` | Gagal terhubung / timeout / context dibatalkan |
| `*dapodik.HTTPError` | `DapodikHttpError` | HTTP non-2xx lainnya |

Semua error dapat diperiksa dengan `errors.As` — `AuthError` dan `HTTPError` meng-unwrap ke `*DapodikError` (seperti pewarisan di Python), dan `ConnectionError.Unwrap()` mengembalikan error transport asli.

```go
resp, err := client.GetSekolah(ctx, "")
if err != nil {
	var authErr *dapodik.AuthError
	var httpErr *dapodik.HTTPError
	var connErr *dapodik.ConnectionError
	switch {
	case errors.As(err, &authErr):
		log.Fatalf("token tidak valid / IP belum di-whitelist: %v", err)
	case errors.As(err, &httpErr):
		log.Fatalf("HTTP %d saat memanggil %s: %s", httpErr.StatusCode, httpErr.Endpoint, httpErr.RawBody)
	case errors.As(err, &connErr):
		log.Fatalf("tidak bisa terhubung ke Dapodik: %v", err)
	default:
		log.Fatal(err)
	}
}
```

Pesan `HTTPError` sama persis dengan Python: `Dapodik WebService HTTP <kode> (<teks>) saat memanggil endpoint '<endpoint>'`.

## Paritas & perbedaan vs Python

Sumber kebenaran: `dapodik-sdk` Python v1.1.0 (`pyproject.toml`; catatan: `__version__` dan User-Agent Python masih tertulis 1.0.0 — port Go ini memakai 1.1.0).

| Operasi Python | Padanan Go | Status |
| --- | --- | --- |
| `DapodikClient(npsn, token, host, port, base_url, timeout)` | `NewClient(npsn, token, WithHost/WithPort/WithBaseURL/WithTimeout/WithHTTPClient...)` | ✅ Port penuh (functional options) |
| `request(method, endpoint, params, body)` | `Request(ctx, method, endpoint, params, body)` | ✅ Port penuh (+ `context.Context` sebagai argumen pertama, idiom Go) |
| `get_sekolah / sekolah` | `GetSekolah / Sekolah` | ✅ |
| `get_pengguna / pengguna` | `GetPengguna / Pengguna` | ✅ |
| `get_gtk / gtk` | `GetGTK / GTK` | ✅ |
| `get_rombongan_belajar / rombel` | `GetRombonganBelajar / Rombel` | ✅ |
| `get_peserta_didik / pd` | `GetPesertaDidik / PD` | ✅ |
| `get_mata_pelajaran / mata_pelajaran` | `GetMataPelajaran / MataPelajaran` | ✅ |
| `get_matev_nilai / matev_nilai` | `GetMatevNilai / MatevNilai` | ✅ |
| `get_prasarana / prasarana` | `GetPrasarana / Prasarana` | ✅ |
| `post / post_nilai / post_matev_rapor` | `Post / PostNilai / PostMatevRapor` | ✅ |
| `fetch_all_peserta_didik` | `FetchAllPesertaDidik` | ✅ |
| `fetch_all_gtk` | `FetchAllGTK` | ✅ |
| `fetch_all_prasarana` | `FetchAllPrasarana` | ✅ |
| `iterate_peserta_didik` | `IteratePesertaDidik` | ✅ (range-over-func) |
| `iterate_prasarana` | `IteratePrasarana` | ✅ (range-over-func) |
| `iterate_gtk` (tercatat di CHANGELOG 1.0.0, **tidak ada di client.py**) | `IterateGTK` | ✅ **Ditambahkan di Go — celah Python yang diperbaiki** |
| `Dapodik(host, port, timeout)` + `api(token, npsn)` | `NewDapodik(host, port, timeout)` + `API(token, npsn)` | ✅ |
| `DapodikResponse.rows / .first / .count / len()` | `Response.Rows / First() / Count() / Len()` | ✅ |
| `DapodikResponse.filter(**conditions)` | `Response.Filter(map[string]any)` | ✅ (perbandingan numerik dilonggarkan: angka JSON di-decode Go sebagai float64, jadi filter `10` cocok dengan JSON `10`) |
| `DapodikResponse.pluck(key)` | `Response.Pluck(key)` | ✅ |
| `DapodikResponse.to_list()` | `Response.ToList()` | ✅ |
| `DapodikResponse.to_dataframe()` (pandas) | — | ❌ **Tidak ada padanan di Go** — diganti `ToJSON()` dan `ToCSV(w io.Writer)` |
| `DapodikError / DapodikAuthError / DapodikConnectionError / DapodikHttpError` | `DapodikError / AuthError / ConnectionError / HTTPError` | ✅ (nama dipendekkan mengikuti idiom Go; `errors.As`-compatible) |
| Header `User-Agent: dapodik-sdk-python/1.0.0` | `User-Agent: dapodik-sdk-go/1.1.0` | ⚠️ Berubah (nama + versi port) |
| `**params` (kwargs Python) | variadic `params ...map[string]any` | ⚠️ Idiom Go; perilaku penggabungan sama |
| Normalisasi rows (`rows` object/list, `data` list, dict lain, list) | `normalizeResponse` | ✅ Sama persis |
| Validasi NPSN/token (kosong, newline) | `NewClient` | ✅ Sama persis (pesan error sama) |
| Penolakan path traversal (`..`, `\`) | `Request` | ✅ Sama persis |
| Respons kosong -> `DapodikResponse([], {})` | `Response{Rows: [], Raw: {}}` | ✅ |
| JSON tidak valid -> `DapodikError` | `*DapodikError` | ✅ (pesan memotong respons ke 200 karakter, seperti Python) |

Perbedaan yang tidak dapat di-port persis (didokumentasikan, bukan diubah diam-diam):

1. **`to_dataframe()` tidak ada di Go.** Tidak ada pandas di ekosistem Go. Penggantinya `ToJSON()` dan `ToCSV()`. Untuk analisis tabular, hasil `ToList()` dapat diolah dengan `encoding/csv` atau pustaka DataFrame Go pilihan pengguna.
2. **Angka JSON adalah `float64`.** `encoding/json` Go men-decode semua angka menjadi `float64`, sedangkan Python memakai `int` bila bulat. `Filter` sudah menangani ini untuk perbandingan numerik, tetapi kode yang melakukan type assertion `row["umur"].(int)` harus memakai `float64` (atau helper konversi sendiri).
3. **Elemen list non-object dilewati.** `Rows` bertipe `[]map[string]any`; bila respons JSON berupa list berisi nilai non-object (mis. angka/string), elemen tersebut tidak masuk ke `Rows` (Python akan menyimpannya apa adanya). Praktiknya WebService Dapodik selalu mengembalikan list of objects.
4. **`page`/`limit` bertipe `int`, bukan `Optional[int]`.** `limit <= 0` berarti "tidak dikirim", sama seperti `limit=None` di Python; `page <= 0` menjadi `1` (Python tidak menormalisasi `page=0`, tetapi nilai itu tidak bermakna di API).
5. **`ConnectionError` membungkus error asli.** Python menyimpan `original_error`; Go menyimpannya di field `Err` dan `Unwrap()`, sehingga `errors.As`/`errors.Is` bekerja.
6. **Tidak ada varian async.** Python menyediakan extra opsional `httpx`; port Go ini sinkron berbasis `context.Context` (pembatalan/timeout lewat context).

## 🧪 Testing

Semua test memakai `httptest` (mock server lokal) — tidak ada panggilan jaringan ke server Dapodik sungguhan.

```bash
go test ./...
go vet ./...
gofmt -l .
```

## 🚀 Rilis

Rilis berupa **tag + GitHub Release** (library — tidak ada binary). Langkah lengkapnya ada di [CONTRIBUTING.md](CONTRIBUTING.md): perbarui CHANGELOG, buat tag `v1.1.0`, push tag, lalu buat Release (via `goreleaser release --clean` dengan [.goreleaser.yml](.goreleaser.yml), atau lewat web UI GitHub dengan catatan dari CHANGELOG).

## ⚖️ Lisensi & Ketentuan Penggunaan Non-Komersial

Proyek ini dirilis di bawah lisensi **[MIT License with Non-Commercial Restriction (MIT-NC)](LICENSE)** — teks lisensi disalin verbatim dari Python SDK.

### Ketentuan Penggunaan:

1. **100% Gratis untuk Pendidikan**: sepenuhnya gratis digunakan oleh seluruh sekolah, guru, operator, siswa, akademisi, dan lembaga pendidikan di Indonesia.
2. **Dilarang untuk Tujuan Komersial (Non-Commercial Only)**: dilarang keras memperjualbelikan, memonetisasi, menjual kembali (*reselling*), atau mengemas SDK ini ke dalam produk perangkat lunak berbayar / layanan berbayar pihak ketiga tanpa izin tertulis dari pemegang hak cipta (**Ryan Ardian & SMA Negeri 1 Gedeg**).
3. **Atribusi Hak Cipta**: Hak Cipta © 2026 **Ryan Ardian** ([inisaya@ardianryan.com](mailto:inisaya@ardianryan.com)) & **SMA Negeri 1 Gedeg** ([@smansagewithai](https://www.instagram.com/smansagewithai/)). Inspirasi & atribusi dasar: adaptasi pustaka PHP Dapodik oleh **Ade Reksi Susanto** ([`adereksisusanto/dapodik-api-php`](https://github.com/adereksisusanto/dapodik-api-php)).
