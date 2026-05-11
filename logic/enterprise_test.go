package logic

import (
	"testing"
	"time"
)

func TestShutdownStopsPassiveIncome(t *testing.T) {
	enterprise := NewEnterprise()
	if err := enterprise.Start(); err != nil {
		t.Fatalf("start enterprise: %v", err)
	}

	time.Sleep(2200 * time.Millisecond)

	if _, err := enterprise.Shutdown(); err != nil {
		t.Fatalf("shutdown enterprise: %v", err)
	}

	balanceAfterShutdown := enterprise.Status().Balance
	time.Sleep(2200 * time.Millisecond)

	if got := enterprise.Status().Balance; got != balanceAfterShutdown {
		t.Fatalf("balance changed after shutdown: got %d, want %d", got, balanceAfterShutdown)
	}
}

func TestAddCoalAfterShutdownReturnsStopped(t *testing.T) {
	enterprise := NewEnterprise()
	if err := enterprise.Start(); err != nil {
		t.Fatalf("start enterprise: %v", err)
	}

	if _, err := enterprise.Shutdown(); err != nil {
		t.Fatalf("shutdown enterprise: %v", err)
	}

	if err := enterprise.AddCoal(1); err != ErrAlreadyStopped {
		t.Fatalf("add coal after shutdown: got %v, want %v", err, ErrAlreadyStopped)
	}
}
