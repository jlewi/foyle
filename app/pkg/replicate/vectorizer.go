package replicate

import (
	"context"
	"encoding/json"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/replicate/replicate-go"
	"go.uber.org/zap"
	"gonum.org/v1/gonum/mat"
)

const (
	vectorLength = 1024
)

func NewVectorizer(client *replicate.Client) (*Vectorizer, error) {
	return &Vectorizer{client: client}, nil
}

// Vectorizer computes embedding representations of text using models on replicate
type Vectorizer struct {
	client *replicate.Client
}

func (v *Vectorizer) Embed(ctx context.Context, text string) (*mat.VecDense, error) {
	log := zapr.NewLogger(zap.L())

	model := "replicate/retriever-embeddings"
	version := "9cf9f015a9cb9c61d1a2610659cdac4a4ca222f2d3707a68517b18c198a9add1"
	modelVersion := model + ":" + version
	texts, err := json.Marshal([]string{text})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to marshal text")
	}
	input := replicate.PredictionInput{
		"texts":                string(texts),
		"convert_to_numpy":     false,
		"normalize_embeddings": true,
	}

	id, err := replicate.ParseIdentifier(modelVersion)
	if err != nil {
		return nil, err
	}

	if id.Version == nil {
		return nil, errors.New("version must be specified")
	}

	prediction, err := v.client.CreatePrediction(ctx, *id.Version, input, nil, false)
	if err != nil {
		return nil, err
	}

	err = v.client.Wait(ctx, prediction)

	if prediction.Status != "succeeded" {
		log.Error(errors.New("Prediction failed"), "Prediction failed", "prediction", prediction)
		return nil, errors.Errorf("Prediction failed; %+v", prediction.Error)
	}

	// The return type is [][]interface{}
	arrays, ok := prediction.Output.([]interface{})

	if !ok {
		return nil, errors.New("Failed to convert output to array")
	}

	if len(arrays) != 1 {
		return nil, errors.Errorf("Expected exactly 1 array but got %d", len(arrays))
	}

	array, ok := arrays[0].([]interface{})
	if !ok {
		return nil, errors.New("Failed to convert output to float64")
	}
	vec := make([]float64, 0, len(array))
	for _, v := range array {
		f, ok := v.(float64)
		if !ok {
			return nil, errors.New("Failed to convert output to float64")
		}
		vec = append(vec, f)

	}
	return mat.NewVecDense(vectorLength, vec), nil
}

func (v *Vectorizer) Length() int {
	return vectorLength
}
