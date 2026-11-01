package engine
import "fmt"
func GetEngine(engineName string) (SignatureEngine, error) {
	switch engineName {
	case "native", "secp256k1", "":
		return &ECDSAEngine{}, nil
	case "pqc-us", "pqc-eu":
		return &PQCGlobalEngine{}, nil
	default:
		return nil, fmt.Errorf("unknown engine: %s", engineName)
	}
}
