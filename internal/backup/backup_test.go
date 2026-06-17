package backup

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestManager(t *testing.T) *Manager {
	t.Helper()

	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	configDir := filepath.Join(tmpDir, "configs")
	stateDir := filepath.Join(tmpDir, "state")

	os.MkdirAll(backupDir, 0755)
	os.MkdirAll(configDir, 0755)
	os.MkdirAll(stateDir, 0755)

	os.WriteFile(filepath.Join(configDir, "test.yaml"), []byte("test: config"), 0644)
	os.WriteFile(filepath.Join(stateDir, "taws.json"), []byte(`{"installed":true}`), 0644)

	return NewManager(backupDir, configDir, stateDir)
}

func TestCreateSnapshot(t *testing.T) {
	m := setupTestManager(t)

	snapshot, err := m.CreateSnapshot("1.0.0")
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	if snapshot.ID == "" {
		t.Error("snapshot ID should not be empty")
	}

	if snapshot.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", snapshot.Version)
	}

	if len(snapshot.Files) == 0 {
		t.Error("snapshot should contain files")
	}

	if snapshot.Size == 0 {
		t.Error("snapshot size should be greater than 0")
	}
}

func TestListSnapshots(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.CreateSnapshot("1.0.0")
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	snapshots, err := m.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}

	if len(snapshots) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(snapshots))
	}
}

func TestGetSnapshot(t *testing.T) {
	m := setupTestManager(t)

	snapshot, _ := m.CreateSnapshot("1.0.0")

	found, err := m.GetSnapshot(snapshot.ID)
	if err != nil {
		t.Fatalf("GetSnapshot failed: %v", err)
	}

	if found.ID != snapshot.ID {
		t.Errorf("expected ID %s, got %s", snapshot.ID, found.ID)
	}
}

func TestGetSnapshotNotFound(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.GetSnapshot("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent snapshot")
	}
}

func TestDeleteSnapshot(t *testing.T) {
	m := setupTestManager(t)

	snapshot, _ := m.CreateSnapshot("1.0.0")

	if err := m.DeleteSnapshot(snapshot.ID); err != nil {
		t.Fatalf("DeleteSnapshot failed: %v", err)
	}

	snapshots, _ := m.ListSnapshots()
	if len(snapshots) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snapshots))
	}
}

func TestRestoreSnapshot(t *testing.T) {
	m := setupTestManager(t)

	snapshot, _ := m.CreateSnapshot("1.0.0")

	if err := m.RestoreSnapshot(snapshot.ID); err != nil {
		t.Fatalf("RestoreSnapshot failed: %v", err)
	}
}

func TestRestoreSnapshotNotFound(t *testing.T) {
	m := setupTestManager(t)

	err := m.RestoreSnapshot("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent snapshot")
	}
}

func TestCollectFiles(t *testing.T) {
	m := setupTestManager(t)

	files, err := m.collectFiles(m.configDir)
	if err != nil {
		t.Fatalf("collectFiles failed: %v", err)
	}

	if len(files) == 0 {
		t.Error("expected at least one file")
	}
}

func TestMetadataPersistence(t *testing.T) {
	m := setupTestManager(t)

	m.CreateSnapshot("1.0.0")
	m.CreateSnapshot("1.0.1")

	metadata, err := m.loadMetadata()
	if err != nil {
		t.Fatalf("loadMetadata failed: %v", err)
	}

	if len(metadata.Snapshots) != 2 {
		t.Errorf("expected 2 snapshots in metadata, got %d", len(metadata.Snapshots))
	}
}

func TestFormatSnapshots(t *testing.T) {
	snapshots := []Snapshot{
		{
			ID:        "20260101-120000",
			Version:   "1.0.0",
			Timestamp: time.Now(),
			Files:     []string{"config.yaml", "taws.json"},
			Size:      1024,
		},
	}

	output := FormatSnapshots(snapshots)

	if output == "" {
		t.Fatal("FormatSnapshots returned empty string")
	}

	if len(output) < 30 {
		t.Error("FormatSnapshots output too short")
	}
}

func TestFormatSnapshotsEmpty(t *testing.T) {
	output := FormatSnapshots([]Snapshot{})

	if output == "" {
		t.Fatal("FormatSnapshots returned empty string")
	}
}

