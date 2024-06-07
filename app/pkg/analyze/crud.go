package analyze

import (
	"fmt"
	"net/http"

	"github.com/cockroachdb/pebble"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// CrudHandler is a handler for CRUD operations on log entries
type CrudHandler struct {
	cfg      config.Config
	blocksDB *pebble.DB
}

func NewCrudHandler(cfg config.Config, blocksDB *pebble.DB) (*CrudHandler, error) {
	return &CrudHandler{
		cfg:      cfg,
		blocksDB: blocksDB,
	}, nil
}

func (h *CrudHandler) GetBlockLog(c *gin.Context) {
	log := zapr.NewLogger(zap.L())

	id := c.Param("id")

	log = log.WithValues("id", id)
	bLog := &logspb.BlockLog{}
	if err := dbutil.GetProto(h.blocksDB, id, bLog); err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("No block with id %s", id)})
			return
		} else {
			log.Error(err, "Failed to read block with id", "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read block with id %s; error %+v", id, err)})
			return
		}

	}

	b, err := protojson.Marshal(bLog)
	if err != nil {
		log.Error(err, "Failed to marshal block log", "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to marshal block with id %s; error %+v", id, err)})
		return

	}

	// Use the id to fetch or manipulate the resource
	// For now, we'll just echo it back
	if _, err := c.Writer.Write(b); err != nil {
		log.Error(err, "Failed to write response", "id", id)
	}
}
