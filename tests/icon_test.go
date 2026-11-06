package tests

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// PNG magic bytes header: \x89 P N G \r \n \x1a \n
var expectedPNGMagic = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

// ICO magic header: 0x00 0x00 0x01 0x00
var expectedICOMagic = []byte{0x00, 0x00, 0x01, 0x00}

func TestEmbeddedWindowIcon_ValidPNG(t *testing.T) {
	iconPath := filepath.Join("..", "build", "appicon.png")
	data, err := os.ReadFile(iconPath)
	if err != nil {
		t.Fatalf("Failed to read build/appicon.png: %v", err)
	}

	if len(data) < 8 {
		t.Fatalf("Icon file is truncated: len=%d", len(data))
	}

	if !bytes.Equal(data[:8], expectedPNGMagic) {
		t.Errorf("Invalid PNG magic bytes: got %x, want %x", data[:8], expectedPNGMagic)
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to decode PNG config: %v", err)
	}

	if cfg.Width != 512 || cfg.Height != 512 {
		t.Errorf("Expected 512x512 dimensions, got %dx%d", cfg.Width, cfg.Height)
	}
}

func TestWindowsIcon_ValidICO(t *testing.T) {
	icoPath := filepath.Join("..", "build", "windows", "icon.ico")
	data, err := os.ReadFile(icoPath)
	if err != nil {
		t.Fatalf("Failed to read build/windows/icon.ico: %v", err)
	}

	if len(data) < 6 {
		t.Fatalf("ICO file is truncated: len=%d", len(data))
	}

	if !bytes.Equal(data[:4], expectedICOMagic) {
		t.Errorf("Invalid ICO magic header: got %x, want %x", data[:4], expectedICOMagic)
	}

	imageCount := binary.LittleEndian.Uint16(data[4:6])
	if imageCount < 5 {
		t.Errorf("Expected at least 5 mipmap layers in icon.ico, found %d", imageCount)
	}
}

func TestWebFavicon_ExistsAndMatchesMaster(t *testing.T) {
	favPath := filepath.Join("..", "frontend", "public", "icon.png")
	data, err := os.ReadFile(favPath)
	if err != nil {
		t.Fatalf("Failed to read frontend/public/icon.png: %v", err)
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to decode favicon PNG config: %v", err)
	}

	if cfg.Width != 512 || cfg.Height != 512 {
		t.Errorf("Expected 512x512 favicon dimensions, got %dx%d", cfg.Width, cfg.Height)
	}

	// Verify index.html contains icon linkage
	indexPath := filepath.Join("..", "frontend", "index.html")
	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read frontend/index.html: %v", err)
	}

	if !strings.Contains(string(indexHTML), `rel="icon"`) || !strings.Contains(string(indexHTML), `href="/icon.png"`) {
		t.Errorf("frontend/index.html missing <link rel=\"icon\" href=\"/icon.png\" /> tag")
	}
}

func TestPNG_ColorTypeAndBitDepth(t *testing.T) {
	iconPath := filepath.Join("..", "build", "appicon.png")
	data, err := os.ReadFile(iconPath)
	if err != nil {
		t.Fatalf("Failed to read build/appicon.png: %v", err)
	}

	// IHDR chunk starts at byte 8 + 4 bytes length + 4 bytes "IHDR" = 16
	// IHDR structure: Width (4), Height (4), BitDepth (1), ColorType (1)
	if len(data) < 26 {
		t.Fatalf("PNG file too small for IHDR: len=%d", len(data))
	}

	bitDepth := data[24]
	colorType := data[25]

	if bitDepth != 8 {
		t.Errorf("Expected 8-bit depth, got %d", bitDepth)
	}

	// ColorType 6 = RGBA (Truecolor with alpha)
	if colorType != 6 {
		t.Errorf("Expected ColorType 6 (RGBA), got %d", colorType)
	}
}
