package components

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

func ToastContainer() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<div id="toast-container" class="toast-container"></div>`)
		return err
	})
}
