// Command quickstart adalah contoh penggunaan cepat dapodik-sdk-go.
// Jalankan dengan:
//
//	DAPODIK_NPSN=20300001 DAPODIK_TOKEN=token DAPODIK_HOST=127.0.0.1 DAPODIK_PORT=5774 go run ./examples/quickstart
//
// Membutuhkan server Dapodik Desktop aktif (WebService port 5774).
// Jika tidak ada server lokal, program mencetak pesan error dengan sopan
// dan keluar dengan kode 0 — contoh ini memang hanya demonstrasi.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/PT-KILAT-CLOUD-INDONESIA/dapodik-sdk-go"
)

func main() {
	npsn := getenv("DAPODIK_NPSN", "20300001")
	token := getenv("DAPODIK_TOKEN", "your_token_here")
	host := getenv("DAPODIK_HOST", "127.0.0.1")
	port, _ := strconv.Atoi(getenv("DAPODIK_PORT", "5774"))

	fmt.Println("Menginisialisasi Client Dapodik...")
	client, err := dapodik.NewClient(npsn, token, dapodik.WithHost(host), dapodik.WithPort(port))
	if err != nil {
		fmt.Println("Info:", err)
		return
	}

	fmt.Println("Base URL:", client.BaseURL)
	fmt.Println("NPSN:", client.NPSN)

	// Catatan: pemanggilan di bawah membutuhkan server Dapodik Desktop
	// aktif di port 5774.
	ctx := context.Background()
	sekolah, err := client.GetSekolah(ctx, "")
	if err != nil {
		printErr(err)
		return
	}
	nama := "N/A"
	if first := sekolah.First(); first != nil {
		if v, ok := first["nama"].(string); ok {
			nama = v
		}
	}
	fmt.Println("Nama Sekolah:", nama)

	pd, err := client.GetPesertaDidik(ctx, 1, 10)
	if err != nil {
		printErr(err)
		return
	}
	fmt.Println("Jumlah Siswa Ditarik:", pd.Len())

	// Contoh filter bawaan SDK.
	laki := pd.Filter(map[string]any{"jenis_kelamin": "L"})
	fmt.Println("Jumlah siswa laki-laki (dari 10 baris pertama):", laki.Len())
}

func printErr(err error) {
	var authErr *dapodik.AuthError
	switch {
	case errors.As(err, &authErr):
		fmt.Println("Info: autentikasi gagal —", err)
	default:
		fmt.Println("Info:", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
