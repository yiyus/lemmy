/*
Copyright 2012 Jesús Galán López (yiyus)

Originally based on https://github.com/nf/goplayer
by Andrew Gerrand.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"os"
)

type Entry struct {
	Name  string // name of the object
	IsDir bool
	Mode  os.FileMode
}

const (
	filePrefix = "/f/"
)

var (
	addr = flag.String("http", ":8080", "http listen address")
	root = flag.String("root", os.Getenv("HOME")+"/music/", "music root")
	web string
)

func main() {
	flag.Parse()
	web = os.Getenv("LEMMY_WEB")
	log.Print("root = ", *root)
	log.Print("web = ", web)
	http.HandleFunc("/", Web)
	http.HandleFunc(filePrefix, File)
	http.ListenAndServe(*addr, nil)
}

func Web(w http.ResponseWriter, r *http.Request) {
	if web == "" {
		if r.URL.Path != "/" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.Write([]byte(index));
		return;
	}
	fn := filepath.Join(web, r.URL.Path)
	_, err := os.Stat(fn)
	log.Print("Web file called: ", fn)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, fn)
}

func File(w http.ResponseWriter, r *http.Request) {
	fn := filepath.Join(*root, r.URL.Path[len(filePrefix):])
	fi, err := os.Stat(fn)
	log.Print("File called: ", fn)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if fi.IsDir() {
		serveDirectory(fn, w, r)
		return
	}
	http.ServeFile(w, r, fn)
}

func serveDirectory(fn string, w http.ResponseWriter,
	r *http.Request) {
	defer func() {
		if err, ok := recover().(error); ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()
	d, err := os.Open(fn)
	if err != nil {
		panic(err)
	}
	log.Print("serverDirectory called: ", fn)

	files, err := d.Readdir(-1)
	if err != nil {
		panic(err)
	}

	// Json Encode isn't working with the FileInfo interface,
	// therefore populate an Array of Entry and add the Name method
	entries := make([]Entry, len(files), len(files))

	for k := range files {
		//log.Print(files[k].Name())
		entries[k].Name = files[k].Name()
		entries[k].IsDir = files[k].IsDir()
		entries[k].Mode = files[k].Mode()
	}

	j := json.NewEncoder(w)

	if err := j.Encode(&entries); err != nil {
		panic(err)
	}
}
