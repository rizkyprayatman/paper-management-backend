package seeders

import (
	"paper-management-backend/database"
	"paper-management-backend/models"
)

func SeedJenisDocument() {
	jenisDocuments := []models.JenisDocument{
		{Name: "Financial Report", Description: "Balance sheet, profit and loss statement, cash flow statement"},
		{Name: "Invoice", Description: "Billing document to customers"},
		{Name: "Receipt", Description: "Proof of payment"},
		{Name: "Tax Invoice", Description: "Tax-related document"},
		{Name: "Business License", Description: "Business licenses such as SIUP, TDP"},
		{Name: "Articles of Incorporation", Description: "Company establishment documents"},
		{Name: "Tax Identification Number (NPWP)", Description: "Taxpayer identification number"},
		{Name: "Delivery Note", Description: "Document for shipment of goods"},
		{Name: "Purchase Order", Description: "Purchase order document"},
		{Name: "Delivery Order", Description: "Delivery document"},
		{Name: "Purchase Receipt", Description: "Proof of purchase"},
		{Name: "Sales Receipt", Description: "Proof of sale"},
		{Name: "Employment Contract", Description: "Employee employment agreement"},
		{Name: "Payroll List", Description: "Payroll documents"},
		{Name: "Appointment Letter", Description: "Appointment letter for employees"},
		{Name: "Quotation Proposal", Description: "Proposal for prospective clients"},
		{Name: "Quotation Letter", Description: "Letter offering products/services"},
		{Name: "Marketing Plan", Description: "Marketing strategy document"},
	}

	for _, jenisDocument := range jenisDocuments {
		database.DB.FirstOrCreate(&jenisDocument, models.JenisDocument{Name: jenisDocument.Name})
	}
}
