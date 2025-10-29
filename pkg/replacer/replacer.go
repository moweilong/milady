// Package replacer is a library of replacement file content, supports replacement of
// files in local directories and embedded directory files via embed.
package replacer

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/moweilong/milady/pkg/gofile"
)

var _ Replacer = (*replacerInfo)(nil)

// Replacer interface. 一个文本替换器接口，用于替换文件中的文本。
type Replacer interface {
	SetReplacementFields(fields []Field)
	// SetSubDirsAndFiles if subDirs or subFiles not empty, then set r.files to all files in subDirs and subFiles.
	SetSubDirsAndFiles(subDirs []string, subFiles ...string)
	SetIgnoreSubDirs(dirs ...string)
	SetIgnoreSubFiles(filenames ...string)
	// SetOutputDir set output directory.
	// if absPath is not empty, return absolute path of absPath.
	// if absPath is empty, return an automatically generated path with name and timestamp in the current directory.
	// e.g: /home/milady/mycode/model_150405
	SetOutputDir(absDir string, name ...string) error
	GetOutputDir() string
	GetSourcePath() string
	// SaveFiles save file with setting
	SaveFiles() error
	ReadFile(filename string) ([]byte, error)
	GetFiles() []string
	SaveTemplateFiles(m map[string]interface{}, parentDir ...string) error
}

// replacerInfo replacer information
type replacerInfo struct {
	// template directory or file
	path string
	// Template directory corresponding to binary objects
	fs embed.FS
	// true: use os to manipulate files, false: use embed.FS to manipulate files
	// New() default is true, NewFS() default is false
	isActual bool
	// list of template files
	files []string
	// ignore the list of replaced files, default is "", e.g. ignore.txt or myDir/ignore.txt
	ignoreFiles []string
	// ignore processed subdirectories, default is ""
	ignoreDirs []string
	// characters to be replaced when converting from a template file to a new file, default is []Field{}
	replacementFields []Field
	// the directory where the file is saved after replacement, default is ""
	outPath string
}

// New create replacer with local directory
// path e.g: /home/milady/.milady
func New(path string) (Replacer, error) {
	files, err := gofile.ListFiles(path)
	if err != nil {
		return nil, err
	}

	path, _ = filepath.Abs(path)
	return &replacerInfo{
		path:              path,
		isActual:          true,
		files:             files,
		replacementFields: []Field{},
	}, nil
}

// NewFS create replacer with embed.FS
// path e.g: /home/milady/.milady
func NewFS(path string, fs embed.FS) (Replacer, error) {
	files, err := listFiles(path, fs)
	if err != nil {
		return nil, err
	}

	return &replacerInfo{
		path:              path,
		fs:                fs,
		isActual:          false,
		files:             files,
		replacementFields: []Field{},
	}, nil
}

// Field replace field information
type Field struct {
	Old             string // old field
	New             string // new field
	IsCaseSensitive bool   // whether the first letter is case-sensitive
}

// SetReplacementFields set the replacement field, note: old characters should not be included in the relationship,
// if they exist, pay attention to the order of precedence when setting the Field
func (r *replacerInfo) SetReplacementFields(fields []Field) {
	var newFields []Field
	for _, field := range fields {
		if field.IsCaseSensitive && isFirstAlphabet(field.Old) { // splitting the initial case field
			if field.New == "" {
				continue
			}
			newFields = append(newFields,
				Field{ // convert the first letter to upper case
					Old: strings.ToUpper(field.Old[:1]) + field.Old[1:],
					New: strings.ToUpper(field.New[:1]) + field.New[1:],
				},
				Field{ // convert the first letter to lower case
					Old: strings.ToLower(field.Old[:1]) + field.Old[1:],
					New: strings.ToLower(field.New[:1]) + field.New[1:],
				},
			)
		} else {
			newFields = append(newFields, field)
		}
	}
	r.replacementFields = newFields
}

// GetFiles get files
func (r *replacerInfo) GetFiles() []string {
	return r.files
}

