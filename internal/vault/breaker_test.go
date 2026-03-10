package vault

import (
	"errors"
	"testing"
	"time"
)

func TestWithCircuitBreaker(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	stateChanges := []string{}
	onStateChange := func(from, to string) {
		stateChanges = append(stateChanges, from+"->"+to)
	}

	config := BreakerConfig{
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
	}

	client.WithCircuitBreaker(config, onStateChange)

	if client.breaker == nil {
		t.Fatal("expected circuit breaker to be set")
	}
}

func TestExecuteWithBreaker_Success(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := BreakerConfig{
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
	}

	client.WithCircuitBreaker(config, nil)

	result, err := client.executeWithBreaker(func() (interface{}, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got: %v", result)
	}
}

func TestExecuteWithBreaker_Failure(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := BreakerConfig{
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
	}

	client.WithCircuitBreaker(config, nil)

	testErr := errors.New("test error")
	_, err = client.executeWithBreaker(func() (interface{}, error) {
		return nil, testErr
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestExecuteWithBreaker_OpensAfterFailures(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	stateChanges := []string{}
	onStateChange := func(from, to string) {
		stateChanges = append(stateChanges, to)
	}

	config := BreakerConfig{
		MaxRequests: 1,
		Interval:    1 * time.Second,
		Timeout:     100 * time.Millisecond,
	}

	client.WithCircuitBreaker(config, onStateChange)

	testErr := errors.New("test error")
	for i := 0; i < 5; i++ {
		_, _ = client.executeWithBreaker(func() (interface{}, error) {
			return nil, testErr
		})
	}

	hasOpen := false
	for _, state := range stateChanges {
		if state == "open" {
			hasOpen = true
			break
		}
	}

	if !hasOpen {
		t.Error("expected circuit breaker to open after failures")
	}
}

func TestExecuteWithBreaker_NoBreaker(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.executeWithBreaker(func() (interface{}, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error without breaker, got: %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got: %v", result)
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	stateChanges := []string{}
	onStateChange := func(from, to string) {
		stateChanges = append(stateChanges, from+"->"+to)
	}

	config := BreakerConfig{
		MaxRequests: 1,
		Interval:    200 * time.Millisecond,
		Timeout:     200 * time.Millisecond,
	}

	client.WithCircuitBreaker(config, onStateChange)

	testErr := errors.New("test error")
	for i := 0; i < 5; i++ {
		_, _ = client.executeWithBreaker(func() (interface{}, error) {
			return nil, testErr
		})
	}

	time.Sleep(250 * time.Millisecond)

	_, _ = client.executeWithBreaker(func() (interface{}, error) {
		return "success", nil
	})

	if len(stateChanges) == 0 {
		t.Error("expected state changes, got none")
	}
}
