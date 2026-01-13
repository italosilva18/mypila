package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// CNPJData represents the response from BrasilAPI
type CNPJData struct {
	CNPJ                  string `json:"cnpj"`
	RazaoSocial           string `json:"razao_social"`
	NomeFantasia          string `json:"nome_fantasia"`
	Logradouro            string `json:"logradouro"`
	Numero                string `json:"numero"`
	Complemento           string `json:"complemento"`
	Bairro                string `json:"bairro"`
	Municipio             string `json:"municipio"`
	UF                    string `json:"uf"`
	CEP                   string `json:"cep"`
	DDDTelefone1          string `json:"ddd_telefone_1"`
	SituacaoCadastral     int    `json:"situacao_cadastral"`
	DescricaoSituacao     string `json:"descricao_situacao_cadastral"`
	CNAEFiscal            int    `json:"cnae_fiscal"`
	CNAEFiscalDescricao   string `json:"cnae_fiscal_descricao"`
	DataInicioAtividade   string `json:"data_inicio_atividade"`
	CapitalSocial         float64 `json:"capital_social"`
}

// CNPJResponse is the normalized response for the frontend
type CNPJResponse struct {
	CNPJ        string `json:"cnpj"`
	RazaoSocial string `json:"razaoSocial"`
	NomeFantasia string `json:"nomeFantasia"`
	Logradouro  string `json:"logradouro"`
	Numero      string `json:"numero"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Municipio   string `json:"municipio"`
	UF          string `json:"uf"`
	CEP         string `json:"cep"`
	Telefone    string `json:"telefone"`
	Situacao    string `json:"situacao"`
	Atividade   string `json:"atividade"`
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// CleanCNPJ removes formatting from CNPJ
func CleanCNPJ(cnpj string) string {
	re := regexp.MustCompile(`[^\d]`)
	return re.ReplaceAllString(cnpj, "")
}

// ValidateCNPJ validates if CNPJ has correct format (14 digits)
func ValidateCNPJ(cnpj string) bool {
	clean := CleanCNPJ(cnpj)
	return len(clean) == 14
}

// FormatCNPJ formats CNPJ with punctuation
func FormatCNPJ(cnpj string) string {
	clean := CleanCNPJ(cnpj)
	if len(clean) != 14 {
		return cnpj
	}
	return fmt.Sprintf("%s.%s.%s/%s-%s",
		clean[0:2], clean[2:5], clean[5:8], clean[8:12], clean[12:14])
}

// LookupCNPJ fetches company data from BrasilAPI
func LookupCNPJ(ctx context.Context, cnpj string) (*CNPJResponse, error) {
	clean := CleanCNPJ(cnpj)
	if !ValidateCNPJ(clean) {
		return nil, fmt.Errorf("CNPJ invalido: deve conter 14 digitos")
	}

	url := fmt.Sprintf("https://brasilapi.com.br/api/cnpj/v1/%s", clean)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisicao: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "MyPila/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar CNPJ: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("CNPJ nao encontrado")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erro na API: status %d", resp.StatusCode)
	}

	var data CNPJData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("erro ao processar resposta: %w", err)
	}

	// Build full address
	address := data.Logradouro
	if data.Numero != "" {
		address += ", " + data.Numero
	}
	if data.Complemento != "" {
		address += " - " + data.Complemento
	}
	if data.Bairro != "" {
		address += ", " + data.Bairro
	}

	// Format phone
	phone := strings.TrimSpace(data.DDDTelefone1)

	// Normalize response
	response := &CNPJResponse{
		CNPJ:         FormatCNPJ(data.CNPJ),
		RazaoSocial:  data.RazaoSocial,
		NomeFantasia: data.NomeFantasia,
		Logradouro:   address,
		Numero:       data.Numero,
		Complemento:  data.Complemento,
		Bairro:       data.Bairro,
		Municipio:    data.Municipio,
		UF:           data.UF,
		CEP:          data.CEP,
		Telefone:     phone,
		Situacao:     data.DescricaoSituacao,
		Atividade:    data.CNAEFiscalDescricao,
	}

	return response, nil
}