func TestExtractArchiveZipSlip(t *testing.T) {
	m := setupTestManager(t)

	tmpDir := t.TempDir()
	evilArchive := filepath.Join(tmpDir, "evil.tar.gz")

	f, err := os.Create(evilArchive)
	if err != nil {
		t.Fatalf("creating archive: %v", err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	header := &tar.Header{
		Name:     "../../../etc/passwd",
		Mode:     0644,
		Size:     12,
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("writing header: %v", err)
	}
	if _, err := tw.Write([]byte("malicious")); err != nil {
		t.Fatalf("writing content: %v", err)
	}

	tw.Close()
	gw.Close()
	f.Close()

	err = m.extractArchive(evilArchive)
	if err == nil {
		t.Error("expected error for zip slip attempt")
	}
}

func TestExtractArchiveSymlink(t *testing.T) {
	m := setupTestManager(t)

	tmpDir := t.TempDir()
	evilArchive := filepath.Join(tmpDir, "evil.tar.gz")

	f, err := os.Create(evilArchive)
	if err != nil {
		t.Fatalf("creating archive: %v", err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	header := &tar.Header{
		Name:     "link.txt",
		Mode:     0777,
		Typeflag: tar.TypeSymlink,
		Linkname: "/etc/passwd",
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("writing symlink header: %v", err)
	}

	header2 := &tar.Header{
		Name:     "link.txt",
		Mode:     0644,
		Size:     12,
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(header2); err != nil {
		t.Fatalf("writing reg header: %v", err)
	}
	if _, err := tw.Write([]byte("malicious")); err != nil {
		t.Fatalf("writing content: %v", err)
	}

	tw.Close()
	gw.Close()
	f.Close()

	err = m.extractArchive(evilArchive)
	if err == nil {
		t.Error("expected error for symlink attempt")
	}
}

func TestExtractArchiveHardlinkSkipped(t *testing.T) {
	m := setupTestManager(t)

	tmpDir := t.TempDir()
	evilArchive := filepath.Join(tmpDir, "evil.tar.gz")

	f, err := os.Create(evilArchive)
	if err != nil {
		t.Fatalf("creating archive: %v", err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	header := &tar.Header{
		Name:     "target.txt",
		Mode:     0644,
		Size:     6,
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("writing target header: %v", err)
	}
	if _, err := tw.Write([]byte("target")); err != nil {
		t.Fatalf("writing target: %v", err)
	}

	header2 := &tar.Header{
		Name:     "link.txt",
		Mode:     0644,
		Typeflag: tar.TypeLink,
		Linkname: "target.txt",
	}
	if err := tw.WriteHeader(header2); err != nil {
		t.Fatalf("writing hardlink header: %v", err)
	}

	tw.Close()
	gw.Close()
	f.Close()

	err = m.extractArchive(evilArchive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	targetPath := filepath.Join(filepath.Dir(m.configDir), "target.txt")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("target file should have been extracted")
	}
}

func createTestArchiveWithFiles(t *testing.T, archivePath string, numFiles int) {
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("creating archive: %v", err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	for i := 0; i < numFiles; i++ {
		content := make([]byte, 100)
		header := &tar.Header{
			Name:     filepath.Join("configs", "file"+string(rune(i+'0'))+".yaml"),
			Mode:     0644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("writing header: %v", err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatalf("writing content: %v", err)
		}
	}

	tw.Close()
	gw.Close()
	f.Close()
}

func TestExtractArchiveFileCountLimit(t *testing.T) {
	m := setupTestManager(t)

	tmpDir := t.TempDir()
	evilArchive := filepath.Join(tmpDir, "evil.tar.gz")

	createTestArchiveWithFiles(t, evilArchive, maxExtractFiles+10)

	err := m.extractArchive(evilArchive)
	if err == nil {
		t.Error("expected error for too many files")
	}
}

func TestExtractArchiveSizeLimit(t *testing.T) {
	m := setupTestManager(t)

	tmpDir := t.TempDir()
	evilArchive := filepath.Join(tmpDir, "evil.tar.gz")

	f, err := os.Create(evilArchive)
	if err != nil {
		t.Fatalf("creating archive: %v", err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	largeContent := make([]byte, maxExtractBytes+1000)
	header := &tar.Header{
		Name:     "configs/large.yaml",
		Mode:     0644,
		Size:     int64(len(largeContent)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("writing header: %v", err)
	}
	if _, err := tw.Write(largeContent); err != nil {
		t.Fatalf("writing content: %v", err)
	}

	tw.Close()
	gw.Close()
	f.Close()

	err = m.extractArchive(evilArchive)
	if err == nil {
		t.Error("expected error for size limit exceeded")
	}
}
