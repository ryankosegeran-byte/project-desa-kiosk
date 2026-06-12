package db

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/internal/models"
)

// SeedLocalData seeds the local SQLite database with test residents and initial active types of letters.
// This is extremely helpful for offline demo and manual verification of the kiosk.
func SeedLocalData(db *DB, desaID string) error {
	ctx := context.Background()

	// 1. Seed Warga if table is empty
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM warga").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		log.Info().Msg("Seeding data warga ke SQLite lokal...")
		wargaRepo := NewWargaRepository(db)

		w1 := &models.Warga{
			ID:              uuid.New().String(),
			NIK:             "3201234567890001",
			RFIDUID:         "1234567890", // Mock RFID UID
			Nama:            "Budi Santoso",
			TempatLahir:     "Bandung",
			TanggalLahir:    "1990-05-12",
			JenisKelamin:    "L",
			Alamat:          "Jl. Merdeka No. 10",
			RT:              "01",
			RW:              "03",
			Kelurahan       : "Mekarjaya",
			Kecamatan       : "Sukasari",
			Kabupaten       : "Bandung",
			Provinsi        : "Jawa Barat",
			Agama           : "Islam",
			StatusKawin     : "Kawin",
			Pekerjaan       : "Wiraswasta",
			Kewarganegaraan : "WNI",
			DesaID          : desaID,
			CreatedAt       : time.Now(),
			UpdatedAt       : time.Now(),
		}
		if err := wargaRepo.Upsert(ctx, w1); err != nil {
			return err
		}

		w2 := &models.Warga{
			ID:              uuid.New().String(),
			NIK:             "3201234567890002",
			RFIDUID:         "0987654321", // Another Mock RFID
			Nama:            "Siti Aminah",
			TempatLahir:     "Surabaya",
			TanggalLahir:    "1995-08-21",
			JenisKelamin:    "P",
			Alamat:          "Jl. Pahlawan No. 45",
			RT:              "02",
			RW:              "04",
			Kelurahan       : "Mekarjaya",
			Kecamatan       : "Sukasari",
			Kabupaten       : "Bandung",
			Provinsi        : "Jawa Barat",
			Agama           : "Islam",
			StatusKawin     : "Belum Kawin",
			Pekerjaan       : "Karyawan Swasta",
			Kewarganegaraan : "WNI",
			DesaID          : desaID,
			CreatedAt       : time.Now(),
			UpdatedAt       : time.Now(),
		}
		if err := wargaRepo.Upsert(ctx, w2); err != nil {
			return err
		}

		log.Info().Msg("Seeding warga selesai.")
	}

	// 2. Seed JenisSurat if empty
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM jenis_surat").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		log.Info().Msg("Seeding jenis_surat ke SQLite lokal...")
		jsRepo := NewJenisSuratRepository(db)

		initialJenisSurat := []struct {
			Kode      string
			Nama      string
			Deskripsi string
			Schema    string
		}{
			{
				Kode:      "SK_DOMISILI",
				Nama:      "Surat Keterangan Domisili",
				Deskripsi: "Surat keterangan tempat tinggal warga",
				Schema:    `{"fields":[{"key":"tujuan","label":"Tujuan Surat","type":"text","required":true},{"key":"berlaku_dari","label":"Berlaku Dari","type":"date","required":false}]}`,
			},
			{
				Kode:      "SKTM",
				Nama:      "Surat Keterangan Kurang Mampu",
				Deskripsi: "Surat keterangan untuk warga kurang mampu ekonomi",
				Schema:    `{"fields":[{"key":"tujuan","label":"Tujuan Surat","type":"text","required":true},{"key":"keperluan","label":"Keperluan / Alasan","type":"textarea","required":true}]}`,
			},
			{
				Kode:      "SK_USAHA",
				Nama:      "Surat Keterangan Usaha",
				Deskripsi: "Surat keterangan kepemilikan usaha warga",
				Schema:    `{"fields":[{"key":"jenis_usaha","label":"Jenis Usaha","type":"select_or_input","required":true,"options":["Perdagangan","Jasa","Pertanian","Peternakan","Perikanan","Industri Kecil","Kerajinan","Bengkel","Warung / Toko","Rumah Makan"],"placeholder":"Pilih atau ketik jenis usaha"},{"key":"merk_usaha","label":"Merk / Nama Usaha","type":"text","required":true,"placeholder":"Contoh: Toko Maju Jaya"},{"key":"tahun_mulai_usaha","label":"Usaha Dimulai Sejak Tahun","type":"number","required":true,"placeholder":"Contoh: 2020"},{"key":"alamat_usaha","label":"Alamat Tempat Usaha","type":"address","required":true,"placeholder":"Klik peta atau ketik alamat"},{"key":"batas_utara","label":"Batas Utara","type":"text","required":false,"placeholder":"Berbatasan dengan..."},{"key":"batas_selatan","label":"Batas Selatan","type":"text","required":false,"placeholder":"Berbatasan dengan..."},{"key":"batas_timur","label":"Batas Timur","type":"text","required":false,"placeholder":"Berbatasan dengan..."},{"key":"batas_barat","label":"Batas Barat","type":"text","required":false,"placeholder":"Berbatasan dengan..."},{"key":"sifat_tempat_usaha","label":"Sifat Tempat Usaha","type":"select","required":true,"options":["Permanen","Sementara","Pinjam"]},{"key":"tahun_kewajiban","label":"Tahun Pelunasan Kewajiban (PBB dll)","type":"number","required":true,"placeholder":"Otomatis tahun berjalan"}]}`,
			},
			{
				Kode:      "SK_KELAHIRAN",
				Nama:      "Surat Keterangan Kelahiran",
				Deskripsi: "Surat keterangan kelahiran anak warga",
				Schema:    `{"fields":[{"key":"nama_bayi","label":"Nama Bayi","type":"text","required":true},{"key":"tanggal_lahir_bayi","label":"Tanggal Lahir Bayi","type":"date","required":true},{"key":"tempat_lahir_bayi","label":"Tempat Lahir Bayi","type":"text","required":true},{"key":"nama_ibu","label":"Nama Lengkap Ibu","type":"text","required":true},{"key":"nama_ayah","label":"Nama Lengkap Ayah","type":"text","required":true}]}`,
			},
			{
				Kode:      "SK_MENINGGAL",
				Nama:      "Surat Keterangan Meninggal Dunia",
				Deskripsi: "Surat keterangan kematian warga",
				Schema:    `{"fields":[{"key":"tanggal_meninggal","label":"Tanggal Meninggal","type":"date","required":true},{"key":"tempat_meninggal","label":"Tempat Meninggal","type":"text","required":true},{"key":"penyebab","label":"Penyebab Kematian","type":"text","required":true},{"key":"nama_pelapor","label":"Nama Pelapor","type":"text","required":true},{"key":"hubungan_pelapor","label":"Hubungan Pelapor dengan Almarhum/ah","type":"select","required":true,"options":["Suami","Istri","Anak","Orang Tua","Saudara","Lainnya"]}]}`,
			},
			{
				Kode:      "SK_BELUM_MENIKAH",
				Nama:      "Surat Keterangan Belum Pernah Menikah",
				Deskripsi: "Surat keterangan status belum pernah menikah",
				Schema:    `{"fields":[{"key":"tujuan","label":"Tujuan Surat","type":"text","required":true}]}`,
			},
			{
				Kode:      "SK_AHLI_WARIS",
				Nama:      "Surat Keterangan Ahli Waris",
				Deskripsi: "Surat keterangan penunjukan ahli waris",
				Schema:    `{"fields":[{"key":"nama_almarhum","label":"Nama Almarhum/ah","type":"text","required":true},{"key":"tanggal_meninggal","label":"Tanggal Meninggal","type":"date","required":true},{"key":"ahli_waris","label":"Daftar Ahli Waris","type":"repeater","required":true,"sub_fields":[{"key":"nama","label":"Nama Lengkap","type":"text"},{"key":"hubungan","label":"Hubungan Ahli Waris","type":"select","options":["Suami","Istri","Anak","Orang Tua","Saudara"]}]}]}`,
			},
			{
				Kode:      "SK_ORANG_SAMA",
				Nama:      "Surat Keterangan Orang Yang Sama",
				Deskripsi: "Surat pernyataan perbedaan nama di dua dokumen berbeda",
				Schema:    `{"fields":[{"key":"nama_lain","label":"Nama Lain (di dokumen lain)","type":"text","required":true},{"key":"dokumen_lain","label":"Dokumen Lain (e.g. Ijazah/Sertifikat)","type":"text","required":true}]}`,
			},
			{
				Kode:      "PENGAKUAN_KAWIN_ADAT",
				Nama:      "Surat Pengakuan Bersama (Kawin Adat)",
				Deskripsi: "Surat keterangan pengakuan kawin secara adat",
				Schema:    `{"fields":[{"key":"nama_pasangan","label":"Nama Pasangan","type":"text","required":true},{"key":"tanggal_perkawinan","label":"Tanggal Perkawinan Adat","type":"date","required":true},{"key":"pemimpin_adat","label":"Pemimpin Adat/Tokoh Adat","type":"text","required":true}]}`,
			},
			{
				Kode:      "IJIN_ORANG_TUA",
				Nama:      "Surat Ijin Orang Tua",
				Deskripsi: "Surat ijin dari orang tua untuk keperluan anak",
				Schema:    `{"fields":[{"key":"nama_anak","label":"Nama Anak","type":"text","required":true},{"key":"nik_anak","label":"NIK Anak","type":"text","required":true},{"key":"keperluan","label":"Keperluan / Ijin Untuk","type":"textarea","required":true}]}`,
			},
			{
				Kode:      "IJIN_KERAMAIAN",
				Nama:      "Surat Permohonan Ijin Keramaian",
				Deskripsi: "Surat pengantar ijin keramaian untuk acara warga",
				Schema:    `{"fields":[{"key":"nama_acara","label":"Nama Acara / Kegiatan","type":"text","required":true},{"key":"tanggal_acara","label":"Tanggal Acara","type":"date","required":true},{"key":"tempat_acara","label":"Tempat Acara","type":"text","required":true},{"key":"jumlah_undangan","label":"Estimasi Jumlah Undangan","type":"number","required":true}]}`,
			},
			{
				Kode:      "PENGANTAR_SKCK",
				Nama:      "Surat Pengantar Pembuatan SKCK",
				Deskripsi: "Surat pengantar untuk pembuatan SKCK di Kepolisian",
				Schema:    `{"fields":[{"key":"keperluan","label":"Keperluan SKCK","type":"text","required":true}]}`,
			},
		}

		for idx, js := range initialJenisSurat {
			jsID := uuid.New().String()
			model := &models.JenisSurat{
				ID:           jsID,
				Kode:         js.Kode,
				Nama:         js.Nama,
				Deskripsi:    js.Deskripsi,
				FieldsSchema: json.RawMessage(js.Schema),
				Aktif:        true,
				Urutan:       idx + 1,
				UpdatedAt:    time.Now(),
			}

			if err := jsRepo.Upsert(ctx, model); err != nil {
				return err
			}

			// Add a simple default template for each jenis_surat
			tID := uuid.New().String()
			templateHTML := `<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
        .header { text-align: center; margin-bottom: 30px; border-bottom: 3px double #000; padding-bottom: 10px; }
        .header h1 { margin: 0; font-size: 20px; text-transform: uppercase; }
        .header p { margin: 5px 0 0 0; font-size: 12px; }
        .title { text-align: center; font-weight: bold; text-decoration: underline; text-transform: uppercase; font-size: 16px; margin-bottom: 20px; }
        .content { margin-bottom: 40px; }
        .footer { float: right; text-align: center; width: 200px; margin-top: 50px; }
        .signature-line { margin-top: 60px; font-weight: bold; border-top: 1px solid #000; display: inline-block; width: 100%; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Pemerintah Kabupaten Bandung</h1>
        <h1>Kecamatan Sukasari - Desa Mekarjaya</h1>
        <p>Alamat Kantor Desa Mekarjaya No. 1, Telp: (022) 123456</p>
    </div>
    
    <div class="title">` + js.Nama + `</div>
    
    <div class="content">
        <p>Yang bertanda tangan di bawah ini Kepala Desa Mekarjaya, Kecamatan Sukasari, Kabupaten Bandung, menerangkan dengan sebenarnya bahwa:</p>
        <table style="width: 100%; margin-left: 20px; margin-bottom: 20px;">
            <tr><td style="width: 150px;">Nama</td><td>: <strong>{{.Warga.Nama}}</strong></td></tr>
            <tr><td>NIK</td><td>: {{.Warga.NIK}}</td></tr>
            <tr><td>Tempat/Tgl Lahir</td><td>: {{.Warga.TempatLahir}}, {{.Warga.TanggalLahir}}</td></tr>
            <tr><td>Jenis Kelamin</td><td>: {{.Warga.JenisKelamin}}</td></tr>
            <tr><td>Pekerjaan</td><td>: {{.Warga.Pekerjaan}}</td></tr>
            <tr><td>Alamat</td><td>: {{.Warga.Alamat}} RT {{.Warga.RT}} / RW {{.Warga.RW}}, Kel. {{.Warga.Kelurahan}}, Kec. {{.Warga.Kecamatan}}</td></tr>
        </table>
        
        <p>Adalah benar warga kami yang berkelakuan baik dan memohon pembuatan surat keterangan dengan detail sebagai berikut:</p>
        <table style="width: 100%; margin-left: 20px; margin-bottom: 20px;">
            {{range $key, $value := .DataSurat}}
            <tr>
                <td style="width: 150px; text-transform: capitalize;">{{$key}}</td>
                <td>: {{$value}}</td>
            </tr>
            {{end}}
        </table>
        
        <p>Demikian surat keterangan ini dibuat untuk dapat dipergunakan sebagaimana mestinya.</p>
    </div>
    
    <div class="footer">
        <p>Mekarjaya, {{.DateToday}}</p>
        <p>Kepala Desa Mekarjaya,</p>
        <div class="signature-line" style="margin-top: 80px;"></div>
        <p><strong>Ujang Hermawan, S.Sos</strong></p>
        <p>NIP. 19750812 200312 1 002</p>
    </div>
</body>
</html>`

			if js.Kode == "SK_USAHA" {
				if content, err := os.ReadFile("kiosk/templates/sku.html"); err == nil {
					templateHTML = string(content)
				} else if content, err := os.ReadFile("templates/sku.html"); err == nil {
					templateHTML = string(content)
				}
			}

			tpl := &models.SuratTemplate{
				ID:           tID,
				JenisSuratID: jsID,
				DesaID:       desaID,
				TemplateHTML: templateHTML,
				Version:      1,
				UpdatedAt:    time.Now(),
			}

			if err := jsRepo.UpsertTemplate(ctx, tpl); err != nil {
				return err
			}
		}

		log.Info().Msg("Seeding jenis_surat selesai.")
	}

	// 3. Seed nomor_surat_batch for each jenis_surat
	nomorRepo := NewNomorSuratRepository(db)
	jsRepo2 := NewJenisSuratRepository(db)
	allJS, _ := jsRepo2.ListAktif(ctx)
	for _, js := range allJS {
		_ = nomorRepo.UpdateBatch(ctx, models.NomorSuratBatch{
			JenisSuratID:  js.ID,
			NomorTerakhir: 0,
			BatasAtas:     100,
			FormatNomor:   "{nomor}/{kode_surat}/{kode_desa}/{bulan_romawi}/{tahun}",
		})
	}
	log.Info().Msg("Seeding nomor_surat_batch selesai.")

	// 4. Seed default desa config
	configRepo := NewConfigRepository(db)
	_ = configRepo.Set(ctx, "desa_kepala_desa", "ALFRIDA, A.Md.Kes")
	_ = configRepo.Set(ctx, "desa_nip", "196902061993032004")
	_ = configRepo.Set(ctx, "kode_desa", "08.10")

	return nil
}
