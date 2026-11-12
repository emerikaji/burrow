package network

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/emerikaji/burrow/pkt"
)

// ─── GET References From A Git Repository ────────────────────────────────────

func DiscoverRefs(url string) (refs map[string]string, capabilities map[string]string, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/info/refs?service=git-upload-pack", url), nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "git/burrow-v0.1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status %v", resp.Status)
		return
	}

	packets, err := pkt.Parse(resp.Body)
	if err != nil {
		return
	}

	if len(packets) < 3 {
		err = errors.New("missing packets")
		return
	}

	if string(packets[0].Data()) != "# service=git-upload-pack\n" {
		err = errors.New("header mismatch")
		return
	}

	if _, ok := packets[1].(*pkt.Flush); !ok {
		err = errors.New("missing flush after header")
		return
	}

	refs = make(map[string]string)

	for _, packet := range packets[2:] {
		if packet.Len() == 0 {
			break
		}

		line := string(packet.Data())
		line, found := strings.CutSuffix(line, "\n")
		if !found {
			err = errors.New("unterminated line")
			return
		}

		if idx := strings.IndexByte(line, '\x00'); capabilities == nil && idx != -1 {
			capabilities = make(map[string]string)

			for capability := range strings.SplitSeq(line[idx+1:], " ") {
				capabilityName, capabilityVal, found := strings.Cut(capability, "=")
				if !found {
					capabilityVal = "true"
				}
				capabilities[capabilityName] = capabilityVal
			}

			line = line[:idx]
		}

		hash, name, found := strings.Cut(line, " ")
		if !found {
			continue
		}
		refs[name] = hash
	}

	return
}

// ─────────────────────────────────────────────────────────────────────────────
