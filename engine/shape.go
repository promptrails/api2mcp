package engine

import "fmt"

// Truncate wraps a ShapeFunc so its rendered output never exceeds max bytes.
// When the limit is hit the text is cut and a clear notice is appended, so the
// LLM knows the payload was clipped rather than silently losing data. A max of
// 0 or less disables truncation.
func Truncate(max int, inner ShapeFunc) ShapeFunc {
	if inner == nil {
		inner = DefaultShape
	}
	return func(r *Response) string {
		out := inner(r)
		if max <= 0 || len(out) <= max {
			return out
		}
		return out[:max] + fmt.Sprintf("\n\n…[truncated: %d of %d bytes shown]", max, len(out))
	}
}
