# Changelog

Semua perubahan penting pada paket **`dapodik-sdk-go`** akan didokumentasikan di file ini.

Format changelog ini mengacu pada [Keep a Changelog](https://keepachangelog.com/id-ID/1.1.0/), dan proyek ini mematuhi [Semantic Versioning](https://semver.org/lang/id/).

---

## [1.1.0] - 2026-11-14

### Ditambahkan
- **Port Go penuh dari dapodik-sdk Python v1.1.0** (Ryan Ardian & SMA Negeri 1 Gedeg):
  - `NewClient` dengan functional options (`WithHost`, `WithPort`, `WithBaseURL`, `WithTimeout`, `WithHTTPClient`); validasi NPSN/token sama seperti Python.
  - Seluruh endpoint GET (`GetSekolah`, `GetPengguna`, `GetGTK`, `GetRombonganBelajar`, `GetPesertaDidik`, `GetMataPelajaran`, `GetMatevNilai`, `GetPrasarana`) beserta alias (`Sekolah`, `Pengguna`, `GTK`, `Rombel`, `PD`, `Prasarana`, `MataPelajaran`, `MatevNilai`).
  - Operasi tulis (`Post`, `PostNilai` dengan `table` default `"rapor"`, `PostMatevRapor`).
  - Auto-pagination (`FetchAllPesertaDidik`, `FetchAllGTK`, `FetchAllPrasarana`) dengan `delay` dan callback progres `ProgressFunc(page, count, total)`.
  - Streaming range-over-func Go 1.23 (`IteratePesertaDidik`, `IteratePrasarana`).
  - Wrapper `Response` (`Len`, `Count`, `First`, `Filter`, `Pluck`, `ToList`) dengan normalisasi rows yang sama seperti Python.
  - Hierarki error `DapodikError` / `AuthError` / `ConnectionError` / `HTTPError` yang kompatibel dengan `errors.As`.
  - Factory `NewDapodik` + `API(token, npsn)` (port class `Dapodik` Python).
- **Ekspor `ToJSON()` dan `ToCSV()`** sebagai pengganti `to_dataframe()` (pandas) yang tidak ada di Go.
- **`IterateGTK` — perbaikan celah Python**: CHANGELOG Python 1.0.0 mendokumentasikan `iterate_gtk`, tetapi `client.py` Python tidak mengimplementasikannya. Port Go ini mengimplementasikannya dengan pola yang sama seperti `IteratePesertaDidik` / `IteratePrasarana`.
- Contoh `examples/quickstart` (env `DAPODIK_NPSN` / `DAPODIK_TOKEN` / `DAPODIK_HOST` / `DAPODIK_PORT`).
- Unit test dengan `httptest` (tanpa jaringan ke server Dapodik sungguhan).
- Penanda Kilat Cloud: README, `doc.go`, dan CONTRIBUTING menegaskan SDK ini dibuat dan dirawat oleh Kilat Cloud (PT Kilat Cloud Indonesia).

### Diubah
- User-Agent menjadi `dapodik-sdk-go/1.1.0` (Python: `dapodik-sdk-python/1.0.0` — `__version__` Python masih 1.0.0 padahal `pyproject.toml` sudah 1.1.0; port Go mengikuti versi 1.1.0).
- `**params` Python menjadi variadic `params ...map[string]any` (idiom Go).
- `to_dataframe()` diganti `ToJSON()` / `ToCSV()` (lihat README, bagian "Paritas & perbedaan vs Python").
