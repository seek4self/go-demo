package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

const tmpdir = "./tmp"

var lock sync.WaitGroup

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("./www/*")
	router.Static("/", "./www")
	router.POST("chunkfile", postUpload)
	router.Run(":8080")
}

func postUpload(c *gin.Context) {
	info := chunkInfo{}
	if err := c.ShouldBind(&info); err != nil {
		fmt.Println("parse form data err", err)
		c.String(400, "parse form data err")
		return
	}
	if err := os.MkdirAll(tmpdir, 0666); err != nil {
		fmt.Println("mkdir err", err, tmpdir)
		c.String(500, "mkdir err")
		return
	}
	filePath := info.chunkName(info.Index)
	if err := c.SaveUploadedFile(info.File, filePath); err != nil {
		fmt.Println("save file err", err, filePath)
		c.String(500, "save file err")
		return
	}

	if info.isFinish() {
		// 新文件创建
		newfile := info.newFile()
		chunksize := info.chunkSize()
		// 读取文件片段 进行合并
		for i := 0; i < info.Total; i++ {
			lock.Add(1)
			go mergeFile(newfile, info.chunkName(i), chunksize*int64(i))
		}
		lock.Wait()
	}
	c.String(200, "ok")
}

type chunkInfo struct {
	File  *multipart.FileHeader `form:"file"`
	Index int                   `form:"index"`
	Total int                   `form:"total"`
	Size  int64                 `form:"size"`
}

func (ci chunkInfo) chunkName(index int) string {
	return fmt.Sprintf("%s/%s_%d", tmpdir, ci.File.Filename, index)
}

// 判断是否完成  根据现有文件的大小 与 上传文件大小进行匹配
func (ci chunkInfo) isFinish() bool {
	totalSize := int64(0)
	for i := 0; i < ci.Total; i++ {
		// 分片大小获取
		fi, err := os.Stat(ci.chunkName(i))
		if err == nil {
			totalSize += fi.Size()
		}
	}
	return totalSize == ci.Size
}

func (ci chunkInfo) newFile() string {
	filePath := fmt.Sprintf("./%s", ci.File.Filename)
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			f, _ := os.Create(filePath)
			f.Close()
		}
	}
	return filePath
}

func (ci chunkInfo) chunkSize() int64 {
	fi, _ := os.Stat(ci.chunkName(0))
	return fi.Size()
}

// 合并切片文件
func mergeFile(newFile, chunkFile string, offset int64) {
	defer lock.Done()
	dst, err := os.OpenFile(newFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("open newfile err", err)
	}
	defer dst.Close()
	// 设置文件写入偏移量
	dst.Seek(offset, 0)
	src, err := os.Open(chunkFile)
	if err != nil {
		log.Fatal("open chunkFile err", err)
	}
	fmt.Println("merge file:", chunkFile)
	_, err = io.Copy(dst, src)
	if err != nil {
		fmt.Println("copy file err", err)
	}
	src.Close()
	// fmt.Println("delete file", chunkFile)
	os.Remove(chunkFile)
}
