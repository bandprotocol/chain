package cli

import (
	"encoding/json"
	"os"

	"github.com/bandprotocol/chain/v3/x/tss/types"
)

type Complaints struct {
	Complaints []types.Complaint `json:"complaints"`
}

// parseComplaints reads and parses a JSON file containing complaints into a slice of complain objects.
func parseComplaints(complaintsFile string) ([]types.Complaint, error) {
	var complaints Complaints

	if complaintsFile == "" {
		return complaints.Complaints, nil
	}

	bz, err := os.ReadFile(complaintsFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bz, &complaints)
	if err != nil {
		return nil, err
	}

	return complaints.Complaints, nil
}
