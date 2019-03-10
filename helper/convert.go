package helper

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
)

// 将 PDF 转成 SVG
func ConvertPDF2SVG(pdfFile, svgFile string, pageNO int) (err error) {
	pdf2svg := strings.TrimSpace(GetConfig("depend", "pdf2svg", "pdf2svg"))

	//Usage: pdf2svg <in file.pdf> <out file.svg> [<page no>]
	args := []string{pdfFile, svgFile, strconv.Itoa(pageNO)}
	cmd := exec.Command(pdf2svg, args...)
	if strings.HasPrefix(pdf2svg, "sudo") {
		args = append([]string{strings.TrimPrefix(pdf2svg, "sudo")}, args...)
		cmd = exec.Command("sudo", args...)
	}
	time.AfterFunc(30*time.Second, func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	})
	if Debug {
		Logger.Debug("PDF 转 SVG :%v", cmd.Args)
	}
	err = cmd.Run()
	return
}

//office文档转pdf，返回转化后的文档路径和错误
func OfficeToPDF(office string) (err error) {
	soffice := strings.TrimSpace(GetConfig("depend", "soffice", "soffice"))
	dirSlice := strings.Split(office, "/")
	dir := strings.Join(dirSlice[0:(len(dirSlice)-1)], "/")
	args := []string{"--headless", "--invisible", "--convert-to", "pdf", office, "--outdir", dir}
	if strings.HasPrefix(soffice, "sudo") {
		args = append([]string{strings.TrimPrefix(soffice, "sudo")}, args...)
		soffice = "sudo"
	}
	cmd := exec.Command(soffice, args...)
	expire := GetConfigInt64("depend", "soffice-expire")
	if expire <= 0 {
		expire = 1800
	}
	time.AfterFunc(time.Duration(expire)*time.Second, func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	})
	if Debug {
		Logger.Debug("office 文档转 PDF:%v", cmd.Args)
	}
	err = cmd.Run()
	return
}

//非office文档(.txt,.mobi,.epub)转pdf文档
func UnOfficeToPDF(file string) (pdfFile string, err error) {
	//calibre := beego.AppConfig.DefaultString("calibre", "ebook-convert")
	calibre := strings.TrimSpace(GetConfig("depend", "calibre", "ebook-convert"))
	pdfFile = filepath.Dir(file) + "/" + strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)) + ".pdf"
	args := []string{
		file,
		pdfFile,
		"--paper-size", "a4",
		"--pdf-default-font-size", "16",
		"--pdf-page-margin-bottom", "36",
		"--pdf-page-margin-left", "36",
		"--pdf-page-margin-right", "36",
		"--pdf-page-margin-top", "36",
	}
	cmd := exec.Command(calibre, args...)
	if strings.HasPrefix(calibre, "sudo") {
		calibre = strings.TrimPrefix(calibre, "sudo")
		args = append([]string{calibre}, args...)
		cmd = exec.Command("sudo", args...)
	}
	if Debug {
		beego.Debug("非Office文档转成PDF：", cmd.Args)
	}
	time.AfterFunc(30*time.Second, func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	})
	err = cmd.Run()
	return
}

//将PDF、SVG文件转成jpg图片格式。注意：如果pdf只有一页，则文件后缀不会出现"-0.jpg"这种情况，否则会出现"-0.jpg,-1.jpg"等
func ConvertToJPEG(file string) (cover string, err error) {
	//convert := beego.AppConfig.DefaultString("imagick", "convert")
	convert := strings.TrimSpace(GetConfig("depend", "imagemagick", "convert"))
	cover = file + ".jpg"
	args := []string{"-density", "150", "-quality", "100", file, cover}
	if strings.HasPrefix(convert, "sudo") {
		args = append([]string{strings.TrimPrefix(convert, "sudo")}, args...)
		convert = "sudo"
	}
	cmd := exec.Command(convert, args...)
	if Debug {
		beego.Debug("转化封面图片：", cmd.Args)
	}
	time.AfterFunc(1*time.Minute, func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	})
	err = cmd.Run()
	return
}

//获取PDF中指定页面的文本内容
//@param			file		PDF文件
//@param			from		起始页
//@param			to			截止页
func ExtractTextFromPDF(file string, from, to int) (content string) {
	pdftotext := strings.TrimSpace(GetConfig("depend", "pdftotext"))
	textfile := file + ".txt"
	defer os.Remove(textfile)
	args := []string{"-f", strconv.Itoa(from), "-l", strconv.Itoa(to), file, textfile}
	if strings.HasPrefix(pdftotext, "sudo") {
		args = append([]string{strings.TrimPrefix(pdftotext, "sudo")}, args...)
		pdftotext = "sudo"
	}
	err := exec.Command(pdftotext, args...).Run()
	if err != nil {
		Logger.Error(err.Error())
		return
	}
	if b, err := ioutil.ReadFile(textfile); err == nil {
		content = string(b)
		content = strings.Replace(content, "\t", " ", -1)
		content = strings.Replace(content, "\n", " ", -1)
		content = strings.Replace(content, "\r", " ", -1)
	} else {
		Logger.Error(err.Error())
	}
	return
}
