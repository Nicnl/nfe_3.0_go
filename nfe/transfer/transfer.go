package transfer

import (
	"github.com/gofrs/uuid"
	"time"
)

type State uint8

const (
	StateTransferring      State = 0
	StateFinished          State = 1
	StateInterruptedClient State = 2
	StateInterruptedServer State = 3
)

type Transfer struct {
	Guid uuid.UUID

	// Données changeantes
	CurrentSpeed           int64
	CurrentSpeedLimitDelay time.Duration // Vitesse de DL calculée à partir de CurrentSpeedLimit
	CurrentSpeedLimit      time.Duration // Vitesse de DL telle que
	ShouldInterrupt        bool
	CurrentState           uint8

	// Informations immutables

	ClientIP      string
	StartDate     time.Time
	Url           string
	UrlExpiration time.Time
	UrlSpeedLimit int64
	FileLength    int64
	SectionLength int64
	BufferSize    int64
}
