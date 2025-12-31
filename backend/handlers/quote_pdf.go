package handlers

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Validate ownership e buscar orçamento
	quote, err := helpers.ValidateQuoteOwnership(c, objID)
	if err != nil {
		return err
	}

	// Buscar template se existir
	var template *models.QuoteTemplate
	if !quote.TemplateID.IsZero() {
		templateCollection := database.GetCollection("quote_templates")
		var t models.QuoteTemplate
		err = templateCollection.FindOne(ctx, bson.M{"_id": quote.TemplateID}).Decode(&t)
		if err == nil {
			template = &t
		}
	}

	// Buscar empresa para dados do cabeçalho
	companyCollection := database.GetCollection("companies")
	var company models.Company
	err = companyCollection.FindOne(ctx, bson.M{"_id": quote.CompanyID}).Decode(&company)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar dados da empresa"})
	}

	// Criar PDF
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Cor primária do template ou padrão
	primaryColor := "#78716c"
	if template != nil && template.PrimaryColor != "" {
		primaryColor = template.PrimaryColor
	}

	// Converter hex para RGB
	r, g, b := hexToRGB(primaryColor)

	// ===== CABEÇALHO =====
	pdf.SetFillColor(r, g, b)
	pdf.Rect(0, 0, 210, 35, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetXY(15, 12)
	pdf.Cell(0, 10, company.Name)

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(15, 22)
	if template != nil && template.HeaderText != "" {
		pdf.Cell(0, 5, template.HeaderText)
	}

	// Número do orçamento
	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(140, 12)
	pdf.Cell(0, 10, quote.Number)

	// ===== DADOS DO ORÇAMENTO =====
	pdf.SetTextColor(0, 0, 0)
	pdf.SetY(45)

	// Título
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, quote.Title)
	pdf.Ln(12)

	// Data e Validade
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(90, 6, fmt.Sprintf("Data: %s", formatDate(quote.CreatedAt)))
	pdf.Cell(0, 6, fmt.Sprintf("Validade: %s", formatDate(quote.ValidUntil)))
	pdf.Ln(10)

	// ===== DADOS DO CLIENTE =====
	pdf.SetFillColor(245, 245, 245)
	pdf.Rect(15, pdf.GetY(), 180, 35, "F")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetX(20)
	pdf.Cell(0, 8, "Dados do Cliente")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetX(20)
	pdf.Cell(0, 5, quote.ClientName)
	pdf.Ln(5)

	if quote.ClientDocument != "" {
		pdf.SetX(20)
		pdf.Cell(0, 5, fmt.Sprintf("CPF/CNPJ: %s", quote.ClientDocument))
		pdf.Ln(5)
	}

	if quote.ClientEmail != "" || quote.ClientPhone != "" {
		pdf.SetX(20)
		contactInfo := ""
		if quote.ClientEmail != "" {
			contactInfo = quote.ClientEmail
		}
		if quote.ClientPhone != "" {
			if contactInfo != "" {
				contactInfo += " | "
			}
			contactInfo += quote.ClientPhone
		}
		pdf.Cell(0, 5, contactInfo)
		pdf.Ln(5)
	}

	if quote.ClientAddress != "" {
		pdf.SetX(20)
		address := quote.ClientAddress
		if quote.ClientCity != "" {
			address += " - " + quote.ClientCity
		}
		if quote.ClientState != "" {
			address += "/" + quote.ClientState
		}
		if quote.ClientZipCode != "" {
			address += " - " + quote.ClientZipCode
		}
		pdf.Cell(0, 5, address)
	}

	pdf.Ln(15)

	// ===== DESCRIÇÃO =====
	if quote.Description != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 8, "Descricao")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 5, quote.Description, "", "", false)
		pdf.Ln(5)
	}

	// ===== TABELA DE ITENS =====
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 8, "Itens do Orcamento")
	pdf.Ln(10)

	// Cabeçalho da tabela
	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(90, 8, "Descricao", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 8, "Qtd", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Valor Unit.", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 8, "Total", "1", 1, "R", true, 0, "")

	// Linhas da tabela
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)
	alternate := false

	for _, item := range quote.Items {
		if alternate {
			pdf.SetFillColor(250, 250, 250)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		pdf.CellFormat(90, 7, truncateString(item.Description, 45), "1", 0, "L", true, 0, "")
		pdf.CellFormat(25, 7, fmt.Sprintf("%.2f", item.Quantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 7, formatCurrency(item.UnitPrice), "1", 0, "R", true, 0, "")
		pdf.CellFormat(35, 7, formatCurrency(item.Total), "1", 1, "R", true, 0, "")

		alternate = !alternate
	}

	pdf.Ln(5)

	// ===== TOTAIS =====
	pdf.SetX(115)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(35, 6, "Subtotal:")
	pdf.Cell(30, 6, formatCurrency(quote.Subtotal))
	pdf.Ln(6)

	if quote.Discount > 0 {
		pdf.SetX(115)
		discountLabel := "Desconto:"
		if quote.DiscountType == models.DiscountPercent {
			discountLabel = fmt.Sprintf("Desconto (%.0f%%):", quote.Discount)
		}
		pdf.Cell(35, 6, discountLabel)
		discountValue := quote.Discount
		if quote.DiscountType == models.DiscountPercent {
			discountValue = quote.Subtotal * quote.Discount / 100
		}
		pdf.Cell(30, 6, fmt.Sprintf("- %s", formatCurrency(discountValue)))
		pdf.Ln(6)
	}

	pdf.SetX(115)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(35, 8, "TOTAL:", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 8, formatCurrency(quote.Total), "1", 1, "R", true, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(10)

	// ===== OBSERVAÇÕES =====
	if quote.Notes != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 8, "Observacoes")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 5, quote.Notes, "", "", false)
		pdf.Ln(5)
	}

	// ===== TERMOS E CONDIÇÕES =====
	if template != nil && template.TermsText != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 8, "Termos e Condicoes")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(80, 80, 80)
		pdf.MultiCell(0, 4, template.TermsText, "", "", false)
		pdf.Ln(5)
	}

	// ===== RODAPÉ =====
	pdf.SetY(-30)
	pdf.SetTextColor(100, 100, 100)
	pdf.SetFont("Arial", "", 8)

	if template != nil && template.FooterText != "" {
		pdf.Cell(0, 5, template.FooterText)
		pdf.Ln(5)
	}

	pdf.Cell(0, 5, fmt.Sprintf("Documento gerado em %s", formatDate(time.Now())))

	// Gerar buffer do PDF
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao gerar PDF"})
	}

	// Enviar PDF
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.pdf\"", quote.Number))

	return c.Send(buf.Bytes())
}

// hexToRGB converte cor hex para RGB
func hexToRGB(hex string) (int, int, int) {
	if len(hex) == 7 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 120, 113, 108 // stone-500 default
	}

	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// truncateString trunca string para tamanho máximo
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
