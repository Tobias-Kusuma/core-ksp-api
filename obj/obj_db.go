package obj

import "time"

// 🔽 list struct untuk resembles full returned object dari db
type LOG_HTTP struct {
	LOGID            int64      `json:"-" gorm:"column:logid;type:bigint;primaryKey;autoIncrement"`
	REQUEST_TIME     *time.Time `json:"request_time" gorm:"column:request_time"`
	REQUEST_ENDPOINT *string    `json:"request_endpoint" gorm:"column:request_endpoint"`
	REQUEST_TEXT     *string    `json:"request_text" gorm:"column:request_text"`
	RESPONSE_TIME    *time.Time `json:"response_time" gorm:"column:response_time"`
	RESPONSE_CODE    *string    `json:"response_code" gorm:"column:response_code"`
	RESPONSE_TEXT    *string    `json:"response_text" gorm:"column:response_text"`
	CLIENT_IP        *string    `json:"client_ip" gorm:"column:client_ip"`
}

type MSTCST struct {
	KDCAB      string       `json:"kdcab" gorm:"column:kdcab"`         // mandatory
	CNO        string       `json:"cno" gorm:"column:cno"`             // mandatory
	TIPENAS    *string      `json:"Tipenas" gorm:"column:Tipenas"`     // mandatory
	JNSNAS     *string      `json:"jnsnas" gorm:"column:jnsnas"`       // mandatory
	CNM        *string      `json:"cnm" gorm:"column:cnm"`             // mandatory
	TGLLHR     *time.Time   `json:"tgllhr" gorm:"column:tgllhr"`       // mandatory
	UMUR       *float64     `json:"umur" gorm:"column:umur"`           // mandatory
	TMPLHR     *string      `json:"tmplhr" gorm:"column:tmplhr"`       // mandatory
	JNSKEL     *string      `json:"jnskel" gorm:"column:jnskel"`       // mandatory
	JNSID      *string      `json:"JnsID" gorm:"column:JnsID"`         // mandatory
	MARSTAT    *string      `json:"marstat" gorm:"column:marstat"`     // mandatory
	AGAMA      *string      `json:"agama" gorm:"column:agama"`         // mandatory
	WARGA      *string      `json:"warga" gorm:"column:warga"`         // mandatory
	NOID       *string      `json:"NoID" gorm:"column:NoID"`           // mandatory
	NPWP       *string      `json:"npwp" gorm:"column:npwp"`           // mandatory
	ALAMATID   *string      `json:"AlamatID" gorm:"column:AlamatID"`   // mandatory
	KELID      *string      `json:"KelID" gorm:"column:KelID"`         // mandatory
	KECID      *string      `json:"KecID" gorm:"column:KecID"`         // mandatory
	KOTAID     *string      `json:"KotaID" gorm:"column:KotaID"`       // mandatory
	KDPOSID    *string      `json:"KdposID" gorm:"column:KdposID"`     // mandatory
	ALAMATSKR  *string      `json:"AlamatSkr" gorm:"column:AlamatSkr"` // mandatory
	KELSKR     *string      `json:"KelSkr" gorm:"column:KelSkr"`       // mandatory
	KECSKR     *string      `json:"KecSkr" gorm:"column:KecSkr"`       // mandatory
	KOTAISKR   *string      `json:"KotaISkr" gorm:"column:KotaISkr"`   // mandatory
	KDPOSSKR   *string      `json:"KdposSkr" gorm:"column:KdposSkr"`   // mandatory
	STSRUMAH   *string      `json:"stsrumah" gorm:"column:stsrumah"`   // mandatory
	BENTUKBU   *string      `json:"BentukBU" gorm:"column:BentukBU"`   // mandatory
	BIDUSAHA   *string      `json:"Bidusaha" gorm:"column:Bidusaha"`   // mandatory
	NOAKTE     *string      `json:"Noakte" gorm:"column:Noakte"`
	DOCLAIN    *string      `json:"Doclain" gorm:"column:Doclain"`
	NODOCLAIN  *string      `json:"Nodoclain" gorm:"column:Nodoclain"`
	OMZUSAHA   *string      `json:"Omzusaha" gorm:"column:Omzusaha"` // mandatory
	NOHP       *string      `json:"nohp" gorm:"column:nohp"`         // mandatory
	EMAIL      *string      `json:"email" gorm:"column:email"`       // mandatory
	NMIBU      *string      `json:"nmibu" gorm:"column:nmibu"`       // mandatory
	NASABAHPIC *interface{} `json:"Nasabahpic" gorm:"column:Nasabahpic"`
	NASABAHTTD *interface{} `json:"Nasabahttd" gorm:"column:Nasabahttd"`
	STSPOKOK   *string      `json:"stspokok" gorm:"column:stspokok"` // mandatory
	STSWAJIB   *string      `json:"stswajib" gorm:"column:stswajib"` // mandatory
	STATUS     *string      `json:"status" gorm:"column:status"`     // mandatory
	TGLINP     *time.Time   `json:"tglinp" gorm:"column:tglinp"`     // mandatory
	USRDEL     *string      `json:"usrdel" gorm:"column:usrdel"`
	TGLDEL     *time.Time   `json:"tgldel" gorm:"column:tgldel"`
	USRID      *string      `json:"usrid" gorm:"column:usrid"` // mandatory
	DUPD       *time.Time   `json:"dupd" gorm:"column:dupd"`   // mandatory
	OTR        *string      `json:"otr" gorm:"column:otr"`
	DOTR       *time.Time   `json:"dotr" gorm:"column:dotr"`
}

