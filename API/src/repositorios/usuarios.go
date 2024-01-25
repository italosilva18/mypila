package repositorios

import (
	"api/src/modelos"
	"database/sql"
)

type usuarios struct {
	db *sql.DB
}

// NovoRepositorioUsuarios cria um repositorio de usuarios.
func NovoRepositorioDeUsuarios(db *sql.DB) *Usuarios {
	return &Usuarios{db}
}

// criar insere um usuario no banco de dados
func (u.usuarios) Criar(usuario modelos.Usuario) (uint64, error) {
	return 0, nil
}
