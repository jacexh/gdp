package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	TemplateURL    = "https://github.com/jacexh/golang-ddd-template/archive/master.zip"
	UnzipDirectory = "golang-ddd-template-master"
)

type (
	// Project 项目信息
	Project struct {
		BinFile                    string // 编译的二进制文件名称，比如 app
		Module                     string // 模块名称，比如 github.com/jacexh/golang-ddd-template
		EnvironmentVariablesPrefix string // 项目环境变量的前缀名称，比如APP
	}
)

func DownloadTemplate(template string) (string, error) {
	resp, err := http.Get(template)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fn := strconv.Itoa(int(time.Now().Unix())) + ".zip"
	fn = os.TempDir() + fn

	f, err := os.Create(fn)
	if err != nil {
		return "", err
	}

	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}

	return fn, nil
}

func Unzip(file string) error {
	reader, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(f.Name, os.ModePerm)
			continue
		}

		f1, err := f.Open()
		if err != nil {
			return err
		}
		err = os.MkdirAll(filepath.Dir(f.Name), os.ModePerm)
		if err != nil {
			return err
		}
		f2, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(f2, f1)
		if err != nil {
			return err
		}
	}
	return nil
}

func ProjectSetting() *Project {
	project := new(Project)
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("输入项目Module（如 github.com/jacexh/golang-ddd-template）：")
	text, _ := reader.ReadString('\n')
	project.Module = strings.Trim(text, "\n")
	if project.Module == "" {
		panic("module不能为空")
	}

	fmt.Print("输入项目编译二进制文件名称（如 blog）：")
	text, _ = reader.ReadString('\n')
	project.BinFile = strings.Trim(text, "\n")
	if project.BinFile == "" {
		panic("二进制文件名不能为空")
	}

	fmt.Print("输入项目环境变量前缀（如APP）：")
	text, _ = reader.ReadString('\n')
	project.EnvironmentVariablesPrefix = strings.ToUpper(strings.Trim(text, "\n"))
	if project.EnvironmentVariablesPrefix == "" {
		panic("环境变量前缀不能为空")
	}

	fmt.Printf("请确认配置：%v\n", project)
	return project
}

func RenderTemplateProject(dst string, proj *Project) error {
	return filepath.Walk(dst, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		tpl := template.New(path)
		tpl, err = tpl.Parse(string(data))
		if err != nil {
			panic(err)
			return err
		}
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			panic(err)
			return err
		}
		return tpl.Execute(file, proj)
	})
}

func main() {
	proj := ProjectSetting()
	f, err := DownloadTemplate(TemplateURL)
	if err != nil {
		panic(err)
	}

	err = Unzip(f)
	if err != nil {
		panic(err)
	}
	err = RenderTemplateProject(UnzipDirectory, proj)
	if err != nil {
		panic(err)
	}

	ds := strings.Split(proj.Module, "/")
	os.Rename(UnzipDirectory, ds[len(ds)-1])
}
