package rotas

import "net/http"

var rotasUsuarios = []Rota{
	{
		Uri:    "/usuarios",
		Metodo: http.MethodPost,
		Funcao: func(w http.ResponseWriter, r *http.Request) {

		},
		RequerAutenticacao: false,
	}, {
		Uri:    "/usuarios",
		Metodo: http.MethodGet,
		Funcao: func(w http.ResponseWriter, r *http.Request) {

		},
		RequerAutenticacao: false,
	}, {
		Uri:    "/usuarios/{usuarioId}",
		Metodo: http.MethodGet,
		Funcao: func(w http.ResponseWriter, r *http.Request) {

		},
		RequerAutenticacao: false,
	}, {
		Uri:    "/usuarios/{usuarioId}",
		Metodo: http.MethodPut,
		Funcao: func(w http.ResponseWriter, r *http.Request) {

		},
		RequerAutenticacao: false,
	}, {
		Uri:    "/usuarios/{usuarioId}",
		Metodo: http.MethodDelete,
		Funcao: func(w http.ResponseWriter, r *http.Request) {

		},
		RequerAutenticacao: false,
	},
}
