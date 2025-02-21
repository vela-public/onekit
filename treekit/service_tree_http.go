package treekit

import (
	"net/http"
)

const (
	socket = "127.0.0.1:12345"
)

type Server struct {
	tree *MsTree
	bind string
}

func (te *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		buf := te.tree.Doc()
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(buf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	/*
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
			w.Write(te.tree.Doc())

		})

	*/
	return mux
}

func (te *Server) Listen() error {
	return http.ListenAndServe(socket, te.Handler())
}

func (mt *MsTree) HttpServer() {

}
