package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

// formatCurrency formata valor em BRL
func formatCurrency(value float64) string {
	return fmt.Sprintf("R$ %.2f", value)
}

// formatDate formata data em PT-BR
func formatDate(t time.Time) string {
	return t.Format("02/01/2006")
}

// GenerateQuotePDF gera PDF do orçamento
func GenerateQuotePDF(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	id := c.Params("id")
	quoteID, err := uuid.Parse(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID invalido")
	}

	quote, err := helpers.ValidateQuoteOwnership(c, quoteID)
	if err != nil {
		return err
	}

	quote.Items, _ = getQuoteItems(ctx, quote.ID)

	var template *models.QuoteTemplate
	if quote.TemplateID != nil {
		var t models.QuoteTemplate
		err = database.QueryRow(ctx,
			`SELECT id, company_id, name, header_text, footer_text, terms_text, primary_color, logo_url, is_default, created_at, updated_at
			 FROM quote_templates WHERE id = $1`,
			*quote.TemplateID).Scan(&t.ID, &t.CompanyID, &t.Name, &t.HeaderText, &t.FooterText,
			&t.TermsText, &t.PrimaryColor, &t.LogoURL, &t.IsDefault, &t.CreatedAt, &t.UpdatedAt)
		if err == nil {
			template = &t
		}
	}

	var company models.Company
	var logoURL *string
	err = database.QueryRow(ctx,
		`SELECT id, user_id, name, logo_url, created_at, updated_at FROM companies WHERE id = $1`,
		quote.CompanyID).Scan(&company.ID, &company.UserID, &company.Name, &logoURL, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar dados da empresa"})
	}
	if logoURL != nil {
		company.LogoURL = *logoURL
	}

	// Criar PDF
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Cor primária
	primaryColor := "#374151" // gray-700
	if template != nil && template.PrimaryColor != "" {
		primaryColor = template.PrimaryColor
	}
	r, g, b := hexToRGB(primaryColor)

	// ========== CABEÇALHO ==========
	headerHeight := 32.0
	pdf.SetFillColor(r, g, b)
	pdf.Rect(0, 0, 210, headerHeight, "F")

	// Logo
	nameX := 12.0
	if company.LogoURL != "" {
		logoData, logoType := getLogoData(company.LogoURL)
		if logoData != nil && logoType != "" {
			pdf.RegisterImageOptionsReader("logo", fpdf.ImageOptions{ImageType: logoType, ReadDpi: true}, bytes.NewReader(logoData))
			pdf.ImageOptions("logo", 10, 4, 24, 24, false, fpdf.ImageOptions{}, 0, "")
			nameX = 38.0
		}
	}

	// Nome da empresa
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(nameX, 8)
	pdf.Cell(90, 6, removeAccents(company.Name))

	// Subtítulo
	if template != nil && template.HeaderText != "" {
		pdf.SetFont("Arial", "", 8)
		pdf.SetXY(nameX, 16)
		pdf.Cell(90, 5, removeAccents(template.HeaderText))
	}

	// Número e datas (lado direito)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(145, 6)
	pdf.Cell(55, 6, quote.Number)

	pdf.SetFont("Arial", "", 8)
	pdf.SetXY(145, 14)
	pdf.Cell(55, 4, fmt.Sprintf("Data: %s", formatDate(quote.CreatedAt)))
	pdf.SetXY(145, 19)
	pdf.Cell(55, 4, fmt.Sprintf("Valido ate: %s", formatDate(quote.ValidUntil)))

	// ========== TÍTULO ==========
	pdf.SetY(headerHeight + 4)
	pdf.SetTextColor(r, g, b)
	pdf.SetFont("Arial", "B", 13)
	pdf.SetX(10)
	pdf.Cell(190, 6, removeAccents(quote.Title))

	// ========== CLIENTE ==========
	pdf.SetY(pdf.GetY() + 8)
	pdf.SetFillColor(245, 245, 245)
	pdf.SetDrawColor(220, 220, 220)

	// Calcular altura do box do cliente
	clientBoxHeight := 18.0
	if quote.ClientAddress != "" {
		clientBoxHeight = 24.0
	}

	clientY := pdf.GetY()
	pdf.Rect(10, clientY, 190, clientBoxHeight, "FD")

	// Nome e documento na mesma linha
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(14, clientY+3)
	pdf.Cell(100, 5, removeAccents(quote.ClientName))

	if quote.ClientDocument != "" {
		pdf.SetFont("Arial", "", 9)
		pdf.SetXY(130, clientY+3)
		pdf.Cell(66, 5, quote.ClientDocument)
	}

	// Contato
	pdf.SetFont("Arial", "", 9)
	pdf.SetXY(14, clientY+9)
	contact := ""
	if quote.ClientPhone != "" {
		contact = quote.ClientPhone
	}
	if quote.ClientEmail != "" {
		if contact != "" {
			contact += " | "
		}
		contact += quote.ClientEmail
	}
	pdf.Cell(180, 5, contact)

	// Endereço
	if quote.ClientAddress != "" {
		pdf.SetXY(14, clientY+15)
		addr := quote.ClientAddress
		if quote.ClientCity != "" {
			addr += " - " + quote.ClientCity
		}
		if quote.ClientState != "" {
			addr += "/" + quote.ClientState
		}
		if quote.ClientZipCode != "" {
			addr += " - " + quote.ClientZipCode
		}
		pdf.Cell(180, 5, removeAccents(addr))
	}

	pdf.SetY(clientY + clientBoxHeight + 4)

	// ========== DESCRIÇÃO ==========
	if quote.Description != "" {
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(r, g, b)
		pdf.SetX(10)
		pdf.Cell(0, 5, "DESCRICAO")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(60, 60, 60)
		pdf.SetX(10)
		pdf.MultiCell(190, 4, removeAccents(quote.Description), "", "L", false)
		pdf.Ln(3)
	}

	// ========== TABELA DE ITENS ==========
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(r, g, b)
	pdf.SetX(10)
	pdf.Cell(0, 5, "ITENS")
	pdf.Ln(6)

	// Cabeçalho da tabela
	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 8)

	colW := []float64{95, 20, 35, 40} // Descrição, Qtd, Unit, Total
	pdf.SetX(10)
	pdf.CellFormat(colW[0], 7, "Descricao", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colW[1], 7, "Qtd", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[2], 7, "Unitario", "1", 0, "R", true, 0, "")
	pdf.CellFormat(colW[3], 7, "Total", "1", 1, "R", true, 0, "")

	// Itens
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 8)
	alt := false

	for _, item := range quote.Items {
		if pdf.GetY() > 265 {
			pdf.AddPage()
			// Repetir cabeçalho
			pdf.SetFillColor(r, g, b)
			pdf.SetTextColor(255, 255, 255)
			pdf.SetFont("Arial", "B", 8)
			pdf.SetX(10)
			pdf.CellFormat(colW[0], 7, "Descricao", "1", 0, "L", true, 0, "")
			pdf.CellFormat(colW[1], 7, "Qtd", "1", 0, "C", true, 0, "")
			pdf.CellFormat(colW[2], 7, "Unitario", "1", 0, "R", true, 0, "")
			pdf.CellFormat(colW[3], 7, "Total", "1", 1, "R", true, 0, "")
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont("Arial", "", 8)
		}

		if alt {
			pdf.SetFillColor(250, 250, 250)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		pdf.SetX(10)
		pdf.CellFormat(colW[0], 6, truncateString(removeAccents(item.Description), 50), "1", 0, "L", true, 0, "")
		pdf.CellFormat(colW[1], 6, fmt.Sprintf("%.0f", item.Quantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW[2], 6, formatCurrency(item.UnitPrice), "1", 0, "R", true, 0, "")
		pdf.CellFormat(colW[3], 6, formatCurrency(item.Total), "1", 1, "R", true, 0, "")
		alt = !alt
	}

	// ========== TOTAIS ==========
	pdf.Ln(3)
	totX := 120.0

	// Subtotal
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.SetX(totX)
	pdf.CellFormat(40, 6, "Subtotal:", "", 0, "L", false, 0, "")
	pdf.CellFormat(40, 6, formatCurrency(quote.Subtotal), "", 1, "R", false, 0, "")

	// Desconto
	if quote.Discount > 0 {
		pdf.SetTextColor(200, 50, 50)
		pdf.SetX(totX)
		lbl := "Desconto:"
		val := quote.Discount
		if quote.DiscountType == models.DiscountPercent {
			lbl = fmt.Sprintf("Desconto (%.0f%%):", quote.Discount)
			val = quote.Subtotal * quote.Discount / 100
		}
		pdf.CellFormat(40, 6, lbl, "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("-"+formatCurrency(val)), "", 1, "R", false, 0, "")
	}

	// Total
	pdf.Ln(1)
	pdf.SetX(totX)
	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(40, 9, "TOTAL", "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 9, formatCurrency(quote.Total), "1", 1, "R", true, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)

	// ========== OBSERVAÇÕES ==========
	if quote.Notes != "" {
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(r, g, b)
		pdf.SetX(10)
		pdf.Cell(0, 5, "OBSERVACOES")
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(60, 60, 60)
		pdf.SetX(10)
		pdf.MultiCell(190, 4, removeAccents(quote.Notes), "", "L", false)
		pdf.Ln(3)
	}

	// ========== TERMOS ==========
	if template != nil && template.TermsText != "" {
		if pdf.GetY() > 245 {
			pdf.AddPage()
		}
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(r, g, b)
		pdf.SetX(10)
		pdf.Cell(0, 5, "TERMOS E CONDICOES")
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 7)
		pdf.SetTextColor(100, 100, 100)
		pdf.SetX(10)
		pdf.MultiCell(190, 3, removeAccents(template.TermsText), "", "L", false)
	}

	// ========== RODAPÉ ==========
	pdf.SetY(-12)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY()-2, 200, pdf.GetY()-2)
	pdf.SetTextColor(140, 140, 140)
	pdf.SetFont("Arial", "", 7)
	footerText := fmt.Sprintf("Gerado em %s | %s", formatDate(time.Now()), quote.Number)
	if template != nil && template.FooterText != "" {
		footerText = removeAccents(template.FooterText) + " | " + footerText
	}
	pdf.SetX(10)
	pdf.Cell(190, 4, footerText)

	// Gerar PDF
	var buf bytes.Buffer
	if err = pdf.Output(&buf); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao gerar PDF"})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.pdf\"", quote.Number))
	return c.Send(buf.Bytes())
}

