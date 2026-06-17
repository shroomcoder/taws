package backup

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Manager struct {
	backupDir string
	configDir string
	stateDir  string
}

type Snapshot struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Files     []string  `json:"files"`
	Size      int64     `json:"size"`
}

type Metadata struct {
	Snapshots []Snapshot `json:"snapshots"`
}

func NewManager(backupDir, configDir, stateDir string) *Manager {
	return &Manager{
		backupDir: backupDir,
		configDir: configDir,
		stateDir:  stateDir,
	}
}

func (m *Manager) CreateSnapshot(version string) (*Snapshot, error) {
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("creating backup directory: %w", err)
	}

	id := time.Now().Format("20060102-150405")
	snapshotDir := filepath.Join(m.backupDir, id)

	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return nil, fmt.Errorf("creating snapshot directory: %w", err)
	}

	var files []string

	configFiles, _ := m.collectFiles(m.configDir)
	files = append(files, configFiles...)

	stateFiles, _ := m.collectFiles(m.stateDir)
	files = append(files, stateFiles...)

	if len(files) == 0 {
		return nil, fmt.Errorf("no files to backup")
	}

	archivePath := filepath.Join(snapshotDir, "backup.tar.gz")
	size, err := m.createArchive(archivePath, files)
	if err != nil {
		return nil, fmt.Errorf("creating archive: %w", err)
	}

	snapshot := &Snapshot{
		ID:        id,
		Timestamp: time.Now(),
		Version:   version,
		Files:     files,
		Size:      size,
	}

	if err := m.saveMetadata(snapshot); err != nil {
		return nil, fmt.Errorf("saving metadata: %w", err)
	}

	return snapshot, nil
}

func (m *Manager) RestoreSnapshot(id string) error {
	snapshot, err := m.GetSnapshot(id)
	if err != nil {
		return fmt.Errorf("getting snapshot: %w", err)
	}

	archivePath := filepath.Join(m.backupDir, id, "backup.tar.gz")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot archive not found: %s", id)
	}

	if err := m.extractArchive(archivePath); err != nil {
		return fmt.Errorf("extracting archive: %w", err)
	}

	_ = snapshot
	return nil
}

func (m *Manager) ListSnapshots() ([]Snapshot, error) {
	metadata, err := m.loadMetadata()
	if err != nil {
		return nil, err
	}

	sort.Slice(metadata.Snapshots, func(i, j int) bool {
		return metadata.Snapshots[i].Timestamp.After(metadata.Snapshots[j].Timestamp)
	})

	return metadata.Snapshots, nil
}

func (m *Manager) GetSnapshot(id string) (*Snapshot, error) {
	metadata, err := m.loadMetadata()
	if err != nil {
		return nil, err
	}

	for i := range metadata.Snapshots {
		if metadata.Snapshots[i].ID == id {
			return &metadata.Snapshots[i], nil
		}
	}

	return nil, fmt.Errorf("snapshot not found: %s", id)
}

func (m *Manager) DeleteSnapshot(id string) error {
	metadata, err := m.loadMetadata()
	if err != nil {
		return err
	}

	for i, s := range metadata.Snapshots {
		if s.ID == id {
			snapshotDir := filepath.Join(m.backupDir, id)
			if err := os.RemoveAll(snapshotDir); err != nil {
				return fmt.Errorf("removing snapshot directory: %w", err)
			}

			metadata.Snapshots = append(metadata.Snapshots[:i], metadata.Snapshots[i+1:]...)
			break
		}
	}

	return m.saveMetadataFile(metadata)
}

func (m *Manager) collectFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(dir, path)
			files = append(files, filepath.Join(dir, relPath))
		}
		return nil
	})

	return files, err
}

func (m *Manager) createArchive(archivePath string, files []string) (int64, error) {
	f, err := os.Create(archivePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	var totalSize int64

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			continue
		}

		relPath, _ := filepath.Rel(filepath.Dir(m.configDir), file)
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			continue
		}

		f, err := os.Open(file)
		if err != nil {
			continue
		}

		written, err := io.Copy(tw, f)
		f.Close()
		if err == nil {
			totalSize += written
		}
	}

	return totalSize, nil
}

