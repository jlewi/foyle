package analyze

import (
	"github.com/gin-gonic/gin"
	"github.com/jlewi/foyle/app/pkg/config"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"net/http"
)

// CrudHandler is a handler for CRUD operations on log entries
type CrudHandler struct {
	blockLogs map[string]logspb.BlockLog
	cfg       config.Config
}

func NewCrudHandler(cfg config.Config) (*CrudHandler, error) {
	return &CrudHandler{
		cfg: cfg,
	}, nil
}

func (h *CrudHandler) GetBlockLog(c *gin.Context) {
	//log := zapr.NewLogger(zap.L())

	c.JSON(http.StatusNotImplemented, gin.H{"error": "GetBlockLog needs to be updated now that we are using pebble"})
	//// lazily load blocklogs
	//if h.blockLogs == nil {
	//	blockLogs, err := LoadLatestBlockLogs(c.Request.Context(), h.cfg.GetProcessedLogDir())
	//	if err != nil {
	//		log.Error(err, "Failed to load cache")
	//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//		return
	//	}
	//	h.blockLogs = blockLogs
	//
	//}
	//id := c.Param("id")
	//
	//b, ok := h.blockLogs[id]
	//if !ok {
	//	c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("No block with id %s", id)})
	//	return
	//}
	//// Use the id to fetch or manipulate the resource
	//// For now, we'll just echo it back
	//c.JSON(http.StatusOK, b)
}

//func LoadLatestBlockLogs(ctx context.Context, logDir string) (map[string]api.BlockLog, error) {
//	glob := filepath.Join(logDir, "blocks*.jsonl")
//	matches, err := filepath.Glob(glob)
//	if err != nil {
//		return nil, errors.Wrapf(err, "Failed to match glob %s", glob)
//	}
//
//	sort.Strings(matches)
//	blockLogs := make(map[string]api.BlockLog)
//
//	latest := matches[len(matches)-1]
//	log := logs.FromContext(ctx)
//	log.Info("Loading block log", "file", latest)
//
//	f, err := os.Open(latest)
//	if err != nil {
//		return nil, errors.Wrapf(err, "Failed to open %s", latest)
//	}
//	defer f.Close()
//	d := json.NewDecoder(f)
//	for {
//		b := &api.BlockLog{}
//		if err := d.Decode(b); err != nil {
//			if err == io.EOF {
//				break
//			}
//			log.Error(err, "Failed to decode block log")
//		}
//		blockLogs[b.ID] = *b
//	}
//
//	return blockLogs, nil
//}
