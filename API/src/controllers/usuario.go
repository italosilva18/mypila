package controllers

import "net/http"

// Controlador CriarUsuario (insere um usuario no banco de dados ) .
func CriarUsuario(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Criando Usuario"))
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
