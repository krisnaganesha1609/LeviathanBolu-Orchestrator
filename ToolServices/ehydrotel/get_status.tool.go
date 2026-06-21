package ehydrotel

import "context"

type GetStatusTool struct{}

func (t *GetStatusTool) Name() string {
	return "ehydrotel_get_status"
}

func (t *GetStatusTool) Description() string {
	return "Get the current status of the eHydroTel hydroponic system: " +
		"water pH, nutrient concentration (PPM), and water temperature. " +
		"Takes no arguments."
}

// Schema describes this tool's INPUT. get_status takes no arguments, so
// this is an empty object schema — NOT a description of what Execute
// returns. (The previous version of this method described
// ph/ppm/water_temp here, as if the LLM had to supply readings as
// input — backwards.)
func (t *GetStatusTool) Schema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *GetStatusTool) Execute(c context.Context, args map[string]any) (any, error) {
	// TODO: replace with a real call to the eHydroTel device/API once
	// that integration exists. Hardcoded values are a placeholder so the
	// tool-calling pipeline (registry -> executor -> LLM) has something
	// real to exercise end-to-end before the actual hardware integration
	// is built.
	return map[string]any{
		"ph":         6.3,
		"ppm":        920,
		"water_temp": 24,
	}, nil
}
