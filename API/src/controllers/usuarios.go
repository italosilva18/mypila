package controllers

import (
	"api/src/banco"
	"api/src/modelos"
	"api/src/repositorios"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// Controlador CriarUsuario (insere um usuario no banco de dados ) .
func CriarUsuario(w http.ResponseWriter, r *http.Request) {
	corpoRequest, erro := ioutil.ReadAll(r.Body)
	if erro != nil{
		log.Fatal(erro)
	}
	var usuario modelos.Usuario
	if erro:= json.Unmarshal(corpoRequest,&usuario); erro!=nil{
		log.Fatal(erro)		
	}
	db, erro:= banco.Conectar()
	if erro!=nil(erro)

	repositorio := repositrios.NovoRepositorioDeUsuarios(db)
	repositorio.Criar(usuario)
}

// BuscarUsuarios busca todos os usuários no banco de dados.
func BuscarUsuarios(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Buscando Todos Usuarios"))
}

// BuscarUsuario busca um usuário específico no banco de dados.
func BuscarUsuario(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Buscando um Usuario"))
}

// AtualizarUsuario atualiza as informações de um usuário no banco de dados.
func AtualizarUsuario(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Atualizando Usuario"))
}

// DeletarUsuario exclui um usuário do banco de dados.
func DeletarUsuario(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Deletando Usuario"))
}