type SHDBLC struct {
	KDCAB     string     `json:"kdcab" gorm:"column:kdcab"`         // mandatory
	NOAC      string     `json:"noac" gorm:"column:noac"`           // mandatory
	CNO       string     `json:"cno" gorm:"column:cno"`             // mandatory
	ALIASNM   *string    `json:"aliasnm" gorm:"column:aliasnm"`     // mandatory
	JENIS     *string    `json:"jenis" gorm:"column:jenis"`         // mandatory
	BLOKIR    *string    `json:"blokir" gorm:"column:blokir"`       // mandatory
	NLHOLD    *float64   `json:"nlhold" gorm:"column:nlhold"`       // mandatory
	TGLHOLD   *time.Time `json:"tglhold" gorm:"column:tglhold"`     // mandatory
	TEMPOHOLD *time.Time `json:"tempohold" gorm:"column:tempohold"` // mandatory
	SLDAWAL   *float64   `json:"sldawal" gorm:"column:sldawal"`     // mandatory
	SLDAKHIR  *float64   `json:"sldakhir" gorm:"column:sldakhir"`   // mandatory
	SLDPAKAI  *float64   `json:"sldpakai" gorm:"column:sldpakai"`   // mandatory
	MINSLD    *float64   `json:"minsld" gorm:"column:minsld"`
	MAXTARIK  *float64   `json:"maxtarik" gorm:"column:maxtarik"`
	NLBIAYA   *float64   `json:"NlBiaya" gorm:"column:NlBiaya"` // mandatory
	MUTDB     *float64   `json:"mutdb" gorm:"column:mutdb"`     // mandatory
	MUTCR     *float64   `json:"mutcr" gorm:"column:mutcr"`     // mandatory
	NLSETOR   *float64   `json:"Nlsetor" gorm:"column:Nlsetor"` // mandatory
	TGLAWAL   *time.Time `json:"Tglawal" gorm:"column:Tglawal"`
	TGLTRN    *time.Time `json:"tgltrn" gorm:"column:tgltrn"`
	JTTEMPO   *time.Time `json:"Jttempo" gorm:"column:Jttempo"`
	TGLTUTUP  *time.Time `json:"tgltutup" gorm:"column:tgltutup"` // mandatory
	TGLKREDIT *float64   `json:"tglkredit" gorm:"column:tglkredit"`
	JKWKT     *float64   `json:"jkwkt" gorm:"column:jkwkt"`
	NLPINALTY *float64   `json:"nlpinalty" gorm:"column:nlpinalty"`
	NOREKCR   *string    `json:"NorekCr" gorm:"column:NorekCr"`
	AUTODB    *string    `json:"AutoDb" gorm:"column:AutoDb"`
	KODEAO    *string    `json:"KodeAO" gorm:"column:KodeAO"` // mandatory
	NOREFBLK  *string    `json:"NorefBlk" gorm:"column:NorefBlk"`
	STATUS    *string    `json:"Status" gorm:"column:Status"` // mandatory
}

