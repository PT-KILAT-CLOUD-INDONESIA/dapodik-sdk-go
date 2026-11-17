# Panduan Kontribusi (Contributing Guidelines)

Terima kasih telah tertarik berkontribusi pada pengembangan **`dapodik-sdk-go`**!

Port Go ini diadaptasi dari `dapodik-sdk` Python karya **Ryan Ardian & SMA Negeri 1 Gedeg**. Atribusi dan lisensi MIT-NC tetap berlaku — lihat [LICENSE](LICENSE) dan [README.md](README.md).

Proyek ini dibuat dan dirawat oleh **[Kilat Cloud](https://github.com/PT-KILAT-CLOUD-INDONESIA)** (PT Kilat Cloud Indonesia).

---

## 🛠️ Pengembangan Lokal

1. **Fork** repositori ini ke akun GitHub Anda.
2. **Clone** hasil fork:
   ```bash
   git clone https://github.com/PT-KILAT-CLOUD-INDONESIA/dapodik-sdk-go.git
   cd dapodik-sdk-go
   ```
3. **Jalankan pengujian unit** (memakai `httptest`, tanpa server Dapodik sungguhan):
   ```bash
   go test -count=1 ./...
   ```

Membutuhkan Go 1.23+. Tidak ada dependensi eksternal — hanya Go standard library.

---

## 🌿 Standar Kode & Pull Request

1. Ikuti `gofmt` (`test -z "$(gofmt -l .)"` harus kosong) dan pastikan `go vet ./...` bersih.
2. Pastikan seluruh pengujian unit lulus 100%.
3. Jaga **paritas dengan dapodik-sdk Python** (`ardianryan/dapodik-sdk-python`): bila menambah/mengubah perilaku endpoint, perbarui tabel "Paritas & perbedaan vs Python" di README. Bila suatu perilaku Python tidak dapat di-port persis, catat alasannya di README — jangan mengubahnya diam-diam.
4. Gunakan pesan commit deskriptif (*Conventional Commits*).
5. Jangan pernah mengunggah token WebService Dapodik atau data pribadi siswa/guru riil.

---

## 🚀 Rilis (hanya untuk pemelihara)

Repo ini adalah **library** — rilis berupa tag + GitHub Release, bukan binary (lihat [.goreleaser.yml](.goreleaser.yml)).

1. Perbarui [CHANGELOG.md](CHANGELOG.md) untuk versi baru (format Keep a Changelog).
2. Pastikan CI hijau: `gofmt` bersih, `go vet ./...` bersih, `go test -count=1 ./...` lulus.
3. Commit, lalu buat dan push tag:
   ```bash
   git tag -a v1.1.0 -m "dapodik-sdk-go v1.1.0"
   git push origin v1.1.0
   ```
4. Buat GitHub Release dari tag tersebut, salah satu cara:
   - **GoReleaser** (bila terpasang): `goreleaser release --clean` — hanya membuat release + catatan rilis, tidak membangun binary.
   - **Web UI GitHub**: buat Release dari tag, isi catatan rilis dari entri CHANGELOG versi tersebut.
5. Verifikasi instalasi dari tag: `go get github.com/PT-KILAT-CLOUD-INDONESIA/dapodik-sdk-go@v1.1.0`.
