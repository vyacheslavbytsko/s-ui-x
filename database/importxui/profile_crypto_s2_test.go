package importxui

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"path/filepath"
	"testing"

	"github.com/deposist/s-ui-x/config"
	"github.com/deposist/s-ui-x/database"
	"github.com/deposist/s-ui-x/database/model"
	"golang.org/x/crypto/hkdf"
)

// TestProfileEncryptionSettingsSecretSeedAndLegacyFallback covers S2: when a
// random per-install settings.secret is available, new profiles are sealed under
// it instead of the predictable config.GetSecret() value — and a profile
// previously sealed under the legacy seed still decrypts via the fallback
// candidate, so upgrading does not orphan stored remote-panel credentials.
func TestProfileEncryptionSettingsSecretSeedAndLegacyFallback(t *testing.T) {
	t.Setenv("XUI_PROFILE_KEY_FILE", "")
	t.Setenv("SUI_SECRET", "")

	dir := t.TempDir()
	if err := database.InitDB(filepath.Join(dir, "s-ui.db")); err != nil {
		t.Fatal(err)
	}
	// Close the SQLite handle before t.TempDir's RemoveAll so Windows can delete
	// the db file (this cleanup is registered after TempDir's, so it runs first).
	t.Cleanup(func() {
		if d := database.GetDB(); d != nil {
			if sqlDB, err := d.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
	})
	db := database.GetDB()
	if db == nil {
		t.Fatal("expected initialized database")
	}
	const installSecret = "s2-test-install-secret-0123456789ab"
	db.Where("key = ?", "secret").Delete(&model.Setting{})
	if err := db.Create(&model.Setting{Key: "secret", Value: installSecret}).Error; err != nil {
		t.Fatal(err)
	}

	// The per-install secret must differ from the legacy seed, otherwise the
	// fallback path below would not actually be exercised.
	if installSecret == config.GetSecret() {
		t.Fatal("test setup: settings.secret unexpectedly equals the legacy seed")
	}

	plaintext := []byte(`{"type":"xuihttp","baseUrl":"https://panel.example.com","password":"s3cr3t-pass"}`)

	// New scheme: encrypt uses the settings.secret seed and round-trips.
	ciphertext, salt, err := EncryptProfileSource(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ciphertext, []byte("s3cr3t-pass")) {
		t.Fatalf("ciphertext leaked plaintext: %q", ciphertext)
	}
	got, err := DecryptProfileSource(ciphertext, salt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("round-trip mismatch: %s", got)
	}

	// The encryption key must derive from settings.secret, not the legacy seed.
	keys, err := profileEncryptionKeys(salt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(keys[0], deriveProfileKeyForTest(t, []byte(installSecret), salt)) {
		t.Fatal("encryption key is not derived from settings.secret")
	}

	// Backward compatibility: a profile sealed under the legacy config.GetSecret()
	// seed must still decrypt after migrating to settings.secret.
	legacySalt := make([]byte, 16)
	if _, err := rand.Read(legacySalt); err != nil {
		t.Fatal(err)
	}
	legacyKey := deriveProfileKeyForTest(t, []byte(config.GetSecret()), legacySalt)
	legacyCiphertext := sealForTest(t, legacyKey, plaintext)
	got, err = DecryptProfileSource(legacyCiphertext, legacySalt)
	if err != nil {
		t.Fatalf("legacy ciphertext failed to decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("legacy plaintext mismatch: %s", got)
	}
}

// TestProfileEncryptionSurvivesAddingSUISecret guards the S2 transition the
// at-rest warning recommends: a profile sealed under settings.secret (no
// SUI_SECRET) must STILL decrypt after the operator later sets SUI_SECRET.
// With the seed candidates else-chained this orphaned the stored credentials.
func TestProfileEncryptionSurvivesAddingSUISecret(t *testing.T) {
	t.Setenv("XUI_PROFILE_KEY_FILE", "")
	t.Setenv("SUI_SECRET", "")

	dir := t.TempDir()
	if err := database.InitDB(filepath.Join(dir, "s-ui.db")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if d := database.GetDB(); d != nil {
			if sqlDB, err := d.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
	})
	db := database.GetDB()
	db.Where("key = ?", "secret").Delete(&model.Setting{})
	if err := db.Create(&model.Setting{Key: "secret", Value: "s2-add-secret-install-0123456789ab"}).Error; err != nil {
		t.Fatal(err)
	}

	plaintext := []byte(`{"type":"xuihttp","baseUrl":"https://panel.example.com","password":"s3cr3t"}`)

	// Seal while SUI_SECRET is unset (uses the settings.secret seed).
	ciphertext, salt, err := EncryptProfileSource(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	// Operator now sets SUI_SECRET, as the log warning recommends.
	t.Setenv("SUI_SECRET", "an-operator-chosen-out-of-db-secret")

	// The pre-existing profile must still decrypt: settings.secret stays a
	// decrypt candidate even once SUI_SECRET is set.
	got, err := DecryptProfileSource(ciphertext, salt)
	if err != nil {
		t.Fatalf("profile sealed under settings.secret was orphaned after setting SUI_SECRET: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("plaintext mismatch after SUI_SECRET transition: %s", got)
	}
}

func deriveProfileKeyForTest(t *testing.T, seed, rowSalt []byte) []byte {
	t.Helper()
	salt := append([]byte(profileKeySalt+":"), rowSalt...)
	reader := hkdf.New(sha256.New, seed, salt, nil)
	key := make([]byte, 32)
	if _, err := io.ReadFull(reader, key); err != nil {
		t.Fatal(err)
	}
	return key
}

func sealForTest(t *testing.T, key, plaintext []byte) []byte {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}
	out := append([]byte{}, nonce...)
	return gcm.Seal(out, nonce, plaintext, nil)
}
