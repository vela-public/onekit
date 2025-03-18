package treekit

import (
	"net/http"
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

	mux.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		path := r.URL.Query().Get("path")
		err := te.tree.DoServiceFile(name, path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		buf := te.tree.Doc()
		_, err = w.Write(buf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return mux
}

func (te *Server) Listen() error {
	return http.ListenAndServe(te.bind, te.Handler())
}

func (mt *MsTree) Web(bind string, x func(e error)) {
	te := &Server{
		tree: mt,
		bind: bind,
	}

	go func() {
		err := te.Listen()
		x(err)
	}()
}
