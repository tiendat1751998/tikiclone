package logging
import ("github.com/tikiclone/tiki/packages/go-shared/pkg/observability"; "go.uber.org/zap")
func Init(s, l string) *zap.Logger { return observability.InitLogger(s, l) }
