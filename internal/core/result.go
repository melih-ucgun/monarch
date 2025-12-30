package core

// Result, bir kaynağın uygulanması (Apply) sonucunda dönen değerdir.
// Bu yapı, sadece hatayı değil, neyin değiştiğini ve kullanıcıya gösterilecek mesajı da içerir.
type Result struct {
	// Changed: Sistemde bir değişiklik yapıldı mı?
	Changed bool

	// Failed: İşlem başarısız mı oldu? (Error != nil kontrolü yerine flag olarak da tutabiliriz)
	Failed bool

	// Message: Kullanıcıya gösterilecek insan tarafından okunabilir mesaj.
	Message string

	// Error: Eğer işlem başarısızsa teknik hata detayı.
	Error error
}

// SuccessChange, başarılı ve değişiklik içeren bir sonuç döner.
func SuccessChange(msg string) Result {
	return Result{
		Changed: true,
		Failed:  false,
		Message: msg,
	}
}

// SuccessNoChange, başarılı ama değişiklik içermeyen bir sonuç döner.
func SuccessNoChange(msg string) Result {
	return Result{
		Changed: false,
		Failed:  false,
		Message: msg,
	}
}

// Failure, başarısız bir sonuç döner.
func Failure(err error, msg string) Result {
	return Result{
		Changed: false,
		Failed:  true,
		Message: msg,
		Error:   err,
	}
}
