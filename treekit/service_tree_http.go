package treekit

import (
	"net/http"
)

type Server struct {
	tree *MsTree
	bind string
	mux  *http.ServeMux
}

func (srv *Server) NewHandler(path string, fn func(tree *MsTree) http.HandlerFunc) {
	srv.mux.HandleFunc(path, fn(srv.tree))
}

func (srv *Server) Listen(x func(error)) {
	go func() {
		err := http.ListenAndServe(srv.bind, srv.mux)
		x(err)
	}()
}

func (mt *MsTree) LazyWeb(bind string) *Server {
	te := &Server{
		mux:  http.NewServeMux(),
		tree: mt,
		bind: bind,
	}
	return te
}
