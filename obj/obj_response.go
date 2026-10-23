package obj

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
