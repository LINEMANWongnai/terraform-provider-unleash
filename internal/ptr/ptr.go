package ptr

func ToPtr[T any](v T) *T {
	return &v
}

func ToValue[T any](v *T, defaultValue func() T) T {
	if v == nil {
		return defaultValue()
	}
	return *v
}
