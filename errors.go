package dapodik

import "fmt"

// DapodikError adalah error dasar untuk semua error dari SDK ini.
// Mirip dengan DapodikError di Python; gunakan errors.As untuk
// memeriksa tipe error yang lebih spesifik (AuthError, ConnectionError,
// HTTPError) yang semuanya juga memenuhi *DapodikError melalui Unwrap
// — lihat masing-masing tipe.
type DapodikError struct {
	// Message adalah pesan error yang dapat dibaca manusia.
	Message string
}

// Error mengimplementasikan interface error.
func (e *DapodikError) Error() string {
	if e == nil {
		return "dapodik: unknown error"
	}
	return e.Message
}

// NewDapodikError membuat *DapodikError dari pesan.
func NewDapodikError(msg string) *DapodikError {
	return &DapodikError{Message: msg}
}

// AuthError adalah error autentikasi (HTTP 401 atau 403).
// Mirip dengan DapodikAuthError di Python. Pesan defaultnya sama dengan
// pesan default Python.
type AuthError struct {
	// StatusCode adalah kode status HTTP yang diterima (401 atau 403).
	StatusCode int
	// Message adalah pesan error. Jika kosong, Error() memakai pesan default.
	Message string
}

// DefaultAuthMessage adalah pesan default AuthError (sama seperti Python).
const DefaultAuthMessage = "Akses ditolak (401/403). Pastikan token dan IP client sudah didaftarkan di Pengaturan WebService Dapodik."

// Error mengimplementasikan interface error.
func (e *AuthError) Error() string {
	if e == nil {
		return DefaultAuthMessage
	}
	if e.Message != "" {
		return e.Message
	}
	return DefaultAuthMessage
}

// Unwrap mengembalikan *DapodikError agar errors.As(err, &dapodikErr)
// berhasil untuk AuthError, seperti pewarisan DapodikAuthError dari
// DapodikError di Python.
func (e *AuthError) Unwrap() error {
	return &DapodikError{Message: e.Error()}
}

// NewAuthError membuat *AuthError untuk kode status HTTP tertentu dengan
// pesan seperti yang dipakai Python saat request() gagal autentikasi:
// "Akses ditolak (<kode>). Periksa token atau whitelist IP client di
// Pengaturan WebService Dapodik."
func NewAuthError(statusCode int) *AuthError {
	return &AuthError{
		StatusCode: statusCode,
		Message:    fmt.Sprintf("Akses ditolak (%d). Periksa token atau whitelist IP client di Pengaturan WebService Dapodik.", statusCode),
	}
}

// ConnectionError adalah error saat gagal terhubung ke server Dapodik
// atau request timeout. Mirip dengan DapodikConnectionError di Python.
// Err menyimpan error asli (mis. *url.Error) dan dapat diakses lewat
// Unwrap() agar errors.As / errors.Is tetap bekerja.
type ConnectionError struct {
	// Msg adalah pesan error yang dapat dibaca manusia.
	Msg string
	// Err adalah error asli dari transport HTTP (dapat nil).
	Err error
}

// Error mengimplementasikan interface error.
func (e *ConnectionError) Error() string {
	if e == nil {
		return "dapodik: connection error"
	}
	return e.Msg
}

// Unwrap mengembalikan error asli.
func (e *ConnectionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// HTTPError adalah error saat WebService Dapodik mengembalikan status
// HTTP non-2xx (selain 401/403 yang menjadi AuthError).
// Mirip dengan DapodikHttpError di Python.
type HTTPError struct {
	// StatusCode adalah kode status HTTP yang diterima.
	StatusCode int
	// StatusText adalah teks status HTTP (mis. "Internal Server Error").
	StatusText string
	// Endpoint adalah endpoint yang sedang dipanggil saat error terjadi
	// (tanpa leading "/").
	Endpoint string
	// RawBody adalah isi mentah respons error dari server.
	RawBody string
}

// Error mengimplementasikan interface error. Formatnya sama persis dengan
// Python: "Dapodik WebService HTTP <kode> (<teks>) saat memanggil
// endpoint '<endpoint>'".
func (e *HTTPError) Error() string {
	if e == nil {
		return "dapodik: HTTP error"
	}
	return fmt.Sprintf("Dapodik WebService HTTP %d (%s) saat memanggil endpoint '%s'", e.StatusCode, e.StatusText, e.Endpoint)
}

// Unwrap mengembalikan *DapodikError agar errors.As(err, &dapodikErr)
// berhasil untuk HTTPError, seperti pewarisan DapodikHttpError dari
// DapodikError di Python.
func (e *HTTPError) Unwrap() error {
	return &DapodikError{Message: e.Error()}
}
