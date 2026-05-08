package quantum

import (
	"os"
	"testing"
)

func TestGenerateKeyDefaultsToKyber1024(t *testing.T) {
	t.Setenv("HADES_ALLOW_SIMULATED_CRYPTO", "true")

	engine, err := NewCryptographyEngine(nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	key, err := engine.GenerateKey("", "public")
	if err != nil {
		t.Fatalf("expected key generation to succeed: %v", err)
	}
	if key.Algorithm != "kyber1024" {
		t.Fatalf("expected default algorithm kyber1024, got %s", key.Algorithm)
	}
}

func TestGenerateKeyFailsWhenSimulationDisabled(t *testing.T) {
	_ = os.Unsetenv("HADES_ALLOW_SIMULATED_CRYPTO")

	engine, err := NewCryptographyEngine(nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	_, err = engine.GenerateKey("kyber1024", "public")
	if err == nil {
		t.Fatalf("expected simulated crypto to be blocked")
	}
}
