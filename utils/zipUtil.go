package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//压缩
//dest 压缩文件存放地址 例如: D:/zip/dest.zip
func Zip(dest string, filePaths ...string) error {
	var files []*os.File
	defer func() {
		for _, file := range files {
			file.Close()
		}
	}()

	for _, filepath := range filePaths {
		file, err := os.Open(filepath)
		if err != nil {
			log.Fatal(fmt.Sprintf("%s:The system cannot find the path specified.", filepath))
		}
		files = append(files, file)
	}

	err := compress(files, dest)

	return err
}

//解压缩
func UnZip(zipFile, dest string) error {
	err := deCompress(zipFile, dest)
	return err
}

func compress(files []*os.File, dest string) error {
	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer d.Close()

	w := zip.NewWriter(d)
	defer w.Close()

	for _, file := range files {
		err := zipCompress(file, "", w)
		if err != nil {
			return err
		}
	}
	return nil
}

func zipCompress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}

	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			if err = zipCompress(f, prefix, zw); err != nil {
				return err
			}
		}
	} else {
		// 获取压缩头信息
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = prefix + "/" + header.Name
		// 指定文件压缩方式 默认为 Store 方式 该方式不压缩文件 只是转换为zip保存
		header.Method = zip.Deflate
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		// 写入文件到压缩包中
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func deCompress(zipFile, dest string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close() //注意：可能会用尽文件描述符
		//filename := dest + file.Name
		filename := filepath.Join(dest, file.Name)
		err = os.MkdirAll(getDir(filename), 0755)
		if err != nil {
			return err
		}
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close() //注意：可能会用尽文件描述符
		_, err = io.Copy(w, rc)
		if err != nil {
			return err
		}
		w.Close()
		rc.Close()
	}
	return nil
}

func getDir(path string) string {
	return subString(path, 0, strings.LastIndex(path, string(filepath.Separator)))
}
func subString(str string, start, end int) string {
	rs := []rune(str)
	length := len(rs)
	if start < 0 || start > length {
		panic("start is wrong")
	}
	if end < start || end > length {
		panic("end is wrong path:" + str)
	}
	return string(rs[start:end])
}

// nolint
func ExtractZipFile(src *zip.File, fn string) (err error) {
	var (
		f *os.File
		r io.ReadCloser
	)

	if f, err = os.Create(fn); err != nil {
		return
	}
	defer f.Close()

	if r, err = src.Open(); err != nil {
		return
	}
	defer r.Close()

	_, err = io.Copy(f, r)

	return
}

// nolint
func ExtractZipPackage(r io.ReaderAt, sz int64, dirname string) (err error) {
	var (
		zfs *zip.Reader
		zf  *zip.File
		fn  string
	)

	if zfs, err = zip.NewReader(r, sz); err != nil {
		return
	}

	for _, zf = range zfs.File {
		fn = filepath.Join(dirname, zf.Name)
		if err = os.MkdirAll(getDir(fn), os.ModePerm); err != nil {
			return
		}
		if err = ExtractZipFile(zf, fn); err != nil {
			return
		}
	}

	return
}
