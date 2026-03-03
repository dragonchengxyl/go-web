package asset

import (
	"time"

	"github.com/google/uuid"
)

// AssetType represents the type of asset
type AssetType string

const (
	AssetTypeBaseGame AssetType = "base_game"
	AssetTypeDLC      AssetType = "dlc"
	AssetTypeOST      AssetType = "ost"
	AssetTypeArtbook  AssetType = "artbook"
)

// AssetSource represents how the asset was obtained
type AssetSource string

const (
	AssetSourcePurchase  AssetSource = "purchase"
	AssetSourceGift      AssetSource = "gift"
	AssetSourceFreeClaim AssetSource = "free_claim"
	AssetSourceRedeemCode AssetSource = "redeem_code"
)

// UserGameAsset represents a user's game asset
type UserGameAsset struct {
	ID         uuid.UUID   `json:"id"`
	UserID     uuid.UUID   `json:"user_id"`
	GameID     uuid.UUID   `json:"game_id"`
	AssetType  AssetType   `json:"asset_type"`
	AssetID    uuid.UUID   `json:"asset_id"`
	ObtainedAt time.Time   `json:"obtained_at"`
	Source     AssetSource `json:"source"`
	CreatedAt  time.Time   `json:"created_at"`
}

// DownloadLog represents a download log entry
type DownloadLog struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	ReleaseID    uuid.UUID `json:"release_id"`
	ClientIP     string    `json:"client_ip"`
	UserAgent    *string   `json:"user_agent,omitempty"`
	DownloadedAt time.Time `json:"downloaded_at"`
	CreatedAt    time.Time `json:"created_at"`
}
