package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
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

func savaImage(buf []byte, MD5String string) {

	//buf, err := ioutil.ReadAll(file)

	//hash := md5.New()
	//Copy the file in the hash interface and check for any error
	// if _, err := io.Copy(hash, file); err != nil {
	// 	//return returnMD5String, err
	// 	fmt.Printf("copy to hash err: %v", err)
	// }
	//hash.Write(buf)
	//hashInBytes := hash.Sum(nil)[:16]
	//MD5String := hex.EncodeToString(hashInBytes)
	//fmt.Println(MD5String)

	//leveldir := md5[:6]
	//level2dir := md5[3:6]
	//hex.Encode()
	//str := hex.EncodeToString(leveldir)

	//buf := bytes.NewReader(level1dir)
	l1int, err := strconv.ParseUint(MD5String[0:3], 16, 32)
	if err != nil {
		fmt.Println(err)
	}
	l1int = l1int / 4
	le1str := strconv.FormatUint(l1int, 10)

	l2int, err := strconv.ParseUint(MD5String[3:6], 16, 32)
	if err != nil {
		fmt.Println(err)
	}
	l2int = l2int / 4
	l2str := strconv.FormatUint(l2int, 10)

	saveName := "image/" + le1str + "/" + l2str + "/" + MD5String + "/"
	fmt.Println(saveName)

	os.MkdirAll(saveName, 0777)

	f, err := os.OpenFile(saveName+"0_0", os.O_WRONLY|os.O_CREATE, 0666) // 此处假设当前目录下已存在test目录
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	//var buff bytes.Buffer
	//fileSize, err := buff.ReadFrom(file)
	ssdbclient.Set(MD5String, buf, 0)
	f.Write(buf)
	//io.w(f, file)

}
func uploadpost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666) // 此处假设当前目录下已存在test目录
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// defer f.Close()
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
	savaImage(buf, MD5String)

	//ssdbclient.Set(MD5String, buf, 0)
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

	//if _, err = file.Seek(0, 0); err != nil {
	//	fmt.Println(err)
	//}
	//io.Copy(f, file)
}

func uploadget(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	crutime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(crutime, 10))
	token := fmt.Sprintf("%x", h.Sum(nil))

	t, _ := template.ParseFiles("upload.gtpl")
	t.Execute(w, token)
}

func getImagePath(MD5String string) (string, error) {

	if len(MD5String) != 32 {
		err := errors.New("Invalid md5: " + MD5String)
		return string(""), err
	}
	l1int, err := strconv.ParseUint(MD5String[0:3], 16, 32)
	if err != nil {
		fmt.Println(err)
		return string(""), err
	}
	l1int = l1int / 4
	le1str := strconv.FormatUint(l1int, 10)

	l2int, err := strconv.ParseUint(MD5String[3:6], 16, 32)
	if err != nil {
		fmt.Println(err)
		return string(""), err
	}
	l2int = l2int / 4
	l2str := strconv.FormatUint(l2int, 10)

	dirpath := "image/" + le1str + "/" + l2str + "/" + MD5String + "/0_0"
	return dirpath, nil
}
func downloadimage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	md5str := ps.ByName("name")
	fmt.Println(md5str)

	p := r.URL.Query().Get("p")
	fmt.Println("p=" + p)

	path, err := getImagePath(md5str)
	if err != nil {
		fmt.Printf("get Image path failed: %v", err)
		w.WriteHeader(404)
		w.Write([]byte("404"))
		return
	}
	fmt.Println("Image path: " + path)
	data, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Printf("Error with id %s: %v", md5str, err)
		w.WriteHeader(404)
		w.Write([]byte("404"))
		return
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

	ssdbclient = redis.NewClient(&redis.Options{
		Addr:     "192.168.21.225:8888",
		Password: "",
		DB:       0,
	})
	pong, err := ssdbclient.Ping().Result()
	fmt.Println(pong, err)

	log.Fatal(http.ListenAndServe(":8080", router))
}
