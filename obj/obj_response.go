package obj

// 🔽 list struct untuk resembles semua response data keluar mulai dari sini
type ResGenOption struct {
	LABEL string `json:"label" gorm:"column:label"`
	VALUE string `json:"value" gorm:"column:value"`
}

type ResListGenOptions []ResGenOption
