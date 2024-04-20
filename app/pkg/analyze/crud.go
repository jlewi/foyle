package analyze

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

// CrudHandler is a handler for CRUD operations on log entries
type CrudHandler struct {
	blockLogs map[string]BlockLog
	cfg       config.Config
}

func NewCrudHandler(cfg config.Config) (*CrudHandler, error) {
	return &CrudHandler{
		cfg: cfg,
	}, nil
}

func (h *CrudHandler) GetBlockLog(c *gin.Context) {
	log := zapr.NewLogger(zap.L())
	if err := h.loadCache(c); err != nil {
		log.Error(err, "Failed to load cache")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id := c.Param("id")

	b, ok := h.blockLogs[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("No block with id %s", id)})
		return
	}
	// Use the id to fetch or manipulate the resource
	// For now, we'll just echo it back
	c.JSON(http.StatusOK, b)
}

// loadCache lazily loads the cache of block logs
func (h *CrudHandler) loadCache(ctx context.Context) error {
	if h.blockLogs != nil {
		return nil
	}

	pDir := h.cfg.GetProcessedLogDir()
	glob := filepath.Join(pDir, "blocks*.jsonl")
	matches, err := filepath.Glob(glob)
	if err != nil {
		errors.Wrapf(err, "Failed to match glob %s", glob)
		return err
	}

	sort.Strings(matches)
	h.blockLogs = make(map[string]BlockLog)

	latest := matches[len(matches)-1]
	log := logs.FromContext(ctx)
	log.Info("Loading block log", "file", latest)

	f, err := os.Open(latest)
	if err != nil {
		return errors.Wrapf(err, "Failed to open %s", latest)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	for {
		b := &BlockLog{}
		if err := d.Decode(b); err != nil {
			if err == io.EOF {
				break
			}
			log.Error(err, "Failed to decode block log")
		}

		h.blockLogs[b.ID] = *b
	}

	return nil
}
