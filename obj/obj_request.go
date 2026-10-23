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
}

type ReqAnggotaBaru struct {
	ANGGOTA           ReqAnggotaBaruDetail `json:"anggota"`
	SIMPANAN_POKOK    ReqBukaSimpanan      `json:"simpanan_pokok"`
	SIMPANAN_WAJIB    ReqBukaSimpanan      `json:"simpanan_wajib"`
	SIMPANAN_SUKARELA ReqBukaSimpanan      `json:"simpanan_sukarela"`
}
