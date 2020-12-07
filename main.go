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

	"github.com/julienschmidt/httprouter"
	"gopkg.in/gographics/imagick.v2/imagick"
)

//ImageInfo file infomation
type ImageInfo struct {
	//file name
	Filename string
	//file id
	Md5 string
}

type uploadresonse struct {
	Ret   string
	Files []ImageInfo
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

type conf struct {
	IP   string `yaml:"ip"`
	Port string `yaml:"port"`
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

	//图片已经存在
	dir, err := os.Stat(saveName)
	if err == nil {
		if dir.IsDir() {
			fmt.Printf("image %s already existed\n", saveName)
			return
		}
	}

	os.MkdirAll(saveName, 0777)

	f, err := os.OpenFile(saveName+"0_0", os.O_WRONLY|os.O_CREATE, 0666) // 此处假设当前目录下已存在test目录
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	//ssdbclient.Set(MD5String, buf, 0)
	f.Write(buf)

}
func uploadpost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseMultipartForm(32 << 20)
	//file, handler, err := r.FormFile("uploadfile")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer file.Close()
	var resp uploadresonse
	files := r.MultipartForm.File["uploadfile"]
	for _, v := range files {
		file, err := v.Open()

		if err != nil {
			fmt.Println("Open MultipartForm file failed")
			continue
		}
		defer file.Close()
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

		fileinfo := ImageInfo{v.Filename, MD5String}
		resp.Files = append(resp.Files, fileinfo)

	}

	//ssdbclient.Set(MD5String, buf, 0)
	w.Header().Set("Content-type", "application/json;charset=utf-8")
	resp.Ret = "true"
	//b, err := json.Marshal(fileinfo)
	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//fmt.Fprintf(w, "%v", handler.Header)
	fmt.Println(string(b))
	w.Write(b)
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

func processImage(path string) []byte {

	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	err := mw.ReadImage(path)

	if err != nil {
		//fmt.Printf("Error with id %s: %v", md5str, err)
		// wr.WriteHeader(404)
		// wr.Write([]byte("404"))
		return nil
	}

	w := mw.GetImageWidth()
	h := mw.GetImageHeight()

	hWidth := uint(w / 2)
	hHeight := uint(h / 2)

	err = mw.ResizeImage(hWidth, hHeight, imagick.FILTER_LANCZOS, 1)
	if err != nil {
		// wr.WriteHeader(404)
		// wr.Write([]byte("404"))
		return nil
	}

	err = mw.SetImageCompressionQuality(95)
	if err != nil {
		// wr.WriteHeader(404)
		// wr.Write([]byte("404"))
		return nil
	}

	buf := mw.GetImageBlob()
	return buf
}

func downloadImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	//data, err := ioutil.ReadFile(path)
	bg := time.Now().UnixNano()
	data := processImage(path)
	end := time.Now().UnixNano()
	fmt.Println("-------Spend time: ", (end-bg)/1e6)
	if len(data) == 0 {
		//if err != nil {
		fmt.Printf("Error with id %s: %v", md5str, err)
		w.WriteHeader(404)
		w.Write([]byte("404"))
		return
	}

	//w.Header().Set("Content-Type", "application/x-msdownload")
	//w.Header().Set("Content-Type", "image/jpeg")
	//fmt.Println(w.Header()["Content-Type"])
	fileName := md5str + ".jpg"
	w.Header().Set("Content-Disposition", "attachment;filename="+fileName)
	w.Header().Add("Content-Type", "application/x-msdownload")
	fmt.Println(w.Header()["Content-Type"])

	w.Write(data)
}

func main() {
	router := httprouter.New()
	router.GET("/", index)
	router.GET("/hello/:name", hello)
	router.GET("/image/:name", downloadImage)

	router.GET("/user/:uid", getuser)
	router.GET("/upload", uploadget)
	router.POST("/adduser/:uid", adduser)
	router.POST("/upload", uploadpost)
	router.DELETE("/deluser/:uid", deleteuser)
	router.PUT("/moduser/:uid", modifyuser)

	// var s Serverslice
	// s.Servers = append(s.Servers, Server{ServerName: "Shanghai_VPN", ServerIP: "127.0.0.1"})
	// s.Servers = append(s.Servers, Server{ServerName: "Beijing_VPN", ServerIP: "127.0.0.2"})
	// b, err := json.Marshal(s)
	// if err != nil {
	// 	fmt.Println("json err:", err)
	// }
	// fmt.Println(string(b))

	//yamlFile, err := ioutil.ReadFile("conf.yaml")
	//if err != nil {
	//	log.Printf("yamlFile.Get err   #%v ", err)
	//	return
	//}
	//
	//var c conf
	//err = yaml.Unmarshal(yamlFile, &c)
	//if err != nil {
	//	log.Printf("Unmarshal conf err   #%v ", err)
	//	return
	//}
	//fmt.Printf("--- c:\n%v\n\n", c)

	imagick.Initialize()
	defer imagick.Terminate()

	log.Fatal(http.ListenAndServe(":4869", router))
}
