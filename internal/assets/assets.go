package assets

type AssetID uint64

type AssetType int

const (
	AssetTexture AssetType = iota
	AssetMesh
	AssetMaterial
)

type Asset struct {
	ID   AssetID
	Type AssetType
	Path string
	Data any
}
