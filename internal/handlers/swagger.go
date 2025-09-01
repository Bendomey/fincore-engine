package handlers

import "net/http"

func SwaggerUI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
            <!DOCTYPE html>
            <html>
            <head>
              <title>API Docs</title>
              <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
            </head>
            <body>
              <redoc spec-url='/swagger.yaml'></redoc>
            </body>
            </html>
        `))
	}
}
