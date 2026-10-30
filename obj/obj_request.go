package obj

// wrapper untuk standarisasi request
// ==================================

// 🔽 list struct untuk resembles semua request data masuk mulai dari sini
type ReqBukaSimpanan struct {
	KDCAB   string  `json:"kdcab" gorm:"column:kdcab"`
	CNO     string  `json:"cno" gorm:"column:cno"`
	JENIS   string  `json:"jenis" gorm:"column:jenis"`
	NLSETOR float64 `json:"nlsetor" gorm:"column:nlsetor"`
}

type ReqAnggotaBaruDetail struct {
	KDCAB     string `json:"kdcab" gorm:"column:kdcab"`         // mandatory
	CNO       string `json:"cno" gorm:"column:cno"`             // mandatory
	TIPENAS   string `json:"Tipenas" gorm:"column:Tipenas"`     // mandatory
	JNSNAS    string `json:"jnsnas" gorm:"column:jnsnas"`       // mandatory
	CNM       string `json:"cnm" gorm:"column:cnm"`             // mandatory
	TGLLHR    string `json:"tgllhr" gorm:"column:tgllhr"`       // mandatory
	UMUR      int64  `json:"umur" gorm:"column:umur"`           // mandatory
	TMPLHR    string `json:"tmplhr" gorm:"column:tmplhr"`       // mandatory
	JNSKEL    string `json:"jnskel" gorm:"column:jnskel"`       // mandatory
	JNSID     string `json:"JnsID" gorm:"column:JnsID"`         // mandatory
	MARSTAT   string `json:"marstat" gorm:"column:marstat"`     // mandatory
	AGAMA     string `json:"agama" gorm:"column:agama"`         // mandatory
	WARGA     string `json:"warga" gorm:"column:warga"`         // mandatory
	NOID      string `json:"NoID" gorm:"column:NoID"`           // mandatory
	NPWP      string `json:"npwp" gorm:"column:npwp"`           // mandatory
	ALAMATID  string `json:"AlamatID" gorm:"column:AlamatID"`   // mandatory
	KELID     string `json:"KelID" gorm:"column:KelID"`         // mandatory
	KECID     string `json:"KecID" gorm:"column:KecID"`         // mandatory
	KOTAID    string `json:"KotaID" gorm:"column:KotaID"`       // mandatory
	KDPOSID   string `json:"KdposID" gorm:"column:KdposID"`     // mandatory
	ALAMATSKR string `json:"AlamatSkr" gorm:"column:AlamatSkr"` // mandatory
	KELSKR    string `json:"KelSkr" gorm:"column:KelSkr"`       // mandatory
	KECSKR    string `json:"KecSkr" gorm:"column:KecSkr"`       // mandatory
	KOTAISKR  string `json:"KotaISkr" gorm:"column:KotaISkr"`   // mandatory
	KDPOSSKR  string `json:"KdposSkr" gorm:"column:KdposSkr"`   // mandatory
	STSRUMAH  string `json:"stsrumah" gorm:"column:stsrumah"`   // mandatory
	// BENTUKBU  string `json:"BentukBU" gorm:"column:BentukBU"`   // mandatory
	// BIDUSAHA  string `json:"Bidusaha" gorm:"column:Bidusaha"`   // mandatory
	// NOAKTE    string `json:"Noakte" gorm:"column:Noakte"`
	// DOCLAIN   string `json:"Doclain" gorm:"column:Doclain"`
	// NODOCLAIN string `json:"Nodoclain" gorm:"column:Nodoclain"`
	// OMZUSAHA  string `json:"Omzusaha" gorm:"column:Omzusaha"` // mandatory
	NOHP     string `json:"nohp" gorm:"column:nohp"`         // mandatory
	EMAIL    string `json:"email" gorm:"column:email"`       // mandatory
	NMIBU    string `json:"nmibu" gorm:"column:nmibu"`       // mandatory
	STSPOKOK string `json:"stspokok" gorm:"column:stspokok"` // mandatory
	STSWAJIB string `json:"stswajib" gorm:"column:stswajib"` // mandatory

	// kebutuhan login aplikasi mobile koperasi
	PIN string `json:"pin" gorm:"pin"`
}

type ReqAnggotaBaru struct {
	ANGGOTA           ReqAnggotaBaruDetail `json:"anggota"`
	SIMPANAN_POKOK    ReqBukaSimpanan      `json:"simpanan_pokok"`
	SIMPANAN_WAJIB    ReqBukaSimpanan      `json:"simpanan_wajib"`
	SIMPANAN_SUKARELA ReqBukaSimpanan      `json:"simpanan_sukarela"`
}

type ReqDaftarAnggota struct {
	KDCAB string `json:"kdcab" gorm:"column:kdcab"`
}

type ReqDataAnggota struct {
	KDCAB string `json:"kdcab" gorm:"column:kdcab"`
	CNO   string `json:"cno" gorm:"column:cno"`
}

