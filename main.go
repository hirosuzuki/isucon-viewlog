package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Printf("%s\n", err.Error())
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-type", "text/html")
	traces := getTraces()
	if t.Execute(w, map[string]interface{}{"traces": traces}) != nil {
		log.Printf("%s\n", err.Error())
	}
}

func checkTraceID(v string) bool {
	rep := regexp.MustCompile(`^[0-9]{8}-[0-9]{6}$`)
	return rep.MatchString(v)
}

type Trace struct {
	ID            string `json:"id"`
	CPUProf       bool   `json:"cpuprof"`
	Accesslog     bool   `json:"accesslog"`
	SQLLog        bool   `json:"sqllog"`
	VMStat        bool   `json:"vmstat"`
	AccessLogSize int64
	SQLLogSize    int64
	PerfLogSize   int64
	ExecAt        time.Time
}

func getFileSize(filename string) int64 {
	file, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return 0
	}
	return stat.Size()
}

func getTraces() []Trace {
	files, err := ioutil.ReadDir("./logs")
	if err != nil {
		panic(err)
	}
	traceList := []Trace{}
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	for _, file := range files {
		if checkTraceID(file.Name()) {
			traceID := file.Name()
			t, _ := time.Parse("20060102-150405", traceID)
			jt := t.In(jst)
			trace := Trace{ID: traceID, ExecAt: jt}
			trace.AccessLogSize = getFileSize("./logs/" + traceID + "/access.log")
			trace.SQLLogSize = getFileSize("./logs/" + traceID + "/sql.log")
			trace.PerfLogSize = getFileSize("./logs/" + traceID + "/perf.log")
			traceList = append(traceList, trace)
		}
	}
	sort.Slice(traceList, func(i int, j int) bool { return traceList[i].ID > traceList[j].ID })
	return traceList
}

func parseHandler(w http.ResponseWriter, r *http.Request, filename string, progname string) {
	r.ParseForm()
	id := r.FormValue("id")
	if !checkTraceID(id) {
		return
	}
	w.WriteHeader(200)
	w.Header().Add("Content-type", "text/plain")

	fp, err := os.Open("./logs/" + id + "/" + filename)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command(progname)
	cmd.Stdin = fp
	cmd.Stdout = w
	cmd.Run()
}

func kataribeHandler(w http.ResponseWriter, r *http.Request) {
	parseHandler(w, r, "access.log", "./kataribe")
}

func alpHandler(w http.ResponseWriter, r *http.Request) {
	parseHandler(w, r, "access.log", "./alp.sh")
}

func sqlparseHandler(w http.ResponseWriter, r *http.Request) {
	parseHandler(w, r, "sql.log", "./parse_log.py")
}

func perfparseHandler(w http.ResponseWriter, r *http.Request) {
	parseHandler(w, r, "perf.log", "./parse_log.py")
}

func outputFileHandler(w http.ResponseWriter, r *http.Request, filename string) {
	r.ParseForm()
	id := r.FormValue("id")
	if !checkTraceID(id) {
		return
	}
	fp, err := os.Open("./logs/" + id + "/" + filename)
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		panic(err)
	}
	fp.Close()
	w.WriteHeader(200)
	w.Header().Add("Content-type", "text/plain")
	w.Write(buf)
}

func sqllogHandler(w http.ResponseWriter, r *http.Request) {
	outputFileHandler(w, r, "sql.log")
}

func accesslogHandler(w http.ResponseWriter, r *http.Request) {
	outputFileHandler(w, r, "access.log")
}

func vmstatHandler(w http.ResponseWriter, r *http.Request) {
	outputFileHandler(w, r, "vmstat.log")
}

func perflogHandler(w http.ResponseWriter, r *http.Request) {
	outputFileHandler(w, r, "perf.log")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/favicon.ico", fs)

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/kataribe/", kataribeHandler)
	r.HandleFunc("/alp/", alpHandler)
	r.HandleFunc("/sqlparse/", sqlparseHandler)
	r.HandleFunc("/perfparse/", perfparseHandler)
	r.HandleFunc("/vmstat/", vmstatHandler)
	r.HandleFunc("/accesslog/", accesslogHandler)
	r.HandleFunc("/sqllog/", sqllogHandler)
	r.HandleFunc("/perflog/", perflogHandler)

	r.Use(loggingMiddleware)

	http.Handle("/", r)

	log.Printf("Start App: listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
