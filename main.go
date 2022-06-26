package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Clientes struct {
	id    int32
	Cnpj  string
	Razão string
}

var Empresas []Clientes = []Clientes{
	Clientes{
		id:    1,
		Cnpj:  "06134364584",
		Razão: "italo Costa",
	},
	Clientes{
		id:    2,
		Cnpj:  "06597838517",
		Razão: "Laiane Carmo",
	},
}

func rotas(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bem Vindo!\n")
}

func listarClientes(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	encoder.Encode(Empresas)
}

func configRotas() {
	http.HandleFunc("/", rotas)
	http.HandleFunc("/empresas", listarClientes)
}

func ConfigServidor() {
	configRotas()
	fmt.Println("Servidor esta rodando na Porta 1337")
	http.ListenAndServe(":1337", nil)
}

func main() {

	ConfigServidor()

}