// SetSubDirsAndFiles if subDirs or subFiles not empty, then set r.files to all files in subDirs and subFiles.
func (r *replacerInfo) SetSubDirsAndFiles(subDirs []string, subFiles ...string) {
	subDirs = r.convertPathsDelimiter(subDirs...)
	subFiles = r.convertPathsDelimiter(subFiles...)

	var files []string
	isExistFile := make(map[string]struct{}) // use map to avoid duplicate files
	for _, file := range r.files {           // r.files all files of milady
		for _, dir := range subDirs {
			if isSubPath(file, dir) {
				if _, ok := isExistFile[file]; ok {
					continue
				}
				isExistFile[file] = struct{}{}
				files = append(files, file)
			}
		}
		for _, sf := range subFiles {
			if isMatchFile(file, sf) {
				if _, ok := isExistFile[file]; ok {
					continue
				}
				isExistFile[file] = struct{}{}
				files = append(files, file)
			}
		}
	}

	if len(files) == 0 {
		return
	}
	r.files = files
}

// SetIgnoreSubFiles specify files to be ignored
func (r *replacerInfo) SetIgnoreSubFiles(filenames ...string) {
	r.ignoreFiles = append(r.ignoreFiles, filenames...)
}

// SetIgnoreSubDirs specify subdirectories to be ignored
func (r *replacerInfo) SetIgnoreSubDirs(dirs ...string) {
	dirs = r.convertPathsDelimiter(dirs...)
	r.ignoreDirs = append(r.ignoreDirs, dirs...)
}

// SetOutputDir set output directory.
// if absPath is not empty, return absolute path of absPath.
// if absPath is empty, return an automatically generated path with name and timestamp in the current directory.
// e.g: /home/milady/mycode/model_150405
func (r *replacerInfo) SetOutputDir(absPath string, name ...string) error {
	// output e.g: /home/milady/mycode/model
	if absPath != "" {
		abs, err := filepath.Abs(absPath)
		if err != nil {
			return err
		}

		r.outPath = abs
		return nil
	}

	subPath := strings.Join(name, "_")
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	// output e.g: /home/milady/mycode/model_150405
	r.outPath = pwd + gofile.GetPathDelimiter() + subPath + "_" + time.Now().Format("150405")
	return nil
}

// GetOutputDir get output directory
func (r *replacerInfo) GetOutputDir() string {
	return r.outPath
}

// GetSourcePath get source directory
func (r *replacerInfo) GetSourcePath() string {
	return r.path
}

// ReadFile read file content
func (r *replacerInfo) ReadFile(filename string) ([]byte, error) {
	filename = r.convertPathDelimiter(filename)

	foundFile := []string{}
	for _, file := range r.files {
		if strings.Contains(file, filename) && gofile.GetFilename(file) == gofile.GetFilename(filename) {
			foundFile = append(foundFile, file)
		}
	}
	if len(foundFile) != 1 {
		return nil, fmt.Errorf("total %d file named '%s', files=%+v", len(foundFile), filename, foundFile)
	}

	if r.isActual {
		return os.ReadFile(foundFile[0])
	}
	return r.fs.ReadFile(foundFile[0])
}

