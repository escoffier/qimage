package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
)

var ssdbclient *redis.Client

//ImageInfo file infomation
type ImageInfo struct {
	//file name
	Filename string
	//file id
	Md5 string
}

//Server test
type Server struct {
	ServerName string
	ServerIP   string
}

//Serverslice slice
type Serverslice struct {
	Servers []Server
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func getuser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := ps.ByName("uid")
	fmt.Fprintf(w, "you are get user %s", uid)
}

func modifyuser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := ps.ByName("uid")
	fmt.Fprintf(w, "you are modify user %s", uid)
}

func deleteuser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := ps.ByName("uid")
	fmt.Fprintf(w, "you are delete user %s", uid)
}

func adduser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// uid := r.FormValue("uid")
	uid := ps.ByName("uid")
	fmt.Fprintf(w, "you are add user %s", uid)
}

func uploadpost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666) // 此处假设当前目录下已存在test目录
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	//var buff bytes.Buffer
	//fileSize, err := buff.ReadFrom(file)
	buf, err := ioutil.ReadAll(file)

	hash := md5.New()
	//Copy the file in the hash interface and check for any error
	// if _, err := io.Copy(hash, file); err != nil {
	// 	//return returnMD5String, err
	// 	fmt.Printf("copy to hash err: %v", err)
	// }
	hash.Write(buf)
	hashInBytes := hash.Sum(nil)[:16]
	MD5String := hex.EncodeToString(hashInBytes)
	fmt.Println(MD5String)
	ssdbclient.Set(MD5String, buf, 0)
	w.Header().Set("Content-type", "application/json;charset=utf-8")
	fileinfo := ImageInfo{handler.Filename, MD5String}
	b, err := json.Marshal(fileinfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//fmt.Fprintf(w, "%v", handler.Header)
	fmt.Println(string(b))
	w.Write(b)

	if _, err = file.Seek(0, 0); err != nil {
		fmt.Println(err)
	}
	io.Copy(f, file)
}

func uploadget(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	crutime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(crutime, 10))
	token := fmt.Sprintf("%x", h.Sum(nil))

	t, _ := template.ParseFiles("upload.gtpl")
	t.Execute(w, token)
}

func downloadimage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("name")
	fmt.Println(id)

	data, err := ioutil.ReadFile(string("test/11.jpg"))

	if err != nil {
		fmt.Printf("Error with id %s: %v", id, err)
		w.WriteHeader(404)
		w.Write([]byte("404"))
	}

	//w.Header().Set("Content-Type", "application/x-msdownload")
	//w.Header().Set("Content-Type", "image/jpeg")
	//fmt.Println(w.Header()["Content-Type"])
	w.Header().Set("Content-Disposition", "attachment;filename=888.jpg")
	w.Header().Add("Content-Type", "application/x-msdownload")
	fmt.Println(w.Header()["Content-Type"])

	w.Write(data)
	//http.ServeFile(w, r, "test/11.jpg")
}

func main() {
	router := httprouter.New()
	router.GET("/", index)
	router.GET("/hello/:name", hello)
	router.GET("/image/:name", downloadimage)

	router.GET("/user/:uid", getuser)
	router.GET("/upload", uploadget)
	router.POST("/adduser/:uid", adduser)
	router.POST("/upload", uploadpost)
	router.DELETE("/deluser/:uid", deleteuser)
	router.PUT("/moduser/:uid", modifyuser)
	var s Serverslice
	s.Servers = append(s.Servers, Server{ServerName: "Shanghai_VPN", ServerIP: "127.0.0.1"})
	s.Servers = append(s.Servers, Server{ServerName: "Beijing_VPN", ServerIP: "127.0.0.2"})
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Println("json err:", err)
	}
	fmt.Println(string(b))

	var finfo ImageInfo
	finfo = ImageInfo{Filename: "feee", Md5: "33444"}
	f, err := json.Marshal(finfo)
	if err != nil {
		fmt.Println("json err:", err)
	}
	fmt.Println(string(f))

	ssdbclient = redis.NewClient(&redis.Options{
		Addr:     "192.168.21.222:8888",
		Password: "",
		DB:       0,
	})
	pong, err := ssdbclient.Ping().Result()
	fmt.Println(pong, err)

	log.Fatal(http.ListenAndServe(":8080", router))
}
