package filedownloader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/filepathgetter"
)

const (
	filepathURL    = "https://api.telegram.org/file/bot%s/%s"
	tmpFilePattern = "tmp_download_%s_%d"
)

type FileDownloader interface {
	DownloadFile(filepath string) (string, error)
}

type fileDownloader struct {
	client   *http.Client
	token    string
	dir2save string

	filepathGetter filepathgetter.FilepathGetter
}

func NewFileDownloader(client *http.Client, token string, dir2save string) (FileDownloader, error) {
	f := &fileDownloader{
		client:   client,
		token:    token,
		dir2save: dir2save,
	}
	if err := f.init(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *fileDownloader) init() error {
	if _, err := os.Stat(f.dir2save); os.IsNotExist(err) {
		if err := os.MkdirAll(f.dir2save, 0755); err != nil {
			return fmt.Errorf("failed to mkdir %s, %w", f.dir2save, err)
		}
	}
	f.filepathGetter = filepathgetter.NewFilepathGetter(f.client, f.token)
	return nil
}

func (f fileDownloader) DownloadFile(fileID string) (string, error) {
	filepath, err := f.filepathGetter.GetFilepath(fileID)
	if err != nil {
		return "", err
	}

	file, err := f.getFile(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return f.calcHashAndSaveFile(file, fileID)
}

func (f fileDownloader) getFile(filepath string) (io.ReadCloser, error) {
	downloadURL := fmt.Sprintf(filepathURL, f.token, filepath)
	resp, err := f.client.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failet to get file: %w", err)
	}

	if resp.StatusCode >= 300 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download status %d: %s", resp.StatusCode, string(b))
	}
	return resp.Body, nil
}

func (f fileDownloader) calcHashAndSaveFile(src io.ReadCloser, fileID string) (string, error) {
	file, err := os.CreateTemp("", fmt.Sprintf(tmpFilePattern, fileID, time.Now().Unix()))
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	mw := io.MultiWriter(file, hash)

	if _, err := io.Copy(mw, src); err != nil {
		return "", err
	}
	sum := hex.EncodeToString(hash.Sum(nil))

	final := filepath.Join(f.dir2save, sum+".xlsx")
	if err := os.Rename(file.Name(), final); err != nil {
		// если уже есть файл с таким hash — удаляем tmp и считаем ок
		_ = os.Remove(file.Name())
		return sum, nil
	}
	return sum, nil
}
