package obj

import "time"

// 🔽 list struct untuk resembles semua response data keluar mulai dari sini
type ResGenOption struct {
	LABEL string `json:"label" gorm:"column:label"`
	VALUE string `json:"value" gorm:"column:value"`
}

type ResListGenOptions []ResGenOption

type ResDaftarAnggota struct {
	KDCAB   string `json:"kdcab" gorm:"column:kdcab"`
	CNO     string `json:"cno" gorm:"column:cno"`
	TIPENAS string `json:"tipenas" gorm:"column:tipenas"`
	JNSNAS  string `json:"jnsnas" gorm:"column:jnsnas"`
	CNM     string `json:"cnm" gorm:"column:cnm"`
}

type ResDataAnggota struct {
	ANGGOTA         MSTCST   `json:"anggota"`
	DAFTAR_REKENING []SHDBLC `json:"daftar_rekening"`
}

type ResProdukSimpanan struct {
	JENIS      string   `json:"jenis" gorm:"column:jenis"`
	NMPRODUK   string   `json:"nmproduk" gorm:"column:nmproduk"`
	JNSPRD     string   `json:"jnsprd" gorm:"column:jnsprd"`
	JNSTIPE    string   `json:"jnstipe" gorm:"column:jnstipe"`
	NLBIAYA    *float64 `json:"nlbiaya" gorm:"column:nlbiaya"`
	MINSETOR   *float64 `json:"minsetor" gorm:"column:minsetor"`
	WJBSETOR   *float64 `json:"wjbsetor" gorm:"column:wjbsetor"`
	MINSLD     *float64 `json:"minsld" gorm:"column:minsld"`
	MAXTARIK   *float64 `json:"maxtarik" gorm:"column:maxtarik"`
	BIAYATUTUP *float64 `json:"biayatutup" gorm:"column:biayatutup"`
	PENALTY    *float64 `json:"penalty" gorm:"column:penalty"`
	PAJAK      *float64 `json:"pajak" gorm:"column:pajak"`
	STATUS     string   `json:"status" gorm:"column:status"`
}

type ResNewSimpanan struct {
	NAMA string `json:"nama"`
	NOAC string `json:"noac"`
}

type ResDataRekening struct {
	SHDBLC
	NMPRODUK *string `json:"nmproduk" gorm:"column:nmproduk"`
	JNSPRD   *string `json:"jnsprd" gorm:"column:jnsprd"`
	JNSTIPE  *string `json:"jnstipe" gorm:"column:jnstipe"`
}

type ResSetorSimpanan struct {
	KDCAB      string  `json:"kdcab"`
	NOAC       string  `json:"noac"`
	ALIASNM    string  `json:"aliasnm"`
	NOMINAL    float64 `json:"nominal"`
	SLDAKHIR   float64 `json:"sldakhir"`
	SLDPAKAI   float64 `json:"sldpakai"`
	KETERANGAN string  `json:"keterangan"`
}

type ResTarikSimpanan struct {
	KDCAB      string  `json:"kdcab"`
	NOAC       string  `json:"noac"`
	ALIASNM    string  `json:"aliasnm"`
	NOMINAL    float64 `json:"nominal"`
	SLDAKHIR   float64 `json:"sldakhir"`
	SLDPAKAI   float64 `json:"sldpakai"`
	KETERANGAN string  `json:"keterangan"`
}

type ResTransferSimpanan struct {
	NOAC_ASAL      string  `json:"noac_asal"`
	NAMA_PENGIRIM  string  `json:"nama_pengirim"`
	NOAC_TUJUAN    string  `json:"noac_tujuan"`
	NAMA_PENERIMA  string  `json:"nama_penerima"`
	NOMINAL        float64 `json:"nominal"`
	SLDAKHIR_ASAL  float64 `json:"sldakhir_asal"`
	SLDPAKAI_ASAL  float64 `json:"sldpakai_asal"`
	KETERANGAN     string  `json:"keterangan"`
	WAKTU_TRANSFER string  `json:"waktu_transfer"`
}

