// Code generated from Pkl module `henrikvtcodes.tungsten.config.Records`. DO NOT EDIT.
package records

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Records struct {
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Records
func LoadFromPath(ctx context.Context, path string) (ret *Records, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Records
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Records, error) {
	var ret Records
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
