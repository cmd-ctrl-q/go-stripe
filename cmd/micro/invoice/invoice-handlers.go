package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
)

type Order struct {
	ID        int       `json:"id"`
	Quantity  int       `json:"quantity"`
	Amount    int       `json:"amount"`
	Product   string    `json:"product"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (app *application) CreateAndSendInvoice(w http.ResponseWriter, r *http.Request) {
	// receive json
	var order Order

	err := app.readJSON(w, r, &order)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// generate pdf invoice
	err = app.createInvoicePDF(order)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// create email attachment
	// attachments := []string{
	// 	fmt.Sprintf("./invoices/%d.pdf", order.ID),
	// }
	attachments := make(map[string]string)
	attachments[strconv.Itoa(order.ID)] = fmt.Sprintf("./invoices/%d.pdf", order.ID)

	// send mail with attachment
	err = app.SendMail("info@widgets.com", order.Email, "Your invoice", "invoice", attachments, nil)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// send response
	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = fmt.Sprintf("Invoice %d.pdf created and sent to %s", order.ID, order.Email)
	app.writeJSON(w, http.StatusCreated, resp)
}

func (app *application) createInvoicePDF(order Order) error {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10)
	pdf.SetAutoPageBreak(true, 0)

	importer := gofpdi.NewImporter()

	t := importer.ImportPage(pdf, "./pdf-templates/invoice.pdf", 1, "/MediaBox")

	pdf.AddPage()
	importer.UseImportedTemplate(pdf, t, 0, 0, 215.9, 0)

	// write info
	// move down 50mm and right 10mm
	pdf.SetY(50)
	pdf.SetX(10)
	pdf.SetFont("Times", "", 11)
	pdf.CellFormat(97, 8, fmt.Sprintf("Attention %s %s", order.FirstName, order.LastName), "", 0, "L", false, 0, "")
	pdf.Ln(5) // blank line
	pdf.CellFormat(97, 8, order.Email, "", 0, "L", false, 0, "")
	pdf.CellFormat(97, 8, order.CreatedAt.Format("2006-01-02"), "", 0, "L", false, 0, "")

	// if more than one item, use for loop to range over the below
	// populate product info in table
	pdf.SetX(58)
	pdf.SetY(93)
	pdf.CellFormat(155, 8, order.Product, "", 0, "L", false, 0, "")
	pdf.SetX(166)
	pdf.CellFormat(20, 8, fmt.Sprintf("%d", order.Quantity), "", 0, "L", false, 0, "")
	pdf.SetX(185)
	app.infoLog.Println("Amount: ", order.Amount)
	pdf.CellFormat(20, 8, fmt.Sprintf("$%.2f", float32(order.Amount/100.0)), "", 0, "R", false, 0, "")

	// export as pdf
	invoicePath := fmt.Sprintf("./invoices/%d.pdf", order.ID)
	err := pdf.OutputFileAndClose(invoicePath)
	if err != nil {
		return err
	}

	return nil
}
