package main

import (
	"io/ioutil"
	"net/http"
	"html/template"
	"regexp"  //正規表現
	"errors"
	"log"
	"strings"
)


type Page struct {
	Title string //タイトル
	Body  []byte //タイトルの中身
}

var titleValidate = regexp.MustCompile("^[a-zA-Z0-9]+$")

var expend_text = ".txt"

//タイトルのチェックを行う
func getTitle(w http.ResponseWriter, r *http.Request)(title string , err error){
	title =  r.URL.Path[lenPath:]
	if !titleValidate.MatchString(title){
		http.NotFound(w,r)
		err = errors.New("Invalid Title")
		log.Print(err)
	}
	
	return 
}

//パスのアドレスを設定して文字の長さを定数として持つ
const lenPath = len("/view/")

var templates = make(map[string]*template.Template)

//初期化関数
func init(){
	for _,tepl := range []string{"edit", "view"}{
		 t := template.Must(template.ParseFiles(tepl + ".html"))
		 templates[tepl] = t
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title,err := getTitle(w,r)
	if err != nil{
		return
	}
	p, err := loadPage(title)
	if err != nil{
		http.Redirect(w, r, "/edit/"+ title,http.StatusFound)
		return
	}
	renderTemplate(w,"view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request){
	title,err := getTitle(w,r)
	if err != nil{
		return
	}
	p, err := loadPage(title)
	if err != nil{
		p = &Page{Title: title}
	}
	renderTemplate(w,"edit",p)
}

func saveHandler(w http.ResponseWriter, r *http.Request){
	title,err := getTitle(w,r)
	if err != nil{
		return
	}
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err = p.save()
	if err != nil{
		http.Error(w, err.Error(),http.StatusInternalServerError)
	}
	http.Redirect(w,r,"/view/" + title, http.StatusFound)
}

func topHandler(w http.ResponseWriter,r *http.Request){
	//今の階層にある.txtを取得する
	files,err := ioutil.ReadDir("./")
	if err != nil{
		err =	errors.New("所定のディレクトリにテキストファイルが存在しません")
		log.Print(err)
		return
	}
	var pathes []string //テキストデータの名前
	var fileName []string //テキストデータのファイル名
	
	for _, file := range files{
		//対象ファイルの.txtファイルのみを取得する
		if strings.HasSuffix(file.Name(), expend_text){
				 fileName = strings.Split(string(file.Name()), expend_text)
				 pathes = append(pathes, fileName[0])
		}

		if pathes == nil{
			err = errors.New("テキストファイルが存在しません")
			log.Print(err)
		}
	}
	t := template.Must(template.ParseFiles("top.html"))
	err = t.Execute(w,pathes)
	if err != nil{
		http.Error(w, err.Error(),http.StatusInternalServerError)
	}
}

func renderTemplate(w http.ResponseWriter, tepl string,p *Page){
	err := templates[tepl].Execute(w,p)
	if err != nil{
		http.Error(w, err.Error(),http.StatusInternalServerError)
	}
}

//テキストファイルの保存メソッド
func (p *Page) save() error {
	//タイトルの名前でテキストファイルを作成して保存します。
	filename := p.Title + ".txt"
	//0600は、テキストデータを書き込んだり読み込んだりする権限を設定しています。
	return ioutil.WriteFile(filename, p.Body, 0600)
}

//titleからファイル名を読み込んで新しいPageのポインタを返す
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	//errに値が入ったらエラーとしてbodyの値をnilにして返す
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}


func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/",editHandler)
	http.HandleFunc("/save/",saveHandler)
	http.HandleFunc("/top/" ,topHandler)
	http.ListenAndServe(":8080", nil)
}
