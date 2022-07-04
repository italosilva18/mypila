package root

import "net/http"

func ConfigRotas() {
	http.HandleFunc("/", Rotas)
	http.HandleFunc("/empresas", ListarClientes)
}
