package app
mport (
	"context"
	// "fmt"
	"time"
	"net/http"
	"log"

	"golang/internal/repository/_postgres" // Replace with your actual internal path
	"golang/pkg/modules"
	"golang/internal/repository" // Replace with your actual internal path
	"golang/internal/usecase"
	"golang/internal/handler"
	"golang/internal/middleware"

)
func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := initPostgreConfig()
	
	_postgre := _postgres.NewPGXDialect(ctx, dbConfig)

	fmt.Println(_postgre)
}

func initPostgreConfig() *modules.PostgreConfig{
	return &modules.PostgreConfig{
		Host: "localhost",
		Port: "5432".
		Username: "appuser",
		Password: "appuser",
		DBName: "golangdb",
		SSLMode: "disable",
		ExecTimeout: 5 * time.Second,
	}
}