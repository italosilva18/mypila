package rotas

import (
	"api/src/controllers"
	"net/http"
)

// rotasUsuarios é um array que contém várias rotas relacionadas aos usuários.
var rotasUsuarios = []Rota{
	{
		Uri:                "/usuarios", // Rota para criar um novo usuário
		Metodo:             http.MethodPost,
		Funcao:             controllers.CriarUsuario,
		RequerAutenticacao: false, // Não requer autenticação para criar um usuário
	}, {
		Uri:                "/usuarios", // Rota para buscar todos os usuários
		Metodo:             http.MethodGet,
		Funcao:             controllers.BuscarUsuarios,
		RequerAutenticacao: false,
	}, {
		Uri:                "/usuarios/{usuarioId}", // Rota para buscar um usuário específico por ID
		Metodo:             http.MethodGet,
		Funcao:             controllers.BuscarUsuario,
		RequerAutenticacao: false,
	}, {
		Uri:                "/usuarios/{usuarioId}", // Rota para atualizar um usuário específico por ID
		Metodo:             http.MethodPut,
		Funcao:             controllers.AtualizarUsuario,
		RequerAutenticacao: false,
	}, {
		Uri:                "/usuarios/{usuarioId}", // Rota para deletar um usuário específico por ID
		Metodo:             http.MethodDelete,
		Funcao:             controllers.DeletarUsuario,
		RequerAutenticacao: false,
	},
}
