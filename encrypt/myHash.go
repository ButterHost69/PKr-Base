package encrypt

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

func GenerateHashWithFilePath(file_path string) (string, error) {
	
  f, err := os.Open(file_path)
  if err != nil {
    return "", fmt.Errorf("could not generate hash of the file.\nError: %e", err)
  }
  defer f.Close()

  h := sha256.New()
  if _, err := io.Copy(h, f); err != nil {
    log.Fatal(err)
  }

  hash := h.Sum(nil)
  return fmt.Sprintf("%x", hash), nil
}

func GenerateHashWithFileIO(file *os.File) (string, error) {	 
  _, err := file.Seek(0,0)
  if err != nil {
    return "", err
  }

  h := sha256.New()
  if _, err := io.Copy(h, file); err != nil {
    log.Fatal(err)
  }

  hash := h.Sum(nil)
  return fmt.Sprintf("%x", hash), nil
}