package postgres

import (
	"encoding/json"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
)

func marshalStringSlice(values []string) ([]byte, error) {
	if values == nil {
		values = []string{}
	}

	return json.Marshal(values)
}

func unmarshalStringSlice(data []byte) ([]string, error) {
	if len(data) == 0 {
		return []string{}, nil
	}

	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}

	return values, nil
}

func marshalSpecifications(values []domain.ProductSpecification) ([]byte, error) {
	if values == nil {
		values = []domain.ProductSpecification{}
	}

	return json.Marshal(values)
}

func unmarshalSpecifications(data []byte) ([]domain.ProductSpecification, error) {
	if len(data) == 0 {
		return []domain.ProductSpecification{}, nil
	}

	var values []domain.ProductSpecification
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}

	return values, nil
}

func marshalReviews(values []domain.ProductReview) ([]byte, error) {
	if values == nil {
		values = []domain.ProductReview{}
	}

	return json.Marshal(values)
}

func unmarshalReviews(data []byte) ([]domain.ProductReview, error) {
	if len(data) == 0 {
		return []domain.ProductReview{}, nil
	}

	var values []domain.ProductReview
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}

	return values, nil
}
