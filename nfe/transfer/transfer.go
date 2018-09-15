package transfer

import (
	"github.com/gofrs/uuid"
	"nfe_3.0_go/nfe/json_time"
	"nfe_3.0_go/nfe/vfs"
	"time"
)

type State uint8

const (
	StateTransferring      State = 0
	StateFinished          State = 1
	StateInterruptedClient State = 2
	StateInterruptedServer State = 3
	StateInterruptedAdmin  State = 4
)

type Transfer struct {
	Guid uuid.UUID

	// Données changeantes
	CurrentSpeed           int64         `json:"current_speed"`
	CurrentSpeedLimitDelay time.Duration // Vitesse de DL calculée à partir de CurrentSpeedLimit
	CurrentSpeedLimit      int64         `json:"current_speed_limit"` // Limite de vitesse de DL brute en octets/s
	ShouldInterrupt        bool
	CurrentState           uint8 `json:"current_state"`

	// Informations immutables
	Downloaded    int64 `json:"downloaded"`
	ClientIP      string
	StartDate     json_time.JsonTime `json:"start_date"`
	EndDate       json_time.JsonTime `json:"end_date"`
	Url           string
	UrlExpiration time.Time
	UrlSpeedLimit int64
	FileLength    int64  `json:"file_length"`
	FileName      string `json:"file_name"`
	FilePath      string
	SectionStart  int64 `json:"section_start"`
	SectionLength int64 `json:"section_length"`
	BufferSize    int64
}

func (t *Transfer) SetSpeedLimit(speedLimit int64) {
	if speedLimit == 0 {
		t.CurrentSpeedLimitDelay = 0
		return
	}

	nbPackets := speedLimit / t.BufferSize
	t.CurrentSpeedLimitDelay = time.Second / time.Duration(nbPackets)
}

func New(vfs vfs.Vfs, vfsPath string, speedLimit int64, bufferSize int64) (*Transfer, error) {
	// Obtention de la taille du fichier (et on vérifie que le fichier existe vraiment par la même occasion)
	info, err := vfs.Stat(vfsPath)
	if err != nil {
		return nil, err
	}

	// Création de l'instance du transfert
	t := Transfer{
		Guid: uuid.Must(uuid.NewV4()),

		FileName:      info.Name(),
		FilePath:      vfsPath,
		FileLength:    info.Size(),
		SectionLength: 0, // Dépends de la requête du client : calculé par la fonction ServeFile
		BufferSize:    bufferSize,
	}

	// Application de la limite de vitesse
	t.SetSpeedLimit(speedLimit)

	return &t, nil
}