type ReqDataRekening struct {
	KDCAB string `json:"kdcab" gorm:"column:kdcab"`
	NOAC  string `json:"noac" gorm:"column:noac"`
}

type ReqSetorSimpanan struct {
	KDCAB      string  `json:"kdcab"`
	NOAC       string  `json:"noac"`
	NOMINAL    float64 `json:"nominal"`
	KETERANGAN string  `json:"keterangan"`
}

type ReqDaftarRekening struct {
	KDCAB string `json:"kdcab"`
	CNO   string `json:"cno"`
}

type ReqTarikSimpanan struct {
	KDCAB      string  `json:"kdcab"`
	NOAC       string  `json:"noac"`
	NOMINAL    float64 `json:"nominal"`
	KETERANGAN string  `json:"keterangan"`
}

type ReqTransferSimpanan struct {
	NOAC_ASAL   string  `json:"noac_asal"`
	NOAC_TUJUAN string  `json:"noac_tujuan"`
	NOMINAL     float64 `json:"nominal"`
	KETERANGAN  string  `json:"keterangan"`
}

type ReqGetMutasi struct {
	NOAC      string `json:"noac"`
	TGL_AWAL  string `json:"tgl_awal"`
	TGL_AKHIR string `json:"tgl_akhir"`
}

type ReqTutupRekening struct {
	KDCAB       string `json:"kdcab"`
	NOAC        string `json:"noac"`
	NOAC_TUJUAN string `json:"noac_tujuan"`
	ALASAN      string `json:"alasan"`
}

type ReqPinjamanBaru struct {
	KDCAB       string  `json:"kdcab"`
	CNO         string  `json:"cno"`
	JENIS       string  `json:"jenis"`
	PLF_DEBITUR float64 `json:"plf_debitur"`
	JK_WKTAWAL  int64   `json:"jk_wktawal"`
	BUNGA       float64 `json:"bunga"`
	RMSBUNGA    string  `json:"rmsbunga"`
	JNSANGSUR   string  `json:"jnsangsur"`
	JNSKREDIT   string  `json:"jnskredit"`
	TGL_AWAL    string  `json:"tgl_awal"`
	TGLANGSUR   int     `json:"tglangsur"`
	PROVISI     float64 `json:"provisi"`
	ADM         float64 `json:"adm"`
	NOAC        string  `json:"noac"`
	KODEAO      string  `json:"kodeao"`
	NOREF       string  `json:"noref"`
	KETERANGAN  string  `json:"keterangan"`
	NLJAMINAN   float64 `json:"nljaminan"`
}

type ReqGetJadwalAngsuran struct {
	KDCAB string `json:"kdcab"`
	NOFAS string `json:"nofas"`
}

type ReqDaftarPinjaman struct {
	KDCAB string `json:"kdcab"`
	CNO   string `json:"cno"`
}

type ReqDataPinjaman struct {
	KDCAB string `json:"kdcab"`
	NOFAS string `json:"nofas"`
}

type ReqTagihanPinjaman struct {
	KDCAB string `json:"kdcab"`
	NOFAS string `json:"nofas"`
}

type ReqSimulasiPinjaman struct {
	PLF_DEBITUR float64 `json:"plf_debitur"`
	JK_WKTAWAL  int64   `json:"jk_wktawal"`
	BUNGA       float64 `json:"bunga"`
	RMSBUNGA    string  `json:"rmsbunga"`
	JNSANGSUR   string  `json:"jnsangsur"`
	TGL_AWAL    string  `json:"tgl_awal"`
	TGLANGSUR   int     `json:"tglangsur"`
}

type ReqBayarAngsuran struct {
	KDCAB      string  `json:"kdcab"`
	NOFAS      string  `json:"nofas"`
	ANGSURKE   int     `json:"angsurke"`
	CRBAYAR    string  `json:"crbayar"`
	NOAC_DEBET string  `json:"noac_debet"`
	NOMINAL    float64 `json:"nominal"`
	KETERANGAN string  `json:"keterangan"`
}

type ReqPelunasanKredit struct {
	KDCAB      string `json:"kdcab"`
	NOFAS      string `json:"nofas"`
	CRBAYAR    string `json:"crbayar"`
	NOAC_DEBET string `json:"noac_debet"`
	KETERANGAN string `json:"keterangan"`
}

type ReqRiwayatBayar struct {
	KDCAB string `json:"kdcab"`
	NOFAS string `json:"nofas"`
}

type ReqLoginAnggota struct {
	KDCAB string `json:"kdcab"`
	CNO   string `json:"cno"`
	PIN   string `json:"pin"`
}

type ReqRefreshSessionAnggota struct {
	KDCAB         string `json:"kdcab"`
	CNO           string `json:"cno"`
	REFRESH_TOKEN string `json:"refresh_token"`
}
