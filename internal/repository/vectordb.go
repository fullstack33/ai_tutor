package repository

import (
	"context"
)

type VectorPoint struct {
	ID      uint64
	Vector  []float32
	Payload map[string]any
}

type QueryResult struct {
	Point VectorPoint
	Score float32
}

type IVectorDB interface {
	CreateCollection(ctx context.Context, collectionName string, vectorSize uint64) error
	DeleteCollection(ctx context.Context, collectionName string) error
	UpsertVectors(ctx context.Context, collectionName string, points []VectorPoint) error
	QueryVectors(ctx context.Context, collectionName string, vector []float32, topK int) ([]QueryResult, error)
	Close() error
}