// SaveFiles save file with setting
func (r *replacerInfo) SaveFiles() error {
	// TODO delete this line
	if r.outPath == "" {
		r.outPath = gofile.GetRunPath() + gofile.GetPathDelimiter() + "generate_" + time.Now().Format("150405")
	}

	var existFiles []string
	var writeData = make(map[string][]byte)

	// process replacer files
	for _, file := range r.files {
		// skip ignore files or dirs
		if r.isInIgnoreDir(file) || r.isIgnoreFile(file) {
			continue
		}

		var data []byte
		var err error

		// read file content
		if r.isActual {
			data, err = os.ReadFile(file) // read from local files
		} else {
			data, err = r.fs.ReadFile(file) // read from local embed.FS
		}
		if err != nil {
			return err
		}

		// replace text content
		for _, field := range r.replacementFields {
			data = bytes.ReplaceAll(data, []byte(field.Old), []byte(field.New))
		}

		// file name splicing with outPath
		newFilePath := r.getNewFilePath(file)
		dir, filename := filepath.Split(newFilePath)

		// replace dir and filename with replacement rules
		for _, field := range r.replacementFields {
			if strings.Contains(dir, field.Old) {
				dir = strings.ReplaceAll(dir, field.Old, field.New)
			}
			if strings.Contains(filename, field.Old) {
				filename = strings.ReplaceAll(filename, field.Old, field.New)
			}

			if newFilePath != dir+filename {
				newFilePath = dir + filename
			}
		}

		// check if the file already exists
		if gofile.IsExists(newFilePath) {
			existFiles = append(existFiles, newFilePath)
		}
		// map of write file content with new file path
		writeData[newFilePath] = data
	}

	// break if outPath have existing files
	if len(existFiles) > 0 {
		return fmt.Errorf("existing files detected\n    %s\nCode generation has been cancelled\n",
			strings.Join(existFiles, "\n    "))
	}

	// break if generate file is in r.path
	for file, data := range writeData {
		if isForbiddenFile(file, r.path) {
			return fmt.Errorf("disable writing file(%s) to directory(%s), file size=%d", file, r.path, len(data))
		}
	}

	// save files to file system
	for file, data := range writeData {
		err := saveToNewFile(file, data)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveTemplateFiles save file with setting
func (r *replacerInfo) SaveTemplateFiles(m map[string]interface{}, parentDir ...string) error {
	refDir := ""
	if len(parentDir) > 0 {
		refDir = strings.Join(parentDir, gofile.GetPathDelimiter())
	}

	writeData := make(map[string][]byte, len(r.files))
	for _, file := range r.files {
		data, err := replaceTemplateData(file, m)
		if err != nil {
			return err
		}
		newFilePath := r.getNewFilePath2(file, refDir)
		newFilePath = trimExt(newFilePath)
		if gofile.IsExists(newFilePath) {
			return fmt.Errorf("file %s already exists, cancel code generation", newFilePath)
		}
		newFilePath, err = replaceTemplateFilePath(newFilePath, m)
		if err != nil {
			return err
		}
		writeData[newFilePath] = data
	}

	for file, data := range writeData {
		err := saveToNewFile(file, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func replaceTemplateData(file string, m map[string]interface{}) ([]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file failed, err=%s", err)
	}
	if !bytes.Contains(data, []byte("{{")) {
		return data, nil
	}

	builder := bytes.Buffer{}
	tmpl, err := template.New(file).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse data failed, err=%s", err)
	}
	err = tmpl.Execute(&builder, m)
	if err != nil {
		return nil, fmt.Errorf("execute data failed, err=%s", err)
	}
	return builder.Bytes(), nil
}

func replaceTemplateFilePath(file string, m map[string]interface{}) (string, error) {
	if !strings.Contains(file, "{{") {
		return file, nil
	}

	builder := strings.Builder{}
	tmpl, err := template.New("file: " + file).Parse(file)
	if err != nil {
		return file, fmt.Errorf("parse file failed, err=%s", err)
	}
	err = tmpl.Execute(&builder, m)
	if err != nil {
		return file, fmt.Errorf("execute file failed, err=%s", err)
	}
	return builder.String(), nil
}

func trimExt(file string) string {
	file = strings.TrimSuffix(file, ".tmpl")
	file = strings.TrimSuffix(file, ".tpl")
	file = strings.TrimSuffix(file, ".template")
	return file
}

func (r *replacerInfo) isIgnoreFile(file string) bool {
	isIgnore := false
	for _, v := range r.ignoreFiles {
		if isMatchFile(file, v) {
			isIgnore = true
			break
		}
	}
	return isIgnore
}

func (r *replacerInfo) isInIgnoreDir(file string) bool {
	isIgnore := false
	dir, _ := filepath.Split(file)
	for _, v := range r.ignoreDirs {
		if strings.Contains(dir, v) {
			isIgnore = true
			break
		}
	}
	return isIgnore
}

// isForbiddenFile 判断文件是否在指定路径下
//  1. 如果是windows系统，批量转换路径分隔符
//  2. 判断 file 是否包含 path
func isForbiddenFile(file string, path string) bool {
	if gofile.IsWindows() {
		path = strings.ReplaceAll(path, "/", "\\")
		file = strings.ReplaceAll(file, "/", "\\")
	}
	return strings.Contains(file, path)
}

// getNewFilePath delete the path prefix r.path of the file and add the outPath prefix
// e.g: /home/murphy/workspace/golang/src/github.com/moweilong/milady/dao_172320 + /internal/apiserver/model/aa.go
func (r *replacerInfo) getNewFilePath(file string) string {
	newFilePath := r.outPath + strings.Replace(file, r.path, "", 1)

	if gofile.IsWindows() {
		newFilePath = strings.ReplaceAll(newFilePath, "/", "\\")
	}

	return newFilePath
}

func (r *replacerInfo) getNewFilePath2(file string, refDir string) string {
	if refDir == "" {
		return r.getNewFilePath(file)
	}

	newFilePath := r.outPath + gofile.GetPathDelimiter() + refDir + gofile.GetPathDelimiter() + strings.Replace(file, r.path, "", 1)
	if gofile.IsWindows() {
		newFilePath = strings.ReplaceAll(newFilePath, "/", "\\")
	}
	return newFilePath
}

// if windows, convert the path splitter
func (r *replacerInfo) convertPathDelimiter(filePath string) string {
	if r.isActual && gofile.IsWindows() {
		filePath = strings.ReplaceAll(filePath, "/", "\\")
	}
	return filePath
}

// if windows, batch convert path splitters
// 如果是windows系统，批量转换路径分隔符
func (r *replacerInfo) convertPathsDelimiter(filePaths ...string) []string {
	if r.isActual && gofile.IsWindows() {
		filePathsTmp := []string{}
		for _, dir := range filePaths {
			filePathsTmp = append(filePathsTmp, strings.ReplaceAll(dir, "/", "\\"))
		}
		return filePathsTmp
	}
	return filePaths
}

// saveToNewFile save data to filePath
//  1. create directory if not exists
//  2. save file
func saveToNewFile(filePath string, data []byte) error {
	// create directory
	dir, _ := filepath.Split(filePath)
	err := os.MkdirAll(dir, 0766)
	if err != nil {
		return err
	}

	// save file
	err = os.WriteFile(filePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

// iterates over all files in the embedded directory, returning the absolute path to the file
func listFiles(path string, fs embed.FS) ([]string, error) {
	files := []string{}
	err := walkDir(path, &files, fs)
	return files, err
}

// iterating through the embedded catalog
func walkDir(dirPath string, allFiles *[]string, fs embed.FS) error {
	files, err := fs.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		deepFile := dirPath + "/" + file.Name()
		if file.IsDir() {
			_ = walkDir(deepFile, allFiles, fs)
			continue
		}
		*allFiles = append(*allFiles, deepFile)
	}

	return nil
}

// determine if the first character of a string is a letter
func isFirstAlphabet(str string) bool {
	if len(str) == 0 {
		return false
	}

	if (str[0] >= 'A' && str[0] <= 'Z') || (str[0] >= 'a' && str[0] <= 'z') {
		return true
	}

	return false
}

// isSubPath 比对 filePath 是否包含 subPath
//  1. 从 filePath 中提取目录部分
//  2. 检查 subPath 是否是 dir 的子路径
func isSubPath(filePath string, subPath string) bool {
	dir, _ := filepath.Split(filePath)
	return strings.Contains(dir, subPath)
}

// isMatchFile 比对 filePath 和 sf 是否匹配
//  1. 文件名不匹配，直接返回false
//  2. 根据操作系统，批量转换路径分隔符
//  3. 比较 dir1 是否包含 dir2, 如果包含，返回true
func isMatchFile(filePath string, sf string) bool {
	dir1, file1 := filepath.Split(filePath)
	dir2, file2 := filepath.Split(sf)

	if file1 != file2 {
		return false
	}

	if gofile.IsWindows() {
		dir1 = strings.ReplaceAll(dir1, "/", "\\")
		dir2 = strings.ReplaceAll(dir2, "/", "\\")
	} else {
		dir1 = strings.ReplaceAll(dir1, "\\", "/")
		dir2 = strings.ReplaceAll(dir2, "\\", "/")
	}

	// 比较 dir1 是否包含 dir2, 如果包含，返回true
	l1, l2 := len(dir1), len(dir2)
	if l1 >= l2 && dir1[l1-l2:] == dir2 {
		return true
	}
	return false
}