type ResMutasiItem struct {
	KDCAB       string     `json:"kdcab" gorm:"column:kdcab"`
	NOREF       string     `json:"noref" gorm:"column:noref"`
	NOU         *int64     `json:"nou" gorm:"column:nou"`
	NOAC        string     `json:"noac" gorm:"column:noac"`
	TCD         string     `json:"tcd" gorm:"column:tcd"`
	DOC         string     `json:"doc" gorm:"column:doc"`
	TIPE_MUTASI string     `json:"tipe_mutasi" gorm:"column:tipe_mutasi"`
	NILAI       float64    `json:"nilai" gorm:"column:nilai"`
	KET         string     `json:"ket" gorm:"column:ket"`
	TGLTRN      *time.Time `json:"tgltrn" gorm:"column:tgltrn"`
	NOURUT      *int64     `json:"nourut" gorm:"column:nourut"`
}

type ResTutupRekening struct {
	NOAC           string  `json:"noac"`
	ALIASNM        string  `json:"aliasnm"`
	SALDO_TERAKHIR float64 `json:"saldo_terakhir"`
	BIAYA_TUTUP    float64 `json:"biaya_tutup"`
	DANA_DICAIRKAN float64 `json:"dana_dicairkan"`
	NOAC_TUJUAN    string  `json:"noac_tujuan"`
	TGL_TUTUP      string  `json:"tgl_tutup"`
	STATUS         string  `json:"status"`
}

type ResProdukPinjaman struct {
	KD_PRODUK   string   `json:"kd_produk" gorm:"column:kd_produk"`
	NM_PRODUK   string   `json:"nm_produk" gorm:"column:nm_produk"`
	FLGJMN      string   `json:"flgjmn" gorm:"column:flgjmn"`
	JNSANG      string   `json:"jnsang" gorm:"column:jnsang"`
	RMSBUNGA    string   `json:"rmsbunga" gorm:"column:rmsbunga"`
	BASERATE    *float64 `json:"baserate" gorm:"column:baserate"`
	BULAT       *float64 `json:"bulat" gorm:"column:bulat"`
	JNSPINJAMAN string   `json:"jnspinjaman" gorm:"column:jnspinjaman"`
}

type ResJadwalAngsuran struct {
	ANGSURKE float64 `json:"angsurke"`
	TGLVAL   string  `json:"tglval"`
	POKOKJDW float64 `json:"pokok_jdw"`
	BUNGAJDW float64 `json:"bunga_jdw"`
	TOTAL    float64 `json:"total_angsuran"`
	OSDEBJDW float64 `json:"osdeb_jdw"`
}

type ResPinjamanBaru struct {
	KDCAB       string              `json:"kdcab"`
	CNO         string              `json:"cno"`
	NOFAS       string              `json:"nofas"`
	NAMA        string              `json:"nama"`
	JENIS       string              `json:"jenis"`
	NM_PRODUK   string              `json:"nm_produk"`
	PLF_DEBITUR float64             `json:"plf_debitur"`
	JK_WKTAWAL  int64               `json:"jk_wktawal"`
	BUNGA       float64             `json:"bunga"`
	RMSBUNGA    string              `json:"rmsbunga"`
	JNSANGSUR   string              `json:"jnsangsur"`
	TGL_AWAL    string              `json:"tgl_awal"`
	JT_TEMPO    string              `json:"jt_tempo"`
	TGLANGS     string              `json:"tglangs"`
	ANGSUR_DEB  float64             `json:"angsur_deb"`
	ANGSPOKOK   float64             `json:"angspokok"`
	ANGSBUNGA   float64             `json:"angsbunga"`
	TOTBNG      float64             `json:"totbng"`
	PROVISI     float64             `json:"provisi"`
	ADM         float64             `json:"adm"`
	TOTAL_BIAYA float64             `json:"total_biaya"`
	JADWAL      []ResJadwalAngsuran `json:"jadwal"`
}

