package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const tmpl = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<link rel="stylesheet" type="text/css" href="/static/style.css">
	<script src="/static/wasm_exec.js"></script>
	<script>
			if (!WebAssembly.instantiateStreaming) { // polyfill
				WebAssembly.instantiateStreaming = async (resp, importObject) => {
					const source = await (await resp).arrayBuffer();
					return await WebAssembly.instantiate(source, importObject);
				};
			}

			const go = new Go();
			WebAssembly.instantiateStreaming(fetch("/static/snake.wasm"), go.importObject).then((result) => {
				go.run(result.instance);
			});
		</script>
    <title>Snake</title>
</head>
<body>
<div id="canvas-container">
	<canvas id="canvas"></canvas>
</div>
</body>
</html>
`

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("\nload config: %w\n", err)
		os.Exit(2)
	}

	fs := http.FileServer(http.Dir("static"))

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(tmpl)); err != nil {
			fmt.Printf("couldn't write data: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})

	api := http.Server{
		Addr:         cfg.HTTP.Addr,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		Handler:      mux,
	}
	apiErrors := make(chan error, 1)

	go func() {
		fmt.Printf("API started on %s\n", api.Addr)
		apiErrors <- api.ListenAndServe()
	}()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-apiErrors:
		fmt.Printf("unexpected API error: %v\n", err)
	case <-osSignals:
		ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			fmt.Printf("API gracefule shutdown failed: %v\n", err)
			_ = api.Close()
		}
	}
}
