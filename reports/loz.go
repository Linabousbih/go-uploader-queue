package reports

import (
	"async/store"
	"context"
)

// Everything in here is just dummy

type Loz struct {
	name string
}
type LozResponse struct {
	Data []string
}

func (l *Loz) GenerateReport(ctx context.Context, report *store.Report) (*LozResponse, error) {
	return &LozResponse{}, nil
}

func (l *Loz) GetMonsters() (*LozResponse, error) {
	return &LozResponse{}, nil
}