type RET500 struct {
	JENIS      string     `json:"jenis" gorm:"column:jenis"`       // mandatory
	NMPRODUK   *string    `json:"nmproduk" gorm:"column:nmproduk"` // mandatory
	JNSPRD     *string    `json:"jnsprd" gorm:"column:jnsprd"`     // mandatory
	JNSTIPE    *string    `json:"jnstipe" gorm:"column:jnstipe"`   // mandatory
	NLBIAYA    *float64   `json:"nlbiaya" gorm:"column:nlbiaya"`
	MINSETOR   *float64   `json:"minsetor" gorm:"column:minsetor"`
	WJBSETOR   *float64   `json:"WjbSetor" gorm:"column:WjbSetor"`
	MAXPAJAK   *float64   `json:"maxpajak" gorm:"column:maxpajak"`
	MINSLDBNG  *float64   `json:"minsldbng" gorm:"column:minsldbng"`
	DAYBNG     *float64   `json:"daybng" gorm:"column:daybng"`
	DBGL       *string    `json:"dbgl" gorm:"column:dbgl"`
	CRGL       *string    `json:"crgl" gorm:"column:crgl"`
	TCDBNG     *string    `json:"tcdbng" gorm:"column:tcdbng"`   // mandatory
	TCDADM     *string    `json:"tcdadm" gorm:"column:tcdadm"`   // mandatory
	TCDBCAD    *string    `json:"tcdbcad" gorm:"column:tcdbcad"` // mandatory
	TCDPEN     *string    `json:"tcdpen" gorm:"column:tcdpen"`
	TCDPAJAK   *string    `json:"tcdpajak" gorm:"column:tcdpajak"` // mandatory
	PENALTY    *float64   `json:"penalty" gorm:"column:penalty"`   // mandatory
	MAXTARIK   *float64   `json:"maxtarik" gorm:"column:maxtarik"` // mandatory
	JNSBNG     *string    `json:"jnsbng" gorm:"column:jnsbng"`
	JMLHRBNG   *float64   `json:"jmlhrbng" gorm:"column:jmlhrbng"`
	SKBNG1     *float64   `json:"skbng1" gorm:"column:skbng1"`
	SKBNG2     *float64   `json:"skbng2" gorm:"column:skbng2"`
	SKBNG3     *float64   `json:"skbng3" gorm:"column:skbng3"`
	SKBNG4     *float64   `json:"skbng4" gorm:"column:skbng4"`
	SKBNG5     *float64   `json:"skbng5" gorm:"column:skbng5"`
	SKBNG6     *float64   `json:"skbng6" gorm:"column:skbng6"`
	SKBNG7     *float64   `json:"skbng7" gorm:"column:skbng7"`
	SKBNG8     *float64   `json:"skbng8" gorm:"column:skbng8"`
	SALDO1     *float64   `json:"saldo1" gorm:"column:saldo1"`
	SALDO2     *float64   `json:"saldo2" gorm:"column:saldo2"`
	SALDO3     *float64   `json:"saldo3" gorm:"column:saldo3"`
	SALDO4     *float64   `json:"saldo4" gorm:"column:saldo4"`
	SALDO5     *float64   `json:"saldo5" gorm:"column:saldo5"`
	SALDO6     *float64   `json:"saldo6" gorm:"column:saldo6"`
	SALDO7     *float64   `json:"saldo7" gorm:"column:saldo7"`
	SALDO8     *float64   `json:"saldo8" gorm:"column:saldo8"`
	PAJAK      *float64   `json:"pajak" gorm:"column:pajak"`           // mandatory
	MINSLD     *float64   `json:"minsld" gorm:"column:minsld"`         // mandatory
	BIAYATUTUP *float64   `json:"biayatutup" gorm:"column:biayatutup"` // mandatory
	HRNAKTIF   *float64   `json:"hrnaktif" gorm:"column:hrnaktif"`
	USRID      *string    `json:"usrid" gorm:"column:usrid"` // mandatory
	DUPD       *time.Time `json:"dupd" gorm:"column:dupd"`   // mandatory
	OTR        *string    `json:"otr" gorm:"column:otr"`
	DOTR       *time.Time `json:"dotr" gorm:"column:dotr"`
	STATUS     *string    `json:"status" gorm:"column:status"` // mandatory
}
