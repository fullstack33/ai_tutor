package qdrant

import (
	"context"
	"fmt"

	"github.com/fullstack33/ai-tutor/internal/repository"
	"github.com/qdrant/go-client/qdrant"
)

type QdrantRepository struct {
	client *qdrant.Client
}

func NewQdrantRepository(host string, port int) (repository.IVectorDB, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant: %w", err)
	}
	return &QdrantRepository{
		client: client,
	}, nil
}

func (q *QdrantRepository) CreateCollection(ctx context.Context, collectionName string, vectorSize uint64) error {
	exists, err := q.client.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists: %w", err)
	}
	if exists {
		return nil
	}

	err = q.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	return nil
}

func (q *QdrantRepository) DeleteCollection(ctx context.Context, collectionName string) error {
	exists, err := q.client.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists for deletion: %w", err)
	}
	if !exists {
		return nil
	}
	err = q.client.DeleteCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	return nil
}

func (q *QdrantRepository) UpsertVectors(ctx context.Context, collectionName string, points []repository.VectorPoint) error {
	qPoints := make([]*qdrant.PointStruct, len(points))
	for i, p := range points {
		qPoints[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(p.ID),
			Vectors: qdrant.NewVectors(p.Vector...),
			Payload: qdrant.NewValueMap(p.Payload),
		}
	}

	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         qPoints,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert points: %w", err)
	}
	return nil
}

func (q *QdrantRepository) QueryVectors(ctx context.Context, collectionName string, vector []float32, topK int) ([]repository.QueryResult, error) {
	limit := uint64(topK)
	res, err := q.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(vector...),
		Limit:          &limit,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query Qdrant points: %w", err)
	}

	results := make([]repository.QueryResult, len(res))
	for i, p := range res {
		payloadMap := make(map[string]any)
		for key, val := range p.Payload {
			payloadMap[key] = valueToAny(val)
		}

		var idVal uint64
		if p.Id != nil {
			idVal = p.Id.GetNum()
		}

		results[i] = repository.QueryResult{
			Point: repository.VectorPoint{
				ID:      idVal,
				Payload: payloadMap,
			},
			Score: p.Score,
		}
	}

	return results, nil
}

func (q *QdrantRepository) Close() error {
	if q.client != nil {
		return q.client.Close()
	}
	return nil
}

func valueToAny(v *qdrant.Value) any {
	if v == nil {
		return nil
	}
	switch k := v.Kind.(type) {
	case *qdrant.Value_NullValue:
		return nil
	case *qdrant.Value_DoubleValue:
		return k.DoubleValue
	case *qdrant.Value_IntegerValue:
		return k.IntegerValue
	case *qdrant.Value_StringValue:
		return k.StringValue
	case *qdrant.Value_BoolValue:
		return k.BoolValue
	case *qdrant.Value_StructValue:
		m := make(map[string]any)
		if k.StructValue != nil {
			for key, val := range k.StructValue.Fields {
				m[key] = valueToAny(val)
			}
		}
		return m
	case *qdrant.Value_ListValue:
		var list []any
		if k.ListValue != nil {
			for _, val := range k.ListValue.Values {
				list = append(list, valueToAny(val))
			}
		}
		return list
	}
	return nil
}
