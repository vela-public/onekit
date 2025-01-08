package toolkit

import (
	"github.com/vela-public/onekit/taskit"
	"net/http"
	"os"
)

const (
	socket = "127.0.0.1:12345"
)

type TaskitEx struct {
	tree *taskit.Tree
	bind string
}

func (te *TaskitEx) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		buf := te.tree.View()
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(buf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		path := r.URL.Query().Get("path")
		text, err := os.ReadFile(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = te.tree.Register(key, text)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		te.tree.Wakeup()
		w.WriteHeader(http.StatusOK)
		w.Write(te.tree.View())

	})
	return mux
}

func (te *TaskitEx) Listen() error {
	return http.ListenAndServe(socket, te.Handler())
}

func NewTaskitEx(tree *taskit.Tree) *TaskitEx {
	return &TaskitEx{
		tree: tree,
		bind: socket,
	}
}
