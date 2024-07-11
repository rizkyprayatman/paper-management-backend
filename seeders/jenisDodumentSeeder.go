package seeders

import (
	"paper-management-backend/database"
	"paper-management-backend/models"
)

func SeedJenisDocument() {
	jenisDocuments := []models.JenisDocument{
		{Name: "Laporan Keuangan", Description: "Neraca, laporan laba rugi, laporan arus kas"},
		{Name: "Invoice", Description: "Penagihan kepada pelanggan"},
		{Name: "Kwitansi", Description: "Bukti pembayaran"},
		{Name: "Faktur Pajak", Description: "Dokumen perpajakan"},
		{Name: "Surat Izin Usaha", Description: "SIUP, TDP"},
		{Name: "Akta Pendirian", Description: "Akta pendirian perusahaan"},
		{Name: "NPWP", Description: "Nomor Pokok Wajib Pajak"},
		{Name: "Surat Jalan", Description: "Dokumen pengiriman barang"},
		{Name: "Purchase Order", Description: "Pesanan pembelian"},
		{Name: "Delivery Order", Description: "Dokumen pengiriman"},
		{Name: "Nota Pembelian", Description: "Bukti pembelian"},
		{Name: "Nota Penjualan", Description: "Bukti penjualan"},
		{Name: "Kontrak Kerja", Description: "Perjanjian kerja karyawan"},
		{Name: "Daftar Gaji", Description: "Dokumen penggajian"},
		{Name: "Surat Keputusan", Description: "SK pengangkatan karyawan"},
		{Name: "Proposal Penawaran", Description: "Proposal untuk calon pelanggan"},
		{Name: "Surat Penawaran", Description: "Surat penawaran produk/jasa"},
		{Name: "Rencana Pemasaran", Description: "Dokumen strategi pemasaran"},
	}

	for _, jenisDocument := range jenisDocuments {
		database.DB.FirstOrCreate(&jenisDocument, models.JenisDocument{Name: jenisDocument.Name})
	}
}
