package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceURLConfig(t *testing.T) {
	previous := AppConfig
	t.Cleanup(func() { AppConfig = previous })
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("SSQ_CONFIG", path)
	t.Setenv("SSQ_SOURCE_URL", "")
	t.Setenv("SSQ_FUCAI3D_SOURCE_URL", "https://example.com/3d.txt")
	if err := os.WriteFile(path, []byte("source_url: http://example.com/history.txt\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := InitConfig(); err != nil {
		t.Fatal(err)
	}
	if AppConfig.SSQSourceURL != "http://example.com/history.txt" {
		t.Fatalf("配置文件地址未生效: %q", AppConfig.SSQSourceURL)
	}
	if AppConfig.Fucai3DSourceURL != "https://example.com/3d.txt" {
		t.Fatalf("福彩3D环境变量未生效: %q", AppConfig.Fucai3DSourceURL)
	}
	t.Setenv("SSQ_SOURCE_URL", "https://example.com/override.txt")
	if err := InitConfig(); err != nil {
		t.Fatal(err)
	}
	if AppConfig.SSQSourceURL != "https://example.com/override.txt" {
		t.Fatalf("环境变量地址未生效: %q", AppConfig.SSQSourceURL)
	}
	for _, invalid := range []string{"relative/path", "ftp://example.com/history.txt", "http://"} {
		t.Setenv("SSQ_SOURCE_URL", invalid)
		if err := InitConfig(); err == nil {
			t.Fatalf("应拒绝非法数据源地址: %q", invalid)
		}
	}
	t.Setenv("SSQ_SOURCE_URL", "https://example.com/history.txt")
	t.Setenv("SSQ_FUCAI3D_SOURCE_URL", "file:///tmp/3d.txt")
	if err := InitConfig(); err == nil {
		t.Fatal("应拒绝非法福彩3D数据源地址")
	}
}
