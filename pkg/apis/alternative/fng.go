package alternative

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type AlternativeMetadata struct {
	Error string `json:"error"`
}

type FearAndGreedItem struct {
	Value               string `json:"value"`
	ValueClassification string `json:"value_classification"`
	Timestamp           string `json:"timestamp"`
	TimeUntilUpdate     string `json:"time_until_update"`
}

type GetFearAndGreedIndexResp struct {
	Name     string               `json:"name"`
	Metadata *AlternativeMetadata `json:"metadata"`
	Data     []*FearAndGreedItem  `json:"data"`
}

// Fear and Greed Index API
// https://api.alternative.me/fng/?limit=100&format=json&date_format=cn
func (an *AlternativeClient) GetFearAndGreedIndex(limit int) (*GetFearAndGreedIndexResp, error) {
	url := fmt.Sprintf("%s/fng/?limit=%d&format=json", an.baseURL, limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	log.WithField("url", url).Debug("get fng")

	resp, err := an.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Errorf("response error, status code: %d, detail: %s", resp.StatusCode, body)
	}

	fngResp := &GetFearAndGreedIndexResp{}
	if err := json.NewDecoder(resp.Body).Decode(&fngResp); err != nil {
		return nil, err
	}

	if fngResp != nil && fngResp.Metadata != nil && fngResp.Metadata.Error != "" {
		return nil, errors.Errorf("fng resp metadata error: %s", fngResp.Metadata.Error)
	}

	return fngResp, nil
}
