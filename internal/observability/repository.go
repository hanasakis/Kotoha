package observability

import "gorm.io/gorm"

type EvalRepository struct {
	DB *gorm.DB
}

func NewEvalRepository(db *gorm.DB) *EvalRepository {
	return &EvalRepository{DB: db}
}

func (r *EvalRepository) Migrate() error {
	return r.DB.AutoMigrate(&EvalRun{})
}

func (r *EvalRepository) SaveRun(run *EvalRun) error {
	return r.DB.Create(run).Error
}

func (r *EvalRepository) ListRuns(category string, limit int) ([]EvalRun, error) {
	var runs []EvalRun
	query := r.DB.Order("created_at DESC")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	err := query.Limit(limit).Find(&runs).Error
	return runs, err
}
