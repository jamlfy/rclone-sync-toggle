package i18n

import "testing"

func TestInitSelectsLanguage(t *testing.T) {
	tests := []struct {
		language string
		want     string
	}{
		{language: "en", want: "Manage Cloud Folders"},
		{language: "en_US.UTF-8", want: "Manage Cloud Folders"},
		{language: "es-MX", want: "Gestionar Carpetas en Nube"},
	}

	for _, tt := range tests {
		if err := Init(tt.language); err != nil {
			t.Fatalf("Init(%q) returned an error: %v", tt.language, err)
		}
		if got := T("ui.manager_title"); got != tt.want {
			t.Errorf("T(ui.manager_title) after Init(%q) = %q, want %q", tt.language, got, tt.want)
		}
	}
}

func TestInitRejectsUnsupportedLanguage(t *testing.T) {
	if err := Init("fr"); err == nil {
		t.Fatal("Init(fr) did not reject an unsupported language")
	}
}

func TestInitFallsBackToSpanishForUnsupportedSystemLocale(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "fr_FR.UTF-8")

	if err := Init(""); err != nil {
		t.Fatalf("Init() returned an error: %v", err)
	}
	if got, want := T("ui.manager_title"), "Gestionar Carpetas en Nube"; got != want {
		t.Errorf("T(ui.manager_title) = %q, want %q", got, want)
	}
}
