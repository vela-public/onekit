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

	return mux
}

func (te *Server) Listen() error {
	return http.ListenAndServe(socket, te.Handler())
}

func (mt *MsTree) HttpServer() {

}