type ResDaftarPinjamanItem struct {
	KDCAB       string     `json:"kdcab" gorm:"column:kdcab"`
	CNO         string     `json:"cno" gorm:"column:cno"`
	NOFAS       string     `json:"nofas" gorm:"column:nofas"`
	JENIS       string     `json:"jenis" gorm:"column:jenis"`
	NM_PRODUK   string     `json:"nm_produk" gorm:"column:nm_produk"`
	PLF_DEBITUR float64    `json:"plf_debitur" gorm:"column:plf_debitur"`
	SLDAKHIR    float64    `json:"sldakhir" gorm:"column:sldakhir"`
	TGL_AWAL    *time.Time `json:"tgl_awal" gorm:"column:tgl_awal"`
	JT_TEMPO    *time.Time `json:"jt_tempo" gorm:"column:jt_tempo"`
	ANGSUR_DEB  float64    `json:"angsur_deb" gorm:"column:angsur_deb"`
	KOLEK       string     `json:"kolek" gorm:"column:kolek"`
	STATUS      string     `json:"status" gorm:"column:status"`
	STATUS_KET  string     `json:"status_ket" gorm:"column:status_ket"`
}

type ResDataPinjaman struct {
	KDCAB         string     `json:"kdcab" gorm:"column:kdcab"`
	CNO           string     `json:"cno" gorm:"column:cno"`
	NAMA          string     `json:"nama" gorm:"column:nama"`
	NOFAS         string     `json:"nofas" gorm:"column:nofas"`
	JENIS         string     `json:"jenis" gorm:"column:jenis"`
	NM_PRODUK     string     `json:"nm_produk" gorm:"column:nm_produk"`
	PLF_DEBITUR   float64    `json:"plf_debitur" gorm:"column:plf_debitur"`
	SLDAKHIR      float64    `json:"sldakhir" gorm:"column:sldakhir"`
	JK_WKTAWAL    float64    `json:"jk_wktawal" gorm:"column:jk_wktawal"`
	BUNGA         float64    `json:"bunga" gorm:"column:bunga"`
	RMSBUNGA      string     `json:"rmsbunga" gorm:"column:rmsbunga"`
	RMSBUNGA_KET  string     `json:"rmsbunga_ket" gorm:"column:rmsbunga_ket"`
	JNSANGSUR     string     `json:"jnsangsur" gorm:"column:jnsangsur"`
	JNSANGSUR_KET string     `json:"jnsangsur_ket" gorm:"column:jnsangsur_ket"`
	TGL_AWAL      *time.Time `json:"tgl_awal" gorm:"column:tgl_awal"`
	JT_TEMPO      *time.Time `json:"jt_tempo" gorm:"column:jt_tempo"`
	TGLANGS       *time.Time `json:"tglangs" gorm:"column:tglangs"`
	TGLANGSUR     float64    `json:"tglangsur" gorm:"column:tglangsur"`
	ANGSUR_DEB    float64    `json:"angsur_deb" gorm:"column:angsur_deb"`
	ANGSPOKOK     float64    `json:"angspokok" gorm:"column:angspokok"`
	ANGSBUNGA     float64    `json:"angsbunga" gorm:"column:angsbunga"`
	TOTBNG        float64    `json:"totbng" gorm:"column:totbng"`
	PROVISI       float64    `json:"provisi" gorm:"column:provisi"`
	ADM           float64    `json:"adm" gorm:"column:adm"`
	TOTAL_BIAYA   float64    `json:"total_biaya" gorm:"column:total_biaya"`
	KOLEK         string     `json:"kolek" gorm:"column:kolek"`
	STSKRD        string     `json:"stskrd" gorm:"column:stskrd"`
	STATUS        string     `json:"status" gorm:"column:status"`
	NOAC          string     `json:"noac" gorm:"column:noac"`
	KODEAO        string     `json:"kodeao" gorm:"column:kodeao"`
}

type ResItemTagihan struct {
	ANGSURKE       float64 `json:"angsurke"`
	TGLVAL         string  `json:"tglval"`
	POKOK          float64 `json:"pokok"`
	BUNGA          float64 `json:"bunga"`
	DENDA          float64 `json:"denda"`
	HARI_TERLAMBAT int     `json:"hari_terlambat"`
	TOTAL          float64 `json:"total"`
	STATUS         string  `json:"status"`
}

