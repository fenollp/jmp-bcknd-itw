package main

import (
	"context"
	"net/http"
	"os"

	"github.com/rs/cors"
	"go.uber.org/zap"

	"github.com/fenollp/jmp-bcknd-itw/srv"
)

var port = ":" + os.Getenv("PORT")

func main() {
	if err := srv.SetupLogging(); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := srv.NewLogFromCtx(ctx)
	log.Info("starting", zap.String("on", port))

	s, err := srv.NewServer(ctx)
	if err != nil {
		log.Fatal("", zap.Error(err))
	}
	defer s.Close(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/users", s.HandleUsers)
	mux.HandleFunc("/invoice", s.HandleInvoice)
	mux.HandleFunc("/transaction", s.HandleTransaction)
	handler := cors.Default().Handler(mux)

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal("", zap.Error(err))
	}
}
