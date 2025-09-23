package huma_utils

import "github.com/danielgtaylor/huma/v2"

func GetMetadataBool(m map[string]any, key string) *bool {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	b, ok := v.(bool)
	if !ok {
		return nil
	}
	return &b
}

func HasMetadataTrue2(m map[string]any, key string) bool {
	b := GetMetadataBool(m, key)
	if b == nil {
		return false
	}
	return *b
}

func HasMetadataTrue(ctx huma.Context, key string) bool {
	return HasMetadataTrue2(ctx.Operation().Metadata, key)
}

func HasMetadataFalse2(m map[string]any, key string) bool {
	b := GetMetadataBool(m, key)
	if b == nil {
		return false
	}
	return !*b
}

func HasMetadataFalse(ctx huma.Context, key string) bool {
	return HasMetadataFalse2(ctx.Operation().Metadata, key)
}

func MetadataModifier(key string, value any) func(o *huma.Operation) {
	return func(o *huma.Operation) {
		if o.Metadata != nil {
			_, ok := o.Metadata[key]
			if ok {
				return
			}
		} else {
			o.Metadata = map[string]any{}
		}

		o.Metadata[key] = value
	}
}
