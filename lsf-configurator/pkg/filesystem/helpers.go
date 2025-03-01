package filesystem

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func SaveMultiPartFile(fileHeader *multipart.FileHeader, saveLocation string) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	savePath := filepath.Join(saveLocation, fileHeader.Filename)
	outFile, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return err
	}

	return nil
}

func CopyFileToDstFolder(source string, dstFolder string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	_, fileName := filepath.Split(source)
	destPath := filepath.Join(dstFolder, fileName)

	dstFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	return nil
}

func CreateDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func DeleteDir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to delete folder %s: %w", path, err)
	}
	return nil
}
