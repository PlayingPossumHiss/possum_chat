package app_updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Service struct {
	githubClient GithubClient
	mx           *sync.Mutex
}

func New(
	githubClient GithubClient,
) *Service {
	return &Service{
		githubClient: githubClient,
		mx:           &sync.Mutex{},
	}
}

func (s *Service) LatestVersion(ctx context.Context) (*entity.SourceVersion, error) {
	return s.githubClient.LatestVersion(ctx)
}

func (s *Service) Update(ctx context.Context, version entity.SourceVersion) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error on get working directory: %w", err)
	}
	releaseArchivePath := filepath.Join(currentDir, "possum_chat.tar.gz")

	if err := downloadFile(ctx, version.DownloadURL, releaseArchivePath); err != nil {
		return fmt.Errorf("error on save new app version: %w", err)
	}
	defer func() {
		_ = os.Remove(releaseArchivePath) // архив больше не нужен
	}()

	// Вообще в теории мы можем захотеть еще для чего-то использовать tmp
	// да и есть os.MkdirTemp(), но это будет проблемой опоссума из будущего
	tmpReleasePath := filepath.Join(currentDir, "tmp")
	if err := clearTmp(tmpReleasePath); err != nil {
		return fmt.Errorf("error on clear release tmp files: %w", err)
	}

	if err := untargzFile(releaseArchivePath, tmpReleasePath); err != nil {
		return fmt.Errorf("error on untar new app version: %w", err)
	}

	if err := swapRelease(tmpReleasePath, currentDir); err != nil {
		return fmt.Errorf("error on swap version: %w", err)
	}

	if err := clearTmp(tmpReleasePath); err != nil {
		return fmt.Errorf("error on clear release tmp files: %w", err)
	}

	return nil
}

func swapRelease(
	src string,
	dst string,
) error {
	srcContent, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("error on read source folder content: %w", err)
	}

	for _, srcElement := range srcContent {
		to := filepath.Join(dst, srcElement.Name())
		if err := os.RemoveAll(to); err != nil {
			return fmt.Errorf("error on remove old release data: %w", err)
		}
	}

	if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
		return fmt.Errorf("error on copy new release data: %w", err)
	}

	return nil
}

func clearTmp(tmpReleasePath string) error {
	if err := os.RemoveAll(tmpReleasePath); err != nil {
		return fmt.Errorf("error on delete: %w", err)
	}

	return nil
}

func untargzFile(
	archivePath string,
	destination string,
) error {
	err := os.Mkdir(destination, 0o755) //nolint:mnd
	if err != nil {
		return fmt.Errorf("error on create folder for targz: %w", err)
	}

	// Открываем архив
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("error on open targz: %w", err)
	}
	defer archiveFile.Close()

	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return fmt.Errorf("error on read as targz: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	// Проходим по всем файлам в архиве
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("error on read targz: %w", err)
		}

		targetPath := filepath.Join(destination, header.Name)
		// Защита от zip-slip: имя из архива не должно выводить за пределы destination
		relPath, err := filepath.Rel(destination, targetPath)
		if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
			return fmt.Errorf("error on extract path %q: escapes destination", header.Name) //nolint:err113
		}

		if header.Typeflag == tar.TypeDir {
			err = os.MkdirAll(targetPath, 0o755) //nolint:mnd
			if err != nil {
				return fmt.Errorf("error on create folder for targz: %w", err)
			}

			continue
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Сохраняем права из архива (в т.ч. бит запуска у бинаря),
		// иначе после обновления possum_chat перестанет запускаться
		mode := header.FileInfo().Mode().Perm()
		if mode == 0 {
			mode = 0o644
		}
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return fmt.Errorf("error on create file from targz: %w", err)
		}

		_, err = io.Copy(outFile, tarReader)
		if err != nil {
			_ = outFile.Close()

			return fmt.Errorf("error on copy file from targz: %w", err)
		}
		err = outFile.Close()
		if err != nil {
			return fmt.Errorf("error on close file from targz: %w", err)
		}
	}

	return nil
}

func downloadFile(ctx context.Context, url string, filePath string) error {
	client := http.Client{
		Timeout: time.Minute,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error on create download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error on download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("not 200 status on download file: %d", resp.StatusCode) //nolint:err113
	}

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error on create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error on write file: %w", err)
	}

	return nil
}
