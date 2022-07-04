package main

import (
	"encoding/json"
	"fmt"
	"log"
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

func ConfigRotas() {
	http.HandleFunc("/", rotas)
	http.HandleFunc("/empresas", listarClientes)
}

func rotas(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bem Vindo, %s!\n", "Italo.Costa")
}

func listarClientes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Empresas)
}

func ConfigServidor() {

	ConfigRotas()

	log.Fatal(http.ListenAndServe(":1337", nil))
}

func main() {

	serv := ` " Servidor Esta Rodando na Porta 1337 " 
       __	  .     __     __
 	 /    \   |   /      /    \
	|    __   |   \-- \  |
	 \ __ /   |    __ /  \ __ /
	  
	`
	fmt.Println(serv)
	ConfigServidor()

}
