package main

import (
	"github.com/niemeyer/qml"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	QmlMain()
}

func QmlMain() error {
	qml.Init(nil)
	engine := qml.NewEngine()

	ctrl := &Control{}
	engine.Context().SetVar("ctrl", ctrl)
	engine.Context().SetVar("results", &ctrl.Results)

	component, err := engine.LoadString("main", mainQml)
	if err != nil {
		return err
	}
	window := component.CreateWindow(nil)
	window.Show()
	window.Wait()
	return nil
}

type Control struct {
	Results Results
}

func (*Control) Quit() {
	os.Exit(0)
}

func (c *Control) Search(q string) {
	res, _ := Search(q)
	c.Results.Len = 0
	qml.Changed(&c.Results, &c.Results.Len)
	c.Results.Len = len(res)
	c.Results.sr = res
	qml.Changed(&c.Results, &c.Results.Len)
}

func (ctrl *Control) Select(idx int) {
	go func() {
		hres, _ := http.Get(ctrl.Results.sr[idx].Url)
		dr, _ := ParseDownload(hres.Body)
		hres, _ = http.Get(dr.Url)
		player <- hres.Body
	}()
}

type Results struct {
	Len int
	sr  []SearchResult
}

func (r *Results) Text(idx int) string {
	return r.sr[idx].Title
}

func Player() chan io.ReadCloser {
	ch := make(chan io.ReadCloser)
	go func() {
		var cmd *exec.Cmd
		for {
			reader := <-ch
			if cmd != nil {
				cmd.Process.Kill()
			}
			cmd = exec.Command(`mplayer`, `-vo`, `null`, `-`)
			pipe, _ := cmd.StdinPipe()
			go func() {
				io.Copy(pipe, reader)
				reader.Close()
				pipe.Close()
			}()
			go func() {
				cmd.Run()
			}()
		}
	}()
	return ch
}

var player = Player()

var mainQml = `
import QtQuick 2.1
import QtQuick.Controls 1.0

ApplicationWindow {
	
	Action {
		id: quitAction
		text: "&Quit"
		shortcut: "Ctrl+Q"
		onTriggered: ctrl.quit()
	}

	Column {
      		width: parent.width
      		height: parent.height
		
		TextField {
			onAccepted: ctrl.search(text)
			width: parent.width
		}
		
		ListView {
			y: 14
			height: parent.height-y
		      width: parent.width;
		      model: results.len
		      delegate: Rectangle {
				height: 14
				Text {
				      text: results.text(index)
					MouseArea {
						anchors.fill: parent
						onClicked: ctrl.select(index)
					}
	      	  		}
			}
    		}
	}
}
`