type ResTagihanPinjaman struct {
	KDCAB               string           `json:"kdcab"`
	NOFAS               string           `json:"nofas"`
	NAMA                string           `json:"nama"`
	JENIS               string           `json:"jenis"`
	NM_PRODUK           string           `json:"nm_produk"`
	TOTAL_POKOK         float64          `json:"total_pokok"`
	TOTAL_BUNGA         float64          `json:"total_bunga"`
	TOTAL_DENDA         float64          `json:"total_denda"`
	GRAND_TOTAL         float64          `json:"grand_total"`
	JUMLAH_ANGSURAN_TGK int              `json:"jumlah_angsuran_tgk"`
	TAGIHAN_DETAIL      []ResItemTagihan `json:"tagihan_detail"`
}

type ResSimulasiPinjaman struct {
	PLF_DEBITUR   float64             `json:"plf_debitur"`
	JK_WKTAWAL    int64               `json:"jk_wktawal"`
	BUNGA         float64             `json:"bunga"`
	RMSBUNGA      string              `json:"rmsbunga"`
	RMSBUNGA_KET  string              `json:"rmsbunga_ket"`
	JNSANGSUR     string              `json:"jnsangsur"`
	JNSANGSUR_KET string              `json:"jnsangsur_ket"`
	TGL_AWAL      string              `json:"tgl_awal"`
	JT_TEMPO      string              `json:"jt_tempo"`
	ANGSUR_DEB    float64             `json:"angsur_deb"`
	ANGSPOKOK     float64             `json:"angspokok"`
	ANGSBUNGA     float64             `json:"angsbunga"`
	TOTBNG        float64             `json:"totbng"`
	TOTAL_BAYAR   float64             `json:"total_bayar"`
	JADWAL        []ResJadwalAngsuran `json:"jadwal"`
}

type ResBayarAngsuran struct {
	NOREF            string  `json:"noref"`
	KDCAB            string  `json:"kdcab"`
	NOFAS            string  `json:"nofas"`
	ANGSURKE         float64 `json:"angsurke"`
	TGLBYR           string  `json:"tglbyr"`
	CRBAYAR          string  `json:"crbayar"`
	NOAC_DEBET       string  `json:"noac_debet"`
	POKOK            float64 `json:"pokok"`
	BUNGA            float64 `json:"bunga"`
	DENDA            float64 `json:"denda"`
	TOTAL_BAYAR      float64 `json:"total_bayar"`
	SISA_OUTSTANDING float64 `json:"sisa_outstanding"`
	STATUS_PINJAMAN  string  `json:"status_pinjaman"`
	KETERANGAN       string  `json:"keterangan"`
}

type ResPelunasanKredit struct {
	NOREF           string  `json:"noref"`
	KDCAB           string  `json:"kdcab"`
	NOFAS           string  `json:"nofas"`
	TGL_LUNAS       string  `json:"tgl_lunas"`
	CRBAYAR         string  `json:"crbayar"`
	NOAC_DEBET      string  `json:"noac_debet"`
	SISA_POKOK      float64 `json:"sisa_pokok"`
	BUNGA           float64 `json:"bunga"`
	DENDA           float64 `json:"denda"`
	PENALTY         float64 `json:"penalty"`
	TOTAL_PELUNASAN float64 `json:"total_pelunasan"`
	STATUS          string  `json:"status"`
	KETERANGAN      string  `json:"keterangan"`
}

type ResRiwayatBayarItem struct {
	NOREF        string  `json:"noref" gorm:"column:noref"`
	TGLTRN       string  `json:"tgltrn" gorm:"column:tgltrn"`
	STSTRN       string  `json:"ststrn" gorm:"column:ststrn"`
	STSTRN_DESC  string  `json:"ststrn_desc"`
	KET          string  `json:"ket" gorm:"column:ket"`
	CRBAYAR      string  `json:"crbayar" gorm:"column:crbayar"`
	CRBAYAR_DESC string  `json:"crbayar_desc"`
	NOACDB       string  `json:"noacdb" gorm:"column:noacdb"`
	ANGSURKE     float64 `json:"angsurke" gorm:"column:angsurke"`
	NLPOKOK      float64 `json:"nlpokok" gorm:"column:nlpokok"`
	NLBUNGA      float64 `json:"nlbunga" gorm:"column:nlbunga"`
	NLDENDA      float64 `json:"nldenda" gorm:"column:nldenda"`
	NLPINALTY    float64 `json:"nlpinalty" gorm:"column:nlpinalty"`
	TOTAL        float64 `json:"total"`
}

