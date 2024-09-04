package cognito

import (
	"fmt"
	"os"
	"strconv"
)

const (
	KeyMaxRPS           = "AMAZON_COGNITO_MAX_RPS"
	KeyMaxRPSChunkRatio = "AMAZON_COGNITO_MAX_RPS_CHUNK_RATIO"
)

func GetRowsPerChunk() (rowsPerChunk float32, err error) {
	maxRPS64, err := strconv.ParseFloat(os.Getenv(KeyMaxRPS), 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse env var of %s: %s", KeyMaxRPS, err.Error())
	}

	maxRPSChunkRatio64, err := strconv.ParseFloat(os.Getenv(KeyMaxRPSChunkRatio), 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse env var of %s: %s", KeyMaxRPSChunkRatio, err.Error())
	}

	return float32(maxRPS64) * float32(maxRPSChunkRatio64), nil
}
