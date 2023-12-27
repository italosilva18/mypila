package router

import "github.com/gorilla/mux"

//gerar vai retornar uma router com as rotas configuradas
func Gerar() *mux.Router {
	return mux.NewRouter()
}
