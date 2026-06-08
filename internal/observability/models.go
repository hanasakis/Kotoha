package observability

import "time"

type EvalRun struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Category   string    `gorm:"size:50;index;not null" json:"category"`
	TotalTests int       `gorm:"not null;default:0" json:"total_tests"`
	Passed     int       `gorm:"not null;default:0" json:"passed"`
	Failed     int       `gorm:"not null;default:0" json:"failed"`
	Accuracy   float64   `gorm:"not null;default:0" json:"accuracy"`
	Metrics    string    `gorm:"type:text" json:"metrics"` // JSON blob for detailed metrics
	CreatedAt  time.Time `json:"created_at"`
}
