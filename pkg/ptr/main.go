package ptr

func FromPtr[T any](v *T) T { return *v }

func ToPtr[T any](v T) *T { return &v }
