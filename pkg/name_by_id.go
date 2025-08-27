package gnames

import (
	"context"
	"fmt"

	vlib "github.com/gnames/gnlib/ent/verifier"
)

func (g gnames) NameByID(
	params vlib.NameStringInput,
	fullMatch bool,
) (vlib.NameStringOutput, error) {
	if fullMatch {
		return g.matchIDByName(params)
	} else {
		return g.matchID(params)
	}
}

func (g gnames) matchID(params vlib.NameStringInput) (vlib.NameStringOutput, error) {

	var res vlib.NameStringOutput
	mr, err := g.vf.NameByID(params)
	if err != nil {
		return res, fmt.Errorf("gnames.matchID: %w", err)
	}
	meta := vlib.NameStringMeta{
		ID:             params.ID,
		DataSources:    params.DataSources,
		WithAllMatches: params.WithAllMatches,
	}
	res.NameStringMeta = meta

	if mr == nil {
		return res, nil
	}

	name := outputName(mr, params.WithAllMatches)
	res.Name = &name
	return res, nil
}

func (g gnames) matchIDByName(params vlib.NameStringInput) (vlib.NameStringOutput, error) {
	var res vlib.NameStringOutput
	name, err := g.vf.NameStringByID(params.ID)
	if err != nil {
		return res, fmt.Errorf("gnames.matchIDByName: %w", err)
	}
	input := vlib.Input{
		NameStrings:      []string{name},
		DataSources:      params.DataSources,
		WithAllMatches:   true,
		WithSpeciesGroup: true,
	}
	var out vlib.Output
	out, err = g.Verify(context.Background(), input)
	if err != nil && len(out.Names) == 0 {
		return res, fmt.Errorf("gnames.matchIDByName: %w", err)
	}
	res = vlib.NameStringOutput{
		NameStringMeta: vlib.NameStringMeta{
			ID:             params.ID,
			DataSources:    params.DataSources,
			WithAllMatches: true,
		},
		Name: &out.Names[0],
	}
	return res, nil

}
