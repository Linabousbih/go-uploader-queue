package reports

import (
	"context"

	"async/models"
)

// Everything in here is just dummy

type Loz struct {
	Name string
}
type LozResponse struct {
	Data []string
}

func (l *Loz) GenerateReport(ctx context.Context, report *models.Report) (*LozResponse, error) {
	return &LozResponse{}, nil
}

func (l *Loz) GetMonsters() (*LozResponse, error) {
	return &LozResponse{}, nil
}
