package wireguard

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	wgBinary      = "wg"
	wgQuickBinary = "wg-quick"
	wgInterface   = "wg0"
	commandTimeout = 10 * time.Second
)

// AddPeer adds a new peer to the WireGuard interface
func AddPeer(ctx context.Context, publicKey, allowedIP string) error {
	if err := runCommand(ctx, wgBinary, "set", wgInterface, "peer", publicKey, "allowed-ips", allowedIP+"/32"); err != nil {
		return fmt.Errorf("failed to set peer: %w", err)
	}

	if err := runCommand(ctx, wgQuickBinary, "save", wgInterface); err != nil {
		// Try to remove the peer if save fails
		_ = runCommand(context.Background(), wgBinary, "set", wgInterface, "peer", publicKey, "remove")
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// RemovePeer removes a peer from the WireGuard interface
func RemovePeer(ctx context.Context, publicKey string) error {
	if err := runCommand(ctx, wgBinary, "set", wgInterface, "peer", publicKey, "remove"); err != nil {
		return fmt.Errorf("failed to remove peer: %w", err)
	}

	if err := runCommand(ctx, wgQuickBinary, "save", wgInterface); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetStatus returns the status of the WireGuard interface
func GetStatus(ctx context.Context) ([]byte, error) {
	output, err := runCommandWithOutput(ctx, wgBinary, "show", wgInterface)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	return output, nil
}

// GetPublicKey returns the public key of the WireGuard interface
func GetPublicKey(ctx context.Context) (string, error) {
	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	if publicKeyPath == "" {
		publicKeyPath = "/etc/wireguard/publickey"
	}

	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// CheckInterface verifies that the WireGuard interface is available
func CheckInterface(ctx context.Context) error {
	_, err := runCommandWithOutput(ctx, wgBinary, "show", wgInterface)
	return err
}

// Helper function to run command without capturing output
func runCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %v - %s", name, args, stderr.String())
	}
	return nil
}

// Helper function to run command and capture output
func runCommandWithOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %s %v - %s", name, args, stderr.String())
	}
	return stdout.Bytes(), nil
}