func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 55, 65, 81 // gray-700
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func removeAccents(s string) string {
	r := strings.NewReplacer(
		"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
		"Á", "A", "À", "A", "Ã", "A", "Â", "A", "Ä", "A",
		"É", "E", "È", "E", "Ê", "E", "Ë", "E",
		"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
		"Ó", "O", "Ò", "O", "Õ", "O", "Ô", "O", "Ö", "O",
		"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
		"Ç", "C", "Ñ", "N",
	)
	return r.Replace(s)
}

func getLogoData(logoURL string) ([]byte, string) {
	if logoURL == "" {
		return nil, ""
	}

	// Base64
	if strings.HasPrefix(logoURL, "data:image/") {
		parts := strings.SplitN(logoURL, ",", 2)
		if len(parts) != 2 {
			return nil, ""
		}
		var imgType string
		if strings.Contains(parts[0], "png") {
			imgType = "PNG"
		} else if strings.Contains(parts[0], "jpeg") || strings.Contains(parts[0], "jpg") {
			imgType = "JPEG"
		} else if strings.Contains(parts[0], "gif") {
			imgType = "GIF"
		} else {
			return nil, ""
		}
		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, ""
		}
		return data, imgType
	}

	// URL
	return downloadImage(logoURL)
}

func downloadImage(url string) ([]byte, string) {
	if url == "" {
		return nil, ""
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil, ""
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ""
	}

	ct := resp.Header.Get("Content-Type")
	var imgType string
	switch {
	case strings.Contains(ct, "png"):
		imgType = "PNG"
	case strings.Contains(ct, "jpeg"), strings.Contains(ct, "jpg"):
		imgType = "JPEG"
	case strings.Contains(ct, "gif"):
		imgType = "GIF"
	default:
		if len(data) > 4 {
			if data[0] == 0x89 && data[1] == 0x50 {
				imgType = "PNG"
			} else if data[0] == 0xFF && data[1] == 0xD8 {
				imgType = "JPEG"
			}
		}
	}
	return data, imgType
}
