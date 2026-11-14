// Package dapodik adalah SDK Go (unofficial, port komunitas) untuk
// WebService Dapodik Kemendikdasmen (port 5774).
//
// Pustaka ini adalah port Go dari dapodik-sdk Python v1.1.0 karya
// Ryan Ardian & SMA Negeri 1 Gedeg (https://github.com/ardianryan/dapodik-sdk-python),
// yang mengadaptasi spesifikasi integrasi WebService Dapodik dari pustaka PHP
// karya Ade Reksi Susanto (https://github.com/adereksisusanto/dapodik-api-php).
//
// SDK ini TIDAK resmi dan TIDAK berafiliasi dengan Kementerian Pendidikan
// Dasar dan Menengah (Kemendikdasmen). Seluruh hak cipta nama, logo, dan
// merek dagang Dapodik (Data Pokok Pendidikan) adalah milik sah Kementerian
// Pendidikan Dasar dan Menengah Republik Indonesia.
//
// Penggunaan:
//
//	client, err := dapodik.NewClient("20300001", "TOKEN_WEBSERVICE",
//		dapodik.WithHost("192.168.1.100"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	sekolah, err := client.GetSekolah(context.Background(), "")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(sekolah.First()["nama"])
//
// WebService Dapodik berjalan di komputer operator sekolah (biasanya
// http://127.0.0.1:5774/WebService). Token diambil dari menu Web Service di
// aplikasi Dapodik. SDK ini hanya berbicara dengan WebService lokal satu
// sekolah — bukan dengan server pusat Dapodik.
//
// PERINGATAN UU PDP: Aplikasi Dapodik memproses Data Pribadi Siswa dan Guru
// (NIK, NISN, nomor kontak, data orang tua). Pengguna SDK wajib mematuhi
// UU Perlindungan Data Pribadi No. 27 Tahun 2022, khususnya Pasal 67.
// Jaga kerahasiaan token dan dilarang memublikasikan atau menyalahgunakan
// data tanpa hak.
package dapodik
