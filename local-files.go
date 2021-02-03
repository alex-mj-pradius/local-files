package localfiles

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Unzip -is unzipping src (sourceFileName) to dest (destinationFileName)
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				log.Fatal(err)
				return err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ZipFiles - is putting few files (of []fileNames) in one zip
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

// addFileToZip
func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	//header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

//MoveFile - move local src (fileName) to dest (fileName)
// os.Rename or (io.Copy & os.Remove)
func MoveFile(src, dst string) error {
	//если src и dst на одном диске
	if filepath.VolumeName(src) == filepath.VolumeName(dst) {
		//то используем os.Rename
		err := os.Rename(src, dst)
		if err != nil {
			return err
		}
		return nil
	}
	//проверяем доступ к файлу
	f, err := os.OpenFile(src, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	//если на разных дисках
	//копируем
	err = CopyFile(src, dst)
	if err != nil {
		return err
	}
	//удаляем
	err = os.Remove(src)
	if err != nil {
		if os.IsNotExist(err) {
			//log.WithFields(logrus.Fields{"src": src, "dst": dst, "err": err}).Warn("Файл перемещён, но удалить его не получилось: удалён кем-то ещё.")
			fmt.Println("Файл перемещён, но удалить его не получилось: удалён кем-то ещё.")
			return nil
		}
		return fmt.Errorf("Файл перемещён. Но удалить его не получилось. %v", err)
	}
	return nil
}

//CopyFile - copy local src (fileName) to dest (fileName) (io.Copy)
func CopyFile(src string, dst string) (err error) {

	sourcefile, err := os.Open(src)
	fmt.Println(src + " --->>> " + dst)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer sourcefile.Close()

	destfile, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//копируем содержимое и проверяем коды ошибок
	_, err = io.Copy(destfile, sourcefile)
	if closeErr := destfile.Close(); err == nil {
		//если ошибки в io.Copy нет, то берем ошибку от destfile.Close(), если она была
		err = closeErr
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	sourceinfo, err := os.Stat(src)
	if err == nil {
		err = os.Chmod(dst, sourceinfo.Mode())
	}
	return err
}

//DeleteFile - delete local src (fileName) to dest (fileName) (io.Copy)
func DeleteFile(src string) error {

	//проверяем доступ к файлу
	f, err := os.OpenFile(src, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	//удаляем
	err = os.Remove(src)
	if err != nil {
		if os.IsNotExist(err) {
			//log.WithFields(logrus.Fields{"src": src, "dst": dst, "err": err}).Warn("Файл перемещён, но удалить его не получилось: удалён кем-то ещё.")
			fmt.Printf("Файлa %s не существует, (неправильное имя или удалён кем-то ещё).", src)
			return nil
		}
		return fmt.Errorf("%s -  удалить не получилось. %v", src, err)
	}
	return nil
}
