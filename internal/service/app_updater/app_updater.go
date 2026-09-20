package app_updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"archive/tar"
	"compress/gzip"

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
		return fmt.Errorf("erro on get working directory: %w", err)
	}
	releaseArchivePath := fmt.Sprintf("%s/possum_chat.tar.gz", currentDir)

	err = downloadFile(
		version.DownloadURL,
		releaseArchivePath,
	)
	if err != nil {
		return fmt.Errorf("error on save new app version: %w", err)
	}

	// Вообще в теории мы можем захотеть еще для чего-то использовать tmp
	// да и есть os.MkdirTemp(), но это будет проблемой опоссума из будущего
	tmpReleasePath := fmt.Sprintf("%s/tmp", currentDir)
	err = clearTmp(tmpReleasePath)
	if err != nil {
		return fmt.Errorf("error cleart release tmp files: %w", err)
	}

	err = untargzFile(releaseArchivePath, tmpReleasePath)
	if err != nil {
		return fmt.Errorf("error on untar new app version: %w", err)
	}

	err = swapRelease(
		tmpReleasePath,
		currentDir,
	)
	if err != nil {
		return fmt.Errorf("error on swap version: %w", err)
	}

	err = clearTmp(tmpReleasePath)
	if err != nil {
		return fmt.Errorf("error cleart release tmp files: %w", err)
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

	for _, srcElment := range srcContent {
		to := fmt.Sprintf("%s/%s", dst, srcElment.Name())
		err = os.RemoveAll(to)
		if err != nil {
			return fmt.Errorf("error on remove old release data: %w", err)
		}
	}

	err = os.CopyFS(dst, os.DirFS(src))
	if err != nil {
		return fmt.Errorf("error on copy new release data: %w", err)
	}

	return nil
}

func clearTmp(tmpReleasePath string) error {
	err := os.RemoveAll(tmpReleasePath)
	if err != nil {
		return fmt.Errorf("error ondelete: %w", err)
	}

	return nil
}

func untargzFile(
	archivePath string,
	destination string,
) error {
	err := os.Mkdir(destination, 0755)
	if err != nil {
		return fmt.Errorf("error on create folder for targz: %w", err)
	}

	// Открываем архив
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("error on open targz: %w", err)
	}
	defer f.Close()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("error on read as targz: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	// Проходим по всем файлам в архиве
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error on read targz: %w", err)
		}

		targetPath := filepath.Join(destination, header.Name)

		if header.Typeflag == tar.TypeDir {
			err = os.MkdirAll(targetPath, 0755)
			if err != nil {
				return fmt.Errorf("error on create folder for targz: %w", err)
			}
			continue
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		outFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("error on create file from targz: %w", err)
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, tarReader)
		if err != nil {
			return fmt.Errorf("error on copy file from targz: %w", err)
		}
	}

	return nil
}

func downloadFile(url string, filepath string) error {
	client := http.Client{
		Timeout: time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error on download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("not 200 status on download file: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
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
