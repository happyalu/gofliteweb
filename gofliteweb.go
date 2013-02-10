// Copyright 2013, Carnegie Mellon University. All Rights Reserved.
// Use of this code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Author: Alok Parlikar <aup@cs.cmu.edu>

// GoFliteWeb: Flite Speech Synthesis Demo on the Web
package main

import (
	"flag"
	"github.com/happyalu/goflite"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
)

var (
	serverAddr = flag.String("addr", "localhost:8080", "Server Address")
	voxpath    = flag.String("voxpath", "", "Absolute path to *.flitevox files")
	voicelist  []string
)

func main() {
	flag.Parse()

	// Look for *.flitevox files in given voxpath and add voices to goflite
	if *voxpath != "" {
		voxdirFiles, _ := ioutil.ReadDir(*voxpath)
		for _, v := range voxdirFiles {
			name := v.Name()
			if strings.HasSuffix(name, ".flitevox") {
				err := goflite.AddVoice(name, path.Join(*voxpath, name))
				if err != nil {
					log.Printf("FAILED to add voice %s", name)
				} else {
					voicelist = append(voicelist, name)
					log.Printf("ADDED voice %s", name)
				}
			}
		}
	}

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/wav", WaveHandler)
	http.ListenAndServe(*serverAddr, nil)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	const indexpage = `
<!DOCTYPE html>
<html>
<head>
  <title>Flite Synthesis Demo</title>
  <script type="text/javascript">
	function play_tts() {
		var text = document.getElementById('textarea').value;
		var voice = document.getElementById('voice').options[document.getElementById('voice').selectedIndex].text;
		var audio = document.getElementById('player');
		audio.setAttribute('src', '/wav?text=' + encodeURIComponent(text) + '&voice=' + encodeURIComponent(voice));
		audio.play();
	}
  </script>
</head>
<body>
  <h1>Flite Synthesis Demo</h1>
  
  Choose Voice: 
  <select id="voice">
	{{range .Voices}}
	<option value="{{.}}">{{.}}</option>
	{{end}}
    <option value="Default">Default</option>
  </select> <br /> <br />
  <textarea rows=3 cols=80 id="textarea" name="text">A whole joy was reaping, but they've gone south. Go fetch azure mike!</textarea>
  <br /> <br />
  <input type="submit" value="Speak!" onclick="play_tts();"> <br />
  <audio id="player"></audio>
  <br />
  <br />
  <br />
  <small>Powered by <a href="www.github.com/happyalu/goflite">GoFlite</a>.</small>
  </body>
</html>`

	type Data struct {
		Voices []string
	}

	data := Data{voicelist}

	t := template.Must(template.New("indexpage").Parse(indexpage))
	err := t.Execute(w, data)
	if err != nil {
		log.Println("INDEX: Fail", err)
	}
}

func WaveHandler(w http.ResponseWriter, r *http.Request) {
	text := r.FormValue("text")
	voice := r.FormValue("voice")
	w.Header().Set("Content-Type", "audio/x-wav")

	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	log.Printf("wavegen: %s\t%s\t%s", clientIP, voice, text)

	if voice == "Default" {
		// Empty voice makes goflite choose the default voice
		voice = ""
	}

	wav, err := goflite.TextToWave(text, voice)
	if err != nil {
		log.Printf("WAVE: Could not synthesize %s: %s", voice, text)
		http.Error(w, "Could not Synthesize Speech", 500)
		return
	}

	err = wav.DumpRIFF(w)
	if err != nil {
		log.Println("Could not write waveform")
	}

}
