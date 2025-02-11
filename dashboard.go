package gapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
)

type DashboardMeta struct {
	UID         string `json:"uid"`
	Title       string `json:"title"`
	IsStarred   bool   `json:"isStarred"`
	Slug        string `json:"slug"`
	Folder      int64  `json:"folderId"`
	FolderTitle string `json:"folderTitle"`
}

// DashboardSaveResponse grafana response for create dashboard
type DashboardSaveResponse struct {
	Slug    string `json:"slug"`
	ID      int64  `json:"id"`
	UID     string `json:"uid"`
	URL     string `json:"url"`
	Status  string `json:"status"`
	Version int64  `json:"version"`
}

type Dashboard struct {
	Meta      DashboardMeta  `json:"meta"`
	Model     DashboardModel `json:"dashboard"`
	Folder    int64          `json:"folderId"`
	Overwrite bool           `json:"overwrite"`
}

func (d Dashboard) FrontendURL(dashboardVars map[string][]string) string {
	var exportBase string
	exportBase += "/d"
	exportBase += "/" + d.Meta.UID
	exportBase += "?" + dasboardVarsToQueryString(dashboardVars)

	return exportBase
}

// GetPanelFromDashboard Returns the Panel from a dashboard by given PanelID
func (d Dashboard) GetPanelFromDashboard(panelID int64) (DashboardPanel, error) {

	for _, panel := range d.Model.Panels {
		if panel.ID == panelID {
			return panel, nil
		}
	}
	// panel not found
	var err error
	return DashboardPanel{}, err
}

// Dashboards represent json returned by search API
type Dashboards struct {
	ID          int64  `json:"id"`
	UID         string `json:"uid"`
	Title       string `json:"title"`
	URI         string `json:"uri"`
	URL         string `json:"url"`
	Starred     bool   `json:"isStarred"`
	FolderID    int64  `json:"folderId"`
	FolderUID   string `json:"folderUid"`
	FolderTitle string `json:"folderTitle"`
}

type Link struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type DashboardAnnotation struct {
	BuiltIn    int    `json:"builtIn"`
	Datasource string `json:"datasource"`
	Enable     bool   `json:"enable"`
	Hide       bool   `json:"hide"`
	IconColor  string `json:"iconColor"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}
type DashboardModel struct {
	Annotations struct {
		List []DashboardAnnotation `json:"list"`
	} `json:"annotations"`
	Editable      bool             `json:"editable"`
	GnetID        interface{}      `json:"gnetId"`
	GraphTooltip  int              `json:"graphTooltip"`
	ID            int              `json:"id"`
	Iteration     int64            `json:"iteration"`
	Links         []Link           `json:"links"`
	Panels        []DashboardPanel `json:"panels"`
	Refresh       bool             `json:"refresh"`
	SchemaVersion int              `json:"schemaVersion"`
	Style         string           `json:"style"`
	Tags          []interface{}    `json:"tags"`
	Templating    struct {
		List []struct {
			AllValue interface{} `json:"allValue"`
			Current  struct {
				Text  string `json:"text"`
				Value string `json:"value"`
			} `json:"current"`
			Hide       int         `json:"hide"`
			IncludeAll bool        `json:"includeAll"`
			Label      interface{} `json:"label"`
			Multi      bool        `json:"multi"`
			Name       string      `json:"name"`
			Options    []struct {
				Selected bool   `json:"selected"`
				Text     string `json:"text"`
				Value    string `json:"value"`
			} `json:"options"`
			Query       string `json:"query"`
			SkipURLSync bool   `json:"skipUrlSync"`
			Type        string `json:"type"`
		} `json:"list"`
	} `json:"templating"`
	Time       TimeRange `json:"time"`
	Timepicker struct {
		RefreshIntervals []string `json:"refresh_intervals"`
	} `json:"timepicker"`
	Timezone string `json:"timezone"`
	Title    string `json:"title"`
	UID      string `json:"uid"`
	Version  int    `json:"version"`
}

// DashboardDeleteResponse grafana response for delete dashboard
type DashboardDeleteResponse struct {
	Title string `json:title`
}

// Deprecated: use NewDashboard instead
func (c *Client) SaveDashboard(model map[string]interface{}, overwrite bool) (*DashboardSaveResponse, error) {
	wrapper := map[string]interface{}{
		"dashboard": model,
		"overwrite": overwrite,
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest("POST", "/api/dashboards/db", nil, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		data, _ = ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("status: %d, body: %s", resp.StatusCode, data)
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &DashboardSaveResponse{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *Client) NewDashboard(dashboard Dashboard) (*DashboardSaveResponse, error) {
	data, err := json.Marshal(dashboard)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest("POST", "/api/dashboards/import", nil, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &DashboardSaveResponse{}
	err = json.Unmarshal(data, &result)
	return result, err
}

// SearchDashboard search a dashboard in Grafana
func (c *Client) SearchDashboard(query string, folderID string) ([]Dashboards, error) {
	dashboards := make([]Dashboards, 0)
	path := "/api/search"

	params := url.Values{}
	params.Add("type", "dash-db")
	params.Add("query", query)
	params.Add("folderIds", folderID)

	req, err := c.newRequest("GET", path, params, nil)
	if err != nil {
		return dashboards, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return dashboards, err
	}
	if resp.StatusCode != 200 {
		return dashboards, errors.New(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dashboards, err
	}

	err = json.Unmarshal(data, &dashboards)

	return dashboards, err
}

// GetDashboard get a dashboard by UID
func (c *Client) GetDashboard(uid string) (*Dashboard, error) {
	path := fmt.Sprintf("/api/dashboards/uid/%s", uid)
	req, err := c.newRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Dashboard{}
	err = json.Unmarshal(data, &result)
	result.Folder = result.Meta.Folder
	if os.Getenv("GF_LOG") != "" {
		log.Printf("got back dashboard response  %s", data)
	}
	// the dashboard uid is not a part of the response
	result.Meta.UID = uid

	return result, err
}

// Deprecated: use GetDashboard instead
func (c *Client) Dashboard(slug string) (*Dashboard, error) {
	path := fmt.Sprintf("/api/dashboards/db/%s", slug)
	req, err := c.newRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Dashboard{}
	err = json.Unmarshal(data, &result)
	result.Folder = result.Meta.Folder
	if os.Getenv("GF_LOG") != "" {
		log.Printf("got back dashboard response  %s", data)
	}
	return result, err
}

// DeleteDashboard deletes a grafana dashoboard
func (c *Client) DeleteDashboard(uid string) (string, error) {
	deleted := &DashboardDeleteResponse{}
	path := fmt.Sprintf("/api/dashboards/uid/%s", uid)
	req, err := c.newRequest("DELETE", path, nil, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(data, &deleted)
	if err != nil {
		return "", err
	}
	return deleted.Title, nil
}