const (
	maxExtractFiles = 10000
	maxExtractBytes = 1 << 30 // 1GB
)

func (m *Manager) extractArchive(archivePath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	baseDir, err := filepath.Abs(filepath.Dir(m.configDir))
	if err != nil {
		return fmt.Errorf("resolving base directory: %w", err)
	}

	var fileCount int
	var totalBytes int64

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
			continue
		}

		targetPath := filepath.Join(filepath.Dir(m.configDir), header.Name)

		absTargetPath, err := filepath.Abs(targetPath)
		if err != nil {
			return fmt.Errorf("resolving target path: %w", err)
		}

		if !strings.HasPrefix(absTargetPath, baseDir+string(filepath.Separator)) && absTargetPath != baseDir {
			return fmt.Errorf("path traversal attempt detected: %s", header.Name)
		}

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(absTargetPath, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", absTargetPath, err)
			}
			continue
		}

		dir := filepath.Dir(absTargetPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating parent directory %s: %w", dir, err)
		}

		fileCount++
		if fileCount > maxExtractFiles {
			return fmt.Errorf("archive contains too many files (max %d)", maxExtractFiles)
		}

		outFile, err := os.Create(absTargetPath)
		if err != nil {
			return fmt.Errorf("creating file %s: %w", absTargetPath, err)
		}

		written, err := io.Copy(outFile, tr)
		outFile.Close()
		if err != nil {
			return fmt.Errorf("writing file %s: %w", absTargetPath, err)
		}

		totalBytes += written
		if totalBytes > maxExtractBytes {
			return fmt.Errorf("archive exceeds maximum size (max %d bytes)", maxExtractBytes)
		}
	}

	return nil
}

func (m *Manager) metadataFile() string {
	return filepath.Join(m.backupDir, "metadata.json")
}

func (m *Manager) loadMetadata() (*Metadata, error) {
	data, err := os.ReadFile(m.metadataFile())
	if err != nil {
		if os.IsNotExist(err) {
			return &Metadata{Snapshots: []Snapshot{}}, nil
		}
		return nil, fmt.Errorf("reading metadata: %w", err)
	}

	metadata := &Metadata{}
	if err := json.Unmarshal(data, metadata); err != nil {
		return nil, fmt.Errorf("parsing metadata: %w", err)
	}

	return metadata, nil
}

func (m *Manager) saveMetadata(snapshot *Snapshot) error {
	metadata, err := m.loadMetadata()
	if err != nil {
		return err
	}

	metadata.Snapshots = append(metadata.Snapshots, *snapshot)
	return m.saveMetadataFile(metadata)
}

func (m *Manager) saveMetadataFile(metadata *Metadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}

	if err := os.WriteFile(m.metadataFile(), data, 0600); err != nil {
		return fmt.Errorf("writing metadata: %w", err)
	}

	return nil
}

func FormatSnapshots(snapshots []Snapshot) string {
	var sb strings.Builder

	sb.WriteString("TAWS Backup Snapshots\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	if len(snapshots) == 0 {
		sb.WriteString("  No snapshots found.\n")
		sb.WriteString("  Create one with: taws backup\n")
	} else {
		for _, s := range snapshots {
			sb.WriteString(fmt.Sprintf("  ID:        %s\n", s.ID))
			sb.WriteString(fmt.Sprintf("  Version:   %s\n", s.Version))
			sb.WriteString(fmt.Sprintf("  Timestamp: %s\n", s.Timestamp.Format("2006-01-02 15:04:05")))
			sb.WriteString(fmt.Sprintf("  Files:     %d\n", len(s.Files)))
			sb.WriteString(fmt.Sprintf("  Size:      %d bytes\n\n", s.Size))
		}
	}

	sb.WriteString(strings.Repeat("=", 60) + "\n")

	return sb.String()
}
