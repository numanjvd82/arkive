package ads

import (
	"os"
	"strings"
)

func AdblockModalDisabled() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("DISABLE_ADBLOCK_MODAL")))
	return value == "true"
}
